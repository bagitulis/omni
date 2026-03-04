import { useRoute } from "vue-router";

export function useSidebarNavigation() {
  const route = useRoute();

  const isActive = (section: string): boolean => {
    const currentPath = route.path;
    return (() => {
      switch (section) {
        case "operation":
          return currentPath.startsWith("/operation");
        case "product-manager":
          return currentPath.startsWith("/product-manager");
        case "order-manager":
          return currentPath.startsWith("/order-manager");
        case "inventory":
          return currentPath === "/inventory";
        case "settings":
          return currentPath.startsWith("/settings");
        case "settings-google-sheets":
          return currentPath === "/settings/google-sheets";
        case "settings-logs":
          return currentPath === "/settings/logs";
        case "settings-resources":
          return currentPath === "/settings/resources";
        case "settings-script-monitor":
          return currentPath === "/settings/script-monitor";
        case "route-mapping":
          return currentPath === "/route-mapping";
        case "dashboard":
          return currentPath === "/";
        default:
          return false;
      }
    })();
  };

  return {
    isActive,
  };
}
