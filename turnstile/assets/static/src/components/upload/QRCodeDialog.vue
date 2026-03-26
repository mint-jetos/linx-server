<template>
  <DialogContent class="sm:max-w-[280px] p-4 gap-2">
    <div class="flex flex-col items-center justify-center gap-2">
      <div class="p-3 rounded-xl bg-white/5 shadow-inner">
        <QRCodeVue
          :value="url"
          :size="200"
          level="H"
          render-as="svg"
          background="transparent"
          foreground="oklch(0.9 0 0)"
        />
      </div>
      <div class="w-full">
        <Label for="link" class="sr-only">Link</Label>
        <Input
          id="link"
          :model-value="url.split('/').pop()"
          readonly
          tabindex="-1"
          class="text-center text-xs focus-visible:ring-0 border-none bg-transparent h-8"
        />
      </div>
    </div>
  </DialogContent>
</template>

<script setup lang="ts">
import QRCodeVue from "qrcode.vue";
import { DialogContent } from "@/components/ui/dialog/index.js";
import { Input } from "@/components/ui/input/index.js";
import { Label } from "@/components/ui/label/index.js";

defineProps<{
  url: string;
}>();
</script>
