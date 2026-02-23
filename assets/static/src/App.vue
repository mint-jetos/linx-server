<template>
  <div class="flex flex-col min-h-screen min-h-svh bg-background text-foreground">
    <TooltipProvider>
      <header class="grid grid-cols-3 border-b bg-surface px-4 py-2 items-center">
        <div>
          <Button variant="link" class="p-0" as-child>
            <RouterLink to="/">
              <h1 class="text-2xl font-semibold">{{ config.site.site_name }}</h1>
            </RouterLink>
          </Button>
        </div>

        <div class="justify-self-center flex space-x-2">
          <Button
            variant="ghost"
            :class="{ 'bg-muted': ui.activeRootView === 'upload' }"
            @click="ui.setActiveRootView('upload')"
          >
            Upload
          </Button>
          <Button
            variant="ghost"
            :class="{ 'bg-muted': ui.activeRootView === 'paste' }"
            @click="ui.setActiveRootView('paste')"
          >
            Paste
          </Button>
        </div>

        <div class="justify-self-end">
          <Tooltip>
            <TooltipTrigger as-child :key="mode">
              <Button variant="ghost" @click="mode = nextMode" class="rounded-full">
                <component :is="modeIcon" />
                <span class="sr-only">Change to {{ nextMode }} mode</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent>Change to {{ nextMode }} mode</TooltipContent>
          </Tooltip>

        </div>
      </header>

      <main class="flex-1 content-center p-6">
        <router-view />
      </main>
    </TooltipProvider>

    <Toaster />
  </div>
</template>

<script setup lang="ts">
import { useColorMode } from "@vueuse/core";
import { computed } from "vue";

import { Button } from "@/components/ui/button";
import { Toaster } from "@/components/ui/sonner";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip/index.js";
import { useConfigStore } from "@/stores/config.js";
import DarkIcon from "~icons/material-symbols/brightness-2-rounded";
import LightIcon from "~icons/material-symbols/brightness-5-rounded";
import AutoIcon from "~icons/material-symbols/brightness-auto-rounded";
import { useUIStore } from "@/stores/ui.ts"; // Import useUIStore

const config = useConfigStore();
const ui = useUIStore(); // Initialize ui store



const mode = useColorMode({ disableTransition: false, emitAuto: true });

const nextMode = computed(() => {
  if (mode.value === "auto") return "dark";
  if (mode.value === "dark") return "light";
  return "auto";
});

const modeIcon = computed(() => {
  if (mode.value === "auto") return AutoIcon;
  if (mode.value === "light") return LightIcon;
  return DarkIcon;
});
</script>
