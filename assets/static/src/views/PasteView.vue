<template>
  <form @submit.prevent="doUpload" class="space-y-4">
    <div class="flex flex-row items-center gap-2">

      <Input
        v-model="config.filename"
        placeholder="my_file.txt"
        class="w-full sm:flex-1"
        aria-label="Filename"
        :disabled="config.overwrite && canOverwriteExisting"
      />


      <Tooltip v-if="canOverwriteExisting">
        <TooltipTrigger as-child>
          <Toggle
            variant="outline"
            :model-value="config.overwrite"
            @update:model-value="(v) => (config.overwrite = !!v)"
            :data-state="config.overwrite ? 'on' : 'off'"
          >
            <PublishedChangesIcon class="text-2xl" />
            <span class="sr-only">Overwrite existing link</span>
          </Toggle>
        </TooltipTrigger>
        <TooltipContent side="bottom">
          <div class="text-sm">Overwrite existing link</div>
          <div class="text-xs text-muted-foreground">
            Available because you uploaded this file.
          </div>
        </TooltipContent>
      </Tooltip>

      <PasswordInput v-model="config.password" class="w-full sm:flex-1" />
      <ExpirySelect
        v-model="config.expiry"
        :options="config.site?.expiration_times"
        class="w-full sm:w-24"
      />
      <Button type="submit" size="icon" class="shrink-0">
        <ContentPasteIcon class="text-xl" />
        <span class="sr-only">Paste</span>
      </Button>
    </div>

    <Alert v-if="config.editTargetFilename && !canOverwriteExisting">
      <InfoIcon />
      <AlertTitle>
        This file is not in your upload history, so editing will create a new link.
      </AlertTitle>
    </Alert>

    <Textarea
      ref="textarea"
      v-model="config.content"
      placeholder="Paste your text here..."
      class="font-mono h-32"
      autofocus
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
    />
  </form>

  <AuthDialog v-if="config.site?.auth" v-model="showAuth" @submit="doUpload" />
</template>

<script setup lang="ts">
import { useDropZone, useEventListener, useMagicKeys } from "@vueuse/core";
import { isAxiosError } from "axios";
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { Alert, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Toggle } from "@/components/ui/toggle";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import AuthDialog from "@/components/upload/AuthDialog.vue";
import ExpirySelect from "@/components/upload/ExpirySelect.vue";
import PasswordInput from "@/components/upload/PasswordInput.vue";
import { useConfigStore } from "@/stores/config.ts";
import { useUploadStore } from "@/stores/upload.ts";
import ContentPasteIcon from "~icons/material-symbols/content-paste-rounded";
import InfoIcon from "~icons/material-symbols/info-rounded";
import PublishedChangesIcon from "~icons/material-symbols/published-with-changes-rounded";

const config = useConfigStore();
const upload = useUploadStore();
const router = useRouter();
const showAuth = ref(false);
const canOverwriteExisting = computed(() => !!config.editTargetFilename && !!config.editDeleteKey);

const doUpload = async () => {
  let finalFilename = config.filename;
  const defaultExtension = config.extension || 'txt'; // Use stored default or fallback to 'txt'

  const lastDotIndex = finalFilename.lastIndexOf('.');
  // If no dot is found, or the dot is at the beginning (hidden file), or ends with a dot, append the default extension.
  if (lastDotIndex === -1 || lastDotIndex === 0 || finalFilename.endsWith('.')) {
    finalFilename = `${finalFilename}.${defaultExtension}`;
  }
  // Otherwise, assume the user provided the extension correctly.

  const file = new File([config.content], finalFilename);
  try {
    const res =
      config.overwrite && canOverwriteExisting.value
        ? await upload.overwriteFile({
            file,
            filename: config.editTargetFilename,
            deleteKey: config.editDeleteKey,
            expiry: config.expiry,
            password: config.password,
            saveOriginalName: false,
          })
        : await upload.uploadFile({
            file,
            expiry: config.expiry,
            password: config.password,
            saveOriginalName: false,
          });

    if (
      config.overwrite &&
      canOverwriteExisting.value &&
      res.filename !== config.editTargetFilename
    ) {
      toast.warning("Could not overwrite original link.", {
        description: `Created ${res.filename} instead.`,
      });
    }

    config.content = "";
    config.filename = ""; // Reset filename input
    config.extension = "txt"; // Reset extension to default
    config.editTargetFilename = "";
    config.editDeleteKey = "";
    config.overwrite = false;
    await router.push(`/${res.filename}`);
  } catch (err) {
    console.error(err);
    if (isAxiosError(err) && err.response?.status === 401) {
      showAuth.value = true;
    }
  }
};

const { Ctrl_Enter, Meta_Enter } = useMagicKeys();

const ctrlEnter = Ctrl_Enter ?? ref(false);
const metaEnter = Meta_Enter ?? ref(false);

watch(ctrlEnter, (pressed) => pressed && doUpload());
watch(metaEnter, (pressed) => pressed && doUpload());
watch(canOverwriteExisting, (can) => {
  if (!can) config.overwrite = false;
});

const textarea = ref();
onMounted(() => textarea.value.$el.focus());

const loadFile = async (file: File) => {
  if (file.size > 1024 * 1024) return;
  // When loading a file, parse its name and extension
  const parts = file.name.split('.');
  config.extension = parts.pop() || 'txt'; // Get last part as extension, default to 'txt'
  config.filename = parts.join('.'); // Join remaining parts as filename
  config.content = await file.text();
};

useDropZone(document, {
  dataTypes(t) {
    const type = t[0];
    if (!type) return false;
    return type.startsWith("text/") || type === "application/json" || type.endsWith("yaml");
  },
  async onDrop(files) {
    if (!files?.length) return;
    await loadFile(files[0] as File);
  },
  preventDefaultForUnhandled: true,
});

useEventListener(window, "paste", async (e: ClipboardEvent) => {
  if (!e.clipboardData?.files?.length) return;
  await loadFile(e.clipboardData.files[0] as File);
});
</script>
