import { defineStore } from "pinia";
import { ref } from "vue";
import { deobfuscate } from "@/util/obfuscate.ts";

export interface ExpirationTime {
  name: string;
  value: string;
}

export interface WindowConfig {
  site_name: string;
  site_path: string;
  max_size: number;
  force_random: boolean;
  auth: boolean;
  expiration_times: ExpirationTime[];
  custom_pages?: string[];
}

declare global {
  interface Window {
    config: WindowConfig;
    obfuscatedConfig: string;
  }
}

export const useConfigStore = defineStore(
  "config",
  () => {
    let initialConfig: WindowConfig;
    try {
      initialConfig = JSON.parse(deobfuscate(window.obfuscatedConfig));
    } catch (err) {
      console.error("Failed to deobfuscate config", err);
      // Fallback to empty or window.config if exists (for dev)
      initialConfig = window.config || {
        site_name: "Linx",
        site_path: "/",
        max_size: 0,
        force_random: false,
        auth: false,
        expiration_times: [],
      };
    }

    const site = ref(initialConfig);
    const apiKey = ref("");
    const defaultExpiry =
      site.value?.expiration_times?.[site.value.expiration_times.length - 1]?.value ?? "";
    const expiry = ref(defaultExpiry);
    const filename = ref("");
    const extension = ref("txt");
    const randomFilename = ref(true);
    const password = ref("");
    const overwrite = ref(false);
    const editTargetFilename = ref("");
    const editDeleteKey = ref("");
    const content = ref("");

    return {
      site,
      apiKey,
      expiry,
      filename,
      extension,
      randomFilename,
      password,
      overwrite,
      editTargetFilename,
      editDeleteKey,
      content,
    };
  },
  {
    persist: {
      pick: ["apiKey", "expiry"],
      afterHydrate(ctx) {
        const options = (ctx.store.site?.expiration_times ?? []) as ExpirationTime[];
        const fallback = options[options.length - 1]?.value ?? "";
        if (!options.some((opt) => opt.value === ctx.store.expiry)) {
          ctx.store.expiry = fallback;
        }
      },
    },
  },
);
