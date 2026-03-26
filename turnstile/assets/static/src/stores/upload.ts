import { useEventListener, useWakeLock } from "@vueuse/core";
import axios, { type AxiosProgressEvent, isAxiosError } from "axios";
import { defineStore } from "pinia";
import { reactive, ref } from "vue";
import { toast } from "vue-sonner";
import { ApiPath } from "@/config/api.ts";
import { useConfigStore } from "@/stores/config.ts";
import { deobfuscate, obfuscate, obfuscationKey, xorTransform } from "@/util/obfuscate.ts";

let uploadID = 0;

export type InProgressItem = {
  original_name: string;
  progress: AxiosProgressEvent;
  controller: AbortController;
};

export type UploadedItem = {
  url: string;
  filename: string;
  original_name?: string;
  delete_key: string;
  access_key?: string;
  expiry: number;
  uploaded?: Date;
  size: number;
  mimetype?: string;
};

const MaxDelay = 2 ** 31 - 1;

export const useUploadStore = defineStore(
  "uploads",
  () => {
    const config = useConfigStore();

    const version = ref(0);
    const uploads = ref<UploadedItem[]>([]);
    const inProgress = reactive<Record<number, InProgressItem>>({});
    let timeout: ReturnType<typeof setTimeout> | undefined;

    const copy = async (item: UploadedItem) => {
      try {
        const url = document.location.origin + "/" + item.filename;
        await navigator.clipboard.writeText(url);
        toast.success("Copied to clipboard.", {
          description: url,
        });
      } catch (err) {
        console.error(err);
        toast.error("Failed to copy.", {
          description: err instanceof Error ? err.message : String(err),
        });
        throw err;
      }
    };

    const wakelock = reactive(useWakeLock());

    const ensureAuth = async () => {
      if (!config.site.auth) return;
      await axios.post(ApiPath("/api/auth"), null, {
        headers: {
          "Linx-Api-Key": encodeURIComponent(config.apiKey),
        },
        validateStatus: (s) => s === 200,
      });
    };

    const normalizeUploadResponse = (
      data: Record<string, any>,
      file: File,
      saveOriginalName: boolean,
    ) => {
      if (saveOriginalName) {
        data.original_name = file.name;
      }
      data.uploaded = new Date();
      data.expiry = Number(data.expiry ?? 0);
      data.size = Number(data.size ?? 0);
      return data as UploadedItem;
    };

    const onUploadSuccess = (item: UploadedItem) => {
      toast.success("File uploaded", {
        description: item.original_name || item.filename,
        action: {
          label: "Copy",
          onClick: async () => await copy(item),
        },
      });
      uploads.value = [item, ...uploads.value.filter((u) => u.filename !== item.filename)];
      removeExpired();
      return item;
    };

    const uploadErrorDescription = (err: unknown) => {
      let description = err instanceof Error ? err.message : String(err);
      if (isAxiosError(err) && err.response?.data?.error) description = err.response.data.error;
      return description;
    };

    const runUploadRequest = async ({
      file,
      saveOriginalName,
      request,
      canceledByUserText,
    }: {
      file: File;
      saveOriginalName: boolean;
      request: () => Promise<{ data: Record<string, any> }>;
      canceledByUserText?: string;
    }) => {
      try {
        await ensureAuth();
        const res = await request();
        const item = normalizeUploadResponse(res.data, file, saveOriginalName);
        return onUploadSuccess(item);
      } catch (err) {
        let description = uploadErrorDescription(err);
        if (canceledByUserText && description === "canceled") {
          description = canceledByUserText;
        }
        toast.error("Upload failed", { description });
        throw err;
      }
    };

    useEventListener(window, "beforeunload", (e) => {
      if (Object.keys(inProgress).length !== 0) {
        e.preventDefault();
      }
    });

    const uploadFileWS = async ({
      file,
      expiry,
      randomFilename = false,
      password,
      saveOriginalName = true,
      deleteKey,
      filename,
    }: {
      file: File;
      expiry: number | string;
      randomFilename?: boolean;
      password?: string;
      saveOriginalName?: boolean;
      deleteKey?: string;
      filename?: string;
    }) => {
      const controller = new AbortController();
      const upload: InProgressItem = {
        original_name: file.name,
        progress: { progress: 0, loaded: 0, total: file.size } as AxiosProgressEvent,
        controller,
      };
      const id = uploadID++;
      if (Object.keys(inProgress).length === 0) {
        wakelock.request("screen");
      }
      inProgress[id] = upload;

      return new Promise<UploadedItem>((resolve, reject) => {
        const url = new URL(ApiPath("/api/ws"));
        url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
        const ws = new WebSocket(url.toString());

        controller.signal.addEventListener("abort", () => {
          ws.close();
          reject(new Error("canceled"));
        });

        ws.onopen = () => {
          // 1. Send obfuscated metadata as binary
          const metadata = JSON.stringify({
            filename: filename || file.name,
            size: file.size,
            expiry: String(expiry),
            password: password,
            random: randomFilename,
            delete_key: deleteKey,
          });
          
          // XOR transform then Base64 encode
          const encoded = obfuscate(metadata);
          const encoder = new TextEncoder();
          ws.send(encoder.encode(encoded));

          // 2. Stream file content as binary chunks
          const reader = file.stream().getReader();
          let loaded = 0;
          let xorOffset = 0;
          const push = async () => {
            while (true) {
              const { done, value } = await reader.read();
              if (done) break;
              
              // Obfuscate the file chunk with rolling offset
              const transformedValue = xorTransform(value, xorOffset);
              ws.send(transformedValue);
              xorOffset = (xorOffset + value.length) % obfuscationKey.length;
              
              loaded += value.length;
              if (inProgress[id]) {
                inProgress[id].progress = {
                  progress: loaded / file.size,
                  loaded,
                  total: file.size,
                } as AxiosProgressEvent;
              }
            }
          };
          push();
        };

        ws.onmessage = async (event) => {
          try {
            // Server response is now a Base64-encoded binary string (which was XORed)
            let rawData = "";
            if (event.data instanceof Blob) {
              rawData = await event.data.text();
            } else if (typeof event.data === "string") {
              rawData = event.data;
            } else {
              const decoder = new TextDecoder();
              rawData = decoder.decode(event.data);
            }

            const decodedJson = deobfuscate(rawData);
            const data = JSON.parse(decodedJson);

            if (data.type === "error" || data.error) {
              const description = data.error || "Unknown error";
              toast.error("Upload failed", { description });
              reject(new Error(description));
            } else {
              const item = normalizeUploadResponse(data, file, saveOriginalName);
              resolve(onUploadSuccess(item));
            }
          } catch (err) {
            reject(err);
          } finally {
            ws.close();
          }
        };

        ws.onerror = (err) => {
          toast.error("WebSocket error");
          reject(err);
        };

        ws.onclose = () => {
          delete inProgress[id];
          if (Object.keys(inProgress).length === 0) {
            wakelock.release();
          }
        };
      });
    };

    const uploadFile = async ({
      file,
      expiry,
      randomFilename = false,
      password,
      saveOriginalName = true,
    }: {
      file: File;
      expiry: number | string;
      randomFilename?: boolean;
      password?: string;
      saveOriginalName?: boolean;
    }) => {
      const controller = new AbortController();
      const upload: InProgressItem = {
        original_name: file.name,
        progress: { progress: 0 } as AxiosProgressEvent,
        controller,
      };
      const id = uploadID++;
      if (Object.keys(inProgress).length === 0) {
        wakelock.request("screen");
      }
      inProgress[id] = upload;

      const form = new FormData();
      form.append("size", String(file.size));
      form.append("expires", String(expiry));
      if (password) form.append("access_key", password);
      form.append("randomize", randomFilename.toString());
      // This field must be last since it is streamed
      form.append("file", file);

      try {
        return await runUploadRequest({
          file,
          saveOriginalName,
          canceledByUserText: "Canceled by user",
          request: async () =>
            await axios.post(ApiPath(`/`), form, {
              headers: {
                Accept: "application/json",
                "Linx-Api-Key": encodeURIComponent(config.apiKey),
              },
              signal: controller.signal,
              validateStatus: (s) => s === 200,
              onUploadProgress(state) {
                if (inProgress[id]) inProgress[id].progress = state;
              },
            }),
        });
      } finally {
        delete inProgress[id];
        if (Object.keys(inProgress).length === 0) {
          wakelock.release();
        }
      }
    };

    const overwriteFile = async ({
      file,
      filename,
      deleteKey,
      expiry,
      password,
      saveOriginalName = true,
    }: {
      file: File;
      filename: string;
      deleteKey: string;
      expiry: number | string;
      password?: string;
      saveOriginalName?: boolean;
    }) => {
      return await uploadFileWS({
        file,
        expiry,
        password,
        saveOriginalName,
        deleteKey,
        filename,
      });
    };

    const deleteItem = async (upload: UploadedItem) => {
      try {
        const metadata = upload.delete_key ?? "";
        const encoded = obfuscate(metadata);

        await axios.get(ApiPath(`/api/delete/${upload.filename}?d=${encodeURIComponent(encoded)}`), {
          validateStatus: (s) => s === 200 || s === 404,
          headers: {
            Accept: "application/json",
            "Linx-Api-Key": encodeURIComponent(config.apiKey),
          },
        });
        uploads.value = uploads.value.filter((u) => u.filename !== upload.filename);
        toast.success("File deleted", { description: upload.original_name || upload.filename });
      } catch (err) {
        let description = err instanceof Error ? err.message : String(err);
        if (isAxiosError(err) && err.response?.data?.error) description = err.response.data.error;
        toast.error("Delete failed", { description });
        throw err;
      }
    };

    const removeExpired = () => {
      clearTimeout(timeout);
      const now = Math.floor(Date.now() / 1000);
      let closest = Infinity;
      uploads.value = uploads.value.filter((u) => {
        if (u.expiry === 0) return true;
        if (u.expiry <= now) return false;
        if (u.expiry < closest) closest = u.expiry;
        return true;
      });
      if (!Number.isFinite(closest)) return;
      const nextRun = Math.min(MaxDelay, Math.max(500, (closest - now) * 1000));
      timeout = setTimeout(removeExpired, nextRun);
    };

    return {
      version,
      uploads,
      inProgress,
      uploadFile,
      uploadFileWS,
      overwriteFile,
      deleteItem,
      removeExpired,
      copy,
    };
  },
  {
    persist: {
      pick: ["version", "uploads"],
      afterHydrate(ctx) {
        if (ctx.store.$state.version === 0 && ctx.store.$state.uploads?.length) {
          ctx.store.$state.uploads = ctx.store.$state.uploads.map((u: any) => ({
            ...u,
            expiry: Number(u.expiry),
            size: Number(u.size),
          }));
          ctx.store.version = 1;
        }

        ctx.store.removeExpired();
      },
    },
  },
);
