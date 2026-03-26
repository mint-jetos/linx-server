import { useConfigStore } from "@/stores/config.ts";

export const ApiPath = (path = "") => {
  const config = useConfigStore();
  const base = (config.site?.site_path || "/").replace(/\/+$/, "");
  const u = new URL(path, window.location.origin);
  if (path.match(/^\//)) {
    u.pathname = base + u.pathname;
  }
  return u.toString();
};
