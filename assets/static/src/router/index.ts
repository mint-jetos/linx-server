import { createRouter, createWebHistory } from "vue-router";
import FileView from "@/views/FileView.vue";

import UploadView from "@/views/UploadView.vue";

const router = createRouter({
  history: createWebHistory(window.config?.site_path),
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
      path: "/users",
      name: "Admin Login",
      component: () => import("../views/UsersView.vue"),
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
    document.title = String(to.name ?? "") + " · " + window.config.site_name;
  }
});

export default router;
