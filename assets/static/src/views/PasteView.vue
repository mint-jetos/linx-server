<template>
  <form @submit.prevent="doUpload" class="space-y-4">
    <div class="flex flex-row items-center gap-2">

      <Input
        v-model="config.filename"
        @input="config.randomFilename = false"
        placeholder="my_file.txt"
        class="w-full sm:flex-1"
        aria-label="Filename"
        :disabled="config.overwrite && canOverwriteExisting"
      />

      <Label v-if="!config.site?.force_random" class="flex flex-row items-center gap-2 px-2 shrink-0 cursor-pointer">
        <Switch v-model="config.randomFilename" />
        <span class="text-xs whitespace-nowrap">Random</span>
      </Label>

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
      <Button type="button" size="icon" class="shrink-0" :disabled="isPasteButtonDisabled" @click="handlePasteButtonClick">
        <component :is="computedButtonIcon" class="text-xl" />
        <span class="sr-only">{{ computedButtonText }}</span>
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
      class="font-mono h-50"
      autofocus
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      :disabled="isContentLocked"
    />
  </form>

  <AuthDialog v-if="config.site?.auth" v-model="showAuth" @submit="doUpload" />
</template>

<script setup lang="ts">
import { useDropZone, useEventListener, useMagicKeys } from "@vueuse/core";
import { isAxiosError } from "axios";
import { computed, onMounted, ref, watch } from "vue";
import { toast } from "vue-sonner";
import { Alert, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
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
import EditIcon from "~icons/material-symbols/edit-rounded"; // Import EditIcon

const config = useConfigStore();
const upload = useUploadStore();
const showAuth = ref(false);
const isContentLocked = ref(false); // New state variable
const currentGeneratedFilename = ref(''); // New state variable
const canOverwriteExisting = computed(() => !!config.editTargetFilename && !!config.editDeleteKey);

const isPasteButtonDisabled = computed(() => {
  if (isContentLocked.value) { // In Edit mode, button is always enabled to switch back
    return false;
  }
  return !config.content; // In Paste mode, disabled if content is empty
});

const computedButtonIcon = computed(() => isContentLocked.value ? EditIcon : ContentPasteIcon);
const computedButtonText = computed(() => isContentLocked.value ? 'Edit' : 'Paste');

const handlePasteButtonClick = async () => {
  if (isContentLocked.value) { // If currently in 'Edit' mode
    isContentLocked.value = false; // Unlock textarea
    // Optional: focus textarea
    if (textarea.value) {
      textarea.value.$el.focus();
    }
  } else { // If currently in 'Paste' mode
    await doUpload();
  }
};

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
          })
        : await upload.uploadFileWS({
            file,
            randomFilename: config.randomFilename,
            expiry: config.expiry,
            password: config.password,
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

    isContentLocked.value = true;
    config.filename = res.filename; // Populate filename input
    currentGeneratedFilename.value = res.filename;
    config.editTargetFilename = res.filename; // Store for overwrite
    config.editDeleteKey = res.delete_key;    // Store for overwrite
    config.overwrite = true;                  // Default to overwrite for next action
    // config.content remains, but textarea is disabled
    // config.extension = "txt"; // No need to reset
    // router.push is removed
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
  config.randomFilename = false; // Disable random filename when a name is provided
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
