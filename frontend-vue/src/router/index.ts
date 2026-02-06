import { createRouter, createWebHistory } from "vue-router";
import { useRouteState } from "../composables/useRouteState";
import { useNavigationQueue } from "../composables/useNavigationQueue";
import { usePerformanceMonitor } from "../composables/usePerformanceMonitor";

// Extend Router interface for custom properties
declare module "vue-router" {
  interface Router {
    routeStateManager?: ReturnType<typeof useRouteState>;
    navigationQueue?: ReturnType<typeof useNavigationQueue>;
    performanceMonitor?: ReturnType<typeof usePerformanceMonitor>;
  }
}

import { routes } from "./routes";
import {
  setupNavigationGuards,
  attachManagersToRouter,
  routeStateManager,
  navigationQueue,
  performanceMonitor,
} from "./guards";

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition;
    } else {
      return { top: 0 };
    }
  },
});

// Setup navigation guards
setupNavigationGuards(router);

// Attach managers for external access
attachManagersToRouter(router);

export default router;
export { routeStateManager, navigationQueue, performanceMonitor };
