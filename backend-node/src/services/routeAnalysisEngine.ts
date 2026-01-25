import * as path from "path";
import { RouteScanner, ComponentScanner } from "./routeScannerService";
import {
  AnalysisResult,
  ComponentInfo,
  RouteCategories,
  RouteInfo,
} from "./routeTypes";

interface ScannedRoute {
  endpoint: string;
  method: string;
  fileName: string;
}

/**
 * Route Analysis Engine - Pure analysis logic
 * Scans routes/components and detects connections
 */
export class RouteAnalysisEngine {
  analyze(): AnalysisResult {
    const scannedRoutes = new Map<string, RouteInfo>();
    const scannedComponents = new Map<string, ComponentInfo>();

    try {
      const backendRoutes = RouteScanner.scanRoutesFromFolder(
        path.join(__dirname, "../routes")
      );

      const frontendCalls = ComponentScanner.scanComponentsFromFolder(
        path.join(__dirname, "../../../frontend/src/components")
      );

      const mountPrefixes: Record<string, string> = {
        "health.ts": "/api",
        "tokenStatus.ts": "/api",
        "tokenOperation.ts": "/api",
        "orderSync.ts": "/api/orders",
        "settings.ts": "/api/settings",
        "googleSheets.ts": "/api/google",
        "routeMapping.ts": "/api/route-mapping",
      };

      if (backendRoutes.length > 0) {
        backendRoutes.forEach((route: ScannedRoute) => {
          const endpoint = this.normalizeEndpoint(
            route.endpoint,
            route.fileName,
            mountPrefixes
          );

          const method = (route.method || "GET").toUpperCase();
          const key = `${method} ${endpoint}`;

          scannedRoutes.set(key, {
            endpoint,
            method,
            isDynamic: this.isDynamicEndpoint(endpoint),
            category: "general",
            source: route.fileName,
          });
        });
      }

      const componentKeys = Object.keys(frontendCalls);
      if (componentKeys.length > 0) {
        componentKeys.forEach((componentName) => {
          const apiCalls = frontendCalls[componentName] || [];
          scannedComponents.set(componentName, {
            name: componentName,
            category: "general",
            path: componentName,
            routesCalled: apiCalls.map((call) =>
              this.normalizeRouteCall(call)
            ),
            buttons: [],
          });
        });
      }
    } catch (error) {
      // Auto-detection failed, return empty
    }

    const connectedRoutes = new Set<string>();
    const disconnectedCalls = new Map<string, string[]>();

    scannedComponents.forEach((component) => {
      component.routesCalled.forEach((routeCall) => {
        const normalizedCall = this.normalizeRouteCall(routeCall);

        let found = false;
        scannedRoutes.forEach((route, key) => {
          if (
            this.matchesRoute(routeCall, route) ||
            this.matchesRoute(normalizedCall, route)
          ) {
            connectedRoutes.add(key);
            found = true;
            return;
          }
        });

        if (!found) {
          if (!disconnectedCalls.has(normalizedCall)) {
            disconnectedCalls.set(normalizedCall, []);
          }
          disconnectedCalls.get(normalizedCall)!.push(component.name);
        }
      });
    });

    const unusedRoutes = new Map<string, RouteInfo>();
    scannedRoutes.forEach((route, key) => {
      if (!connectedRoutes.has(key)) {
        unusedRoutes.set(key, route);
      }
    });

    const categories = this.categorizeRoutes(
      scannedRoutes,
      connectedRoutes,
      disconnectedCalls,
      unusedRoutes
    );

    return {
      routes: scannedRoutes,
      components: scannedComponents,
      connectedRoutes,
      disconnectedCalls,
      unusedRoutes,
      categories,
    };
  }

  private normalizeEndpoint(
    endpoint: string,
    fileName: string,
    mountPrefixes: Record<string, string>
  ): string {
    let normalized = endpoint.trim();
    if (!normalized.startsWith("/")) {
      normalized = `/${normalized}`;
    }

    for (const [file, prefix] of Object.entries(mountPrefixes)) {
      if (fileName.includes(file)) {
        normalized = normalized.startsWith(prefix)
          ? normalized
          : `${prefix}${normalized}`;
        break;
      }
    }

    if (!normalized.startsWith("/api")) {
      normalized = `/api${normalized}`;
    }

    normalized = normalized.replace(/\/{2,}/g, "/");
    if (normalized.length > 1 && normalized.endsWith("/")) {
      normalized = normalized.slice(0, -1);
    }

    return normalized;
  }

  private normalizeRouteCall(routeCall: string): string {
    if (!routeCall) {
      return "";
    }

    let normalized = routeCall.replace(/\$\{[^}]+\}/g, ":param");
    normalized = normalized.replace(/\$\w+/g, ":param");

    if (!normalized.startsWith("/")) {
      normalized = `/${normalized}`;
    }

    normalized = normalized.replace(/\/{2,}/g, "/");
    return normalized;
  }

  private isDynamicEndpoint(endpoint: string): boolean {
    return endpoint.includes(":");
  }

  private matchesRoute(routeCall: string, route: RouteInfo): boolean {
    if (routeCall === route.endpoint) {
      return true;
    }

    if (route.isDynamic) {
      const routePattern = route.endpoint
        .replace(/:[^/]+/g, "[^/]+")
        .replace(/\//g, "\\/");
      const regex = new RegExp(`^${routePattern}$`);
      return regex.test(routeCall);
    }

    return false;
  }

  private categorizeRoutes(
    scannedRoutes: Map<string, RouteInfo>,
    connectedRoutes: Set<string>,
    disconnectedCalls: Map<string, string[]>,
    unusedRoutes: Map<string, RouteInfo>
  ): RouteCategories {
    const connected: RouteInfo[] = [];
    connectedRoutes.forEach((key) => {
      const route = scannedRoutes.get(key);
      if (route) {
        connected.push({ ...route, category: "connected" });
      }
    });

    const frontendOnly = Array.from(disconnectedCalls.entries()).map(
      ([endpoint, components]) => ({
        endpoint,
        components,
        status: "FRONTEND_CALL_ONLY",
      })
    );

    const backendOnly: RouteInfo[] = [];
    const unused: RouteInfo[] = [];

    unusedRoutes.forEach((route) => {
      const target = this.isBackendInternal(route) ? backendOnly : unused;
      target.push({
        ...route,
        category: this.isBackendInternal(route)
          ? "backend_only"
          : "unused",
      });
    });

    return { connected, frontendOnly, backendOnly, unused };
  }

  private isBackendInternal(route: RouteInfo): boolean {
    const endpoint = route.endpoint.toLowerCase();
    const source = (route.source || "").toLowerCase();
    const keywords = [
      "health",
      "token",
      "sync",
      "cron",
      "webhook",
      "callback",
      "internal",
      "status",
      "stats",
      "monitor",
      "queue",
      "import",
      "export",
    ];

    return keywords.some((key) => endpoint.includes(key) || source.includes(key));
  }
}
