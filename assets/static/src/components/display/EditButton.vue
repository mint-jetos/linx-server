<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <Button variant="outline" @click="edit" v-bind="$attrs">
        <EditIcon class="text-2xl" />
        <span class="sr-only">Edit</span>
      </Button>
    </TooltipTrigger>
    <TooltipContent side="bottom">Edit</TooltipContent>
  </Tooltip>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button/index.js";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { useConfigStore } from "@/stores/config.ts";
import { useUploadStore } from "@/stores/upload.ts";
import { getExtension } from "@/util/extensions.ts";
import EditIcon from "~icons/material-symbols/edit-rounded";
import { useUIStore } from "@/stores/ui.ts"; // Import useUIStore

const props = defineProps({
  meta: { type: Object, required: true },
  content: { type: String, required: true },
});

const config = useConfigStore();
const upload = useUploadStore();
const router = useRouter();
const ui = useUIStore(); // Initialize ui store

const edit = () => {
  const existing = upload.uploads.find((item) => item.filename === props.meta.filename);
  const preferredName = props.meta.original_name || props.meta.filename;

  config.extension = getExtension(preferredName);
  config.filename = preferredName.split(".").slice(0, -1).join(".");
  config.content = props.content;
  config.editTargetFilename = props.meta.filename;
  config.editDeleteKey = existing?.delete_key ?? "";
  config.overwrite = !!existing?.delete_key;
  ui.setActiveRootView('paste'); // Set active view to paste
  router.push("/"); // Navigate to root path
};
</script>
