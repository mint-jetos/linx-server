<template>
  <div class="flex flex-col min-h-screen min-h-svh bg-background text-foreground">
    <TooltipProvider>
      <header class="grid grid-cols-3 border-b bg-surface px-4 py-2 items-center">
        <div>
          <Button v-if="ui.isAuthenticated || (router.currentRoute.value.name !== 'File' && router.currentRoute.value.name !== 'Admin Login')" variant="link" class="p-0" as-child>
            <RouterLink to="/" @click="ui.setActiveRootView('upload')" :class="{ 'opacity-0 pointer-events-none': ui.isDeadLink }">
              <h1 class="text-2xl font-semibold">{{ config.site.site_name }}</h1>
            </RouterLink>
          </Button>
          <h1 v-else class="text-2xl font-semibold px-4">{{ config.site.site_name }}</h1>
        </div>

        <div :style="{ visibility: isRouterReady && router.currentRoute.value.path === '/' ? 'visible' : 'hidden' }" class="justify-self-center flex space-x-2">
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

      <main class="flex-1 p-6">
        <router-view />
      </main>
    </TooltipProvider>

    <Toaster />
  </div>
</template>

<script setup lang="ts">
import { useColorMode } from "@vueuse/core";
import { computed, watch, ref } from "vue";

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
import { useRouter } from "vue-router"; // Import useRouter

const config = useConfigStore();
const ui = useUIStore(); // Initialize ui store
const router = useRouter(); // Initialize router
const isRouterReady = ref(false);

router.isReady().then(() => {
  isRouterReady.value = true;
});

watch(
  () => router.currentRoute.value.name,
  (routeName) => {
    // Set to true for 'File' route to prevent flicker, false for others.
    // DisplayPage will set it to false if the file loads successfully.
    ui.isDeadLink = routeName === 'File';
  }
);

const mode = useColorMode({ disableTransition: false, emitAuto: true, initialValue: 'dark' });

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
