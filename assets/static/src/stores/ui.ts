import { defineStore } from "pinia";
import { ref } from "vue";

export const useUIStore = defineStore("ui", () => {
  const activeRootView = ref<"upload" | "paste">("upload");
  const isDeadLink = ref(false);

  const setActiveRootView = (view: "upload" | "paste") => {
    activeRootView.value = view;
  };

  return { activeRootView, isDeadLink, setActiveRootView };
});
