<template>
  <div
    :class="
      cn(
        'container flex flex-col justify-center gap-6 mx-auto transition-all duration-300',
        ui.activeRootView === 'paste' ? 'max-w-4xl' : 'max-w-2xl',
      )
    "
    v-bind="$attrs"
  >
    <Card>
      <CardContent class="flex flex-col gap-4 pt-6">
        <div v-if="ui.activeRootView === 'upload'">
          <div class="flex flex-row items-center gap-2">
            <Label v-if="!config.site?.force_random">
              <Switch v-model="config.randomFilename" />
              Random filename
            </Label>
            <PasswordInput v-model="config.password" class="w-full sm:flex-1" />
            <ExpirySelect
              v-model="config.expiry"
              :options="config.site?.expiration_times"
              class="w-full sm:w-40"
            />
          </div>
          <DropZone @upload="doUpload" :max-file-size="config.site?.max_size" />
        </div>
        <div v-else-if="ui.activeRootView === 'paste'">
          <PasteView />
        </div>
      </CardContent>
    </Card>

    <UploadList v-model:show-auth="showAuth" />
  </div>

  <AuthDialog v-if="config.site?.auth" v-model="showAuth" @submit="doUpload(retryFile)" />
</template>

<script setup lang="ts">
import { isAxiosError } from "axios";
import { ref } from "vue";
import { Card, CardContent } from "@/components/ui/card/index.js";
import { Label } from "@/components/ui/label/index.js";
import { Switch } from "@/components/ui/switch/index.js";
import AuthDialog from "@/components/upload/AuthDialog.vue";
import DropZone from "@/components/upload/DropZone.vue";
import ExpirySelect from "@/components/upload/ExpirySelect.vue";
import PasswordInput from "@/components/upload/PasswordInput.vue";
import UploadList from "@/components/upload/UploadList.vue";
import { cn } from "@/lib/utils";
import { useConfigStore } from "@/stores/config.ts";
import { useUploadStore } from "@/stores/upload.ts";
import { useUIStore } from "@/stores/ui.ts"; // Import useUIStore
import PasteView from "./PasteView.vue"; // Import PasteView

const config = useConfigStore();
const uploads = useUploadStore();
const ui = useUIStore(); // Access the ui store
const showAuth = ref(false);
let retryFile: File | undefined;

const doUpload = async (file: File | undefined) => {
  if (!file) return;
  retryFile = undefined;
  try {
    await uploads.uploadFile({
      file,
      randomFilename: config.randomFilename,
      expiry: config.expiry,
      password: config.password,
    });
  } catch (err) {
    console.error(err);
    if (isAxiosError(err) && err.response?.status === 401) {
      retryFile = file;
      showAuth.value = true;
    }
  }
};
</script>
