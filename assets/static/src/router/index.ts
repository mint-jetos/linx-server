import { createRouter, createWebHistory } from "vue-router";
import FileView from "@/views/FileView.vue";
import { deobfuscate } from "@/util/obfuscate.ts";
import { useConfigStore } from "@/stores/config.ts";

import UploadView from "@/views/UploadView.vue";

let sitePath = "/";
try {
  if (window.obfuscatedConfig) {
    const conf = JSON.parse(deobfuscate(window.obfuscatedConfig));
    sitePath = conf.site_path || "/";
  }
} catch (e) {
  console.error("Router failed to parse obfuscated config", e);
}

const router = createRouter({
  history: createWebHistory(sitePath),
  routes: [
    {
      path: "/",
      name: "Upload",
      component: UploadView,
    },

    {
      path: "/api",
      name: "API",
      component: () => import("../views/APIView.vue"),
    },
    {
      path: "/login",
      name: "Admin Login",
      component: () => import("../views/LoginView.vue"),
    },
    {
      path: "/:filename(.*)",
      name: "File",
      component: FileView,
      props: true,
    },
  ],
});

router.beforeEach((to) => {
  if (to.name !== "File") {
    const config = useConfigStore();
    document.title = config.site?.site_name || "Linx";
  }
});

export default router;
