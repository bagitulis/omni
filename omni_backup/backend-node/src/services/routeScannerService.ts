/**
 * Enhanced Route Mapper dengan Auto-Detection
 * Fitur: Auto-scan routes, detect connections, track changes
 */

import * as fs from "fs";
import * as path from "path";

/**
 * Scanner untuk detect routes dari file
 */
export class RouteScanner {
  /**
   * Scan routes dari file dengan regex pattern
   */
  static scanRoutesFromFile(
    filePath: string
  ): Array<{ endpoint: string; method: string; fileName: string }> {
    try {
      const content = fs.readFileSync(filePath, "utf-8");
      const routes: Array<{
        endpoint: string;
        method: string;
        fileName: string;
      }> = [];
      const fileName = path.basename(filePath);

      // Pattern untuk Express routes: router.get('/path', ...), app.post('/path', ...)
      const routePattern =
        /(router|app)\.(get|post|put|delete|patch)\(['"`]([^'"`]+)['"`]/gi;
      let match;

      while ((match = routePattern.exec(content)) !== null) {
        routes.push({
          method: match[2].toUpperCase(),
          endpoint: match[3],
          fileName: fileName,
        });
      }

      return routes;
    } catch (error) {
      return [];
    }
  }

  /**
   * Recursively scan routes dari seluruh src/ folder
   */
  static scanRoutesRecursively(
    folderPath: string
  ): Array<{ endpoint: string; method: string; fileName: string }> {
    const routes: Array<{
      endpoint: string;
      method: string;
      fileName: string;
    }> = [];

    try {
      if (!fs.existsSync(folderPath)) {
        return [];
      }

      const walkDir = (dir: string) => {
        const items = fs.readdirSync(dir);

        items.forEach((item) => {
          const fullPath = path.join(dir, item);
          const stat = fs.statSync(fullPath);

          if (stat.isDirectory()) {
            // Skip node_modules, hidden folders, test-scripts, dist, build
            if (
              !item.startsWith(".") &&
              item !== "node_modules" &&
              item !== "test-scripts" &&
              item !== "dist" &&
              item !== "build"
            ) {
              walkDir(fullPath);
            }
          } else if (
            stat.isFile() &&
            (item.endsWith(".ts") || item.endsWith(".js"))
          ) {
            // Scan setiap .ts/.js file untuk route definitions
            const fileRoutes = this.scanRoutesFromFile(fullPath);
            routes.push(...fileRoutes);
          }
        });
      };

      walkDir(folderPath);

      return routes;
    } catch (error) {
      return [];
    }
  }

  /**
   * Scan routes dari folder (non-recursive, kept for backward compatibility)
   */
  static scanRoutesFromFolder(
    folderPath: string
  ): Array<{ endpoint: string; method: string; fileName: string }> {
    const routes: Array<{
      endpoint: string;
      method: string;
      fileName: string;
    }> = [];

    try {
      if (!fs.existsSync(folderPath)) {
        return [];
      }

      const files = fs.readdirSync(folderPath);

      files.forEach((file) => {
        if (file.endsWith(".ts") || file.endsWith(".js")) {
          const filePath = path.join(folderPath, file);
          const fileRoutes = this.scanRoutesFromFile(filePath);
          routes.push(...fileRoutes);
        }
      });

      return routes;
    } catch (error) {
      return [];
    }
  }
}

/**
 * Scanner untuk detect API calls dari frontend components
 */
export class ComponentScanner {
  /**
   * Scan API calls dari Vue/React component
   */
  static scanComponentCalls(filePath: string): string[] {
    try {
      const content = fs.readFileSync(filePath, "utf-8");
      const calls: Set<string> = new Set();

      // Pattern untuk fetch('/api/...') atau axios.get('/api/...')
      // Handle: fetch('/api/...'), fetch(`/api/...`), fetch("/api/...")
      // Also handle: axios.get('/api/...'), axios.post(`/api/...`), etc.
      const fetchPattern =
        /(fetch|axios\.(get|post|put|delete|patch))\(\s*['"`](\/api\/[^'"`]*?)['"`]/gi;

      let match;
      while ((match = fetchPattern.exec(content)) !== null) {
        let endpoint = match[3]; // The /api/... part

        // Template literal handling: /api/orders/${category} → /api/orders/:category
        if (endpoint.includes("${")) {
          // Replace ${variable} with :variable (Express param style)
          endpoint = endpoint.replace(/\$\{([^}]+)\}/g, ":$1");
        }

        if (endpoint.startsWith("/api/")) {
          calls.add(endpoint);
        }
      }

      return Array.from(calls);
    } catch (error) {
      return [];
    }
  }

  /**
   * Scan API calls dari folder
   */
  static scanComponentsFromFolder(
    folderPath: string
  ): Record<string, string[]> {
    const components: Record<string, string[]> = {};

    try {
      if (!fs.existsSync(folderPath)) {
        return {};
      }

      const walkSync = (dir: string) => {
        const files = fs.readdirSync(dir);

        files.forEach((file) => {
          const filePath = path.join(dir, file);
          const stat = fs.statSync(filePath);

          if (stat.isDirectory()) {
            walkSync(filePath);
          } else if (
            file.endsWith(".vue") ||
            file.endsWith(".tsx") ||
            file.endsWith(".jsx")
          ) {
            const calls = this.scanComponentCalls(filePath);
            if (calls.length > 0) {
              const componentName = file.replace(/\.(vue|tsx|jsx)$/, "");
              components[componentName] = calls;
            }
          }
        });
      };

      walkSync(folderPath);
      return components;
    } catch (error) {
      return {};
    }
  }
}

/**
 * Connection Detector - check apakah frontend call match dengan backend route
 */
export class ConnectionDetector {
  /**
   * Check apakah endpoint call match dengan registered route
   */
  static matchEndpoint(callEndpoint: string, registeredRoute: string): boolean {
    // Exact match
    if (callEndpoint === registeredRoute) {
      return true;
    }

    // Dynamic route match: /api/:platform/something match /api/shopee/something
    const routePattern = registeredRoute
      .replace(/:[^/]+/g, "[^/]+")
      .replace(/\//g, "\\/");
    const regex = new RegExp(`^${routePattern}$`);

    return regex.test(callEndpoint);
  }

  /**
   * Detect connections antara components dan routes
   */
  static detectConnections(
    components: Record<string, string[]>,
    routes: Array<{ endpoint: string; method: string }>
  ): {
    connected: number;
    disconnected: Array<{ component: string; endpoint: string }>;
  } {
    const disconnected: Array<{ component: string; endpoint: string }> = [];
    let connected = 0;

    Object.entries(components).forEach(([componentName, endpoints]) => {
      endpoints.forEach((callEndpoint) => {
        const found = routes.some((route) =>
          this.matchEndpoint(callEndpoint, route.endpoint)
        );

        if (found) {
          connected++;
        } else {
          disconnected.push({
            component: componentName,
            endpoint: callEndpoint,
          });
        }
      });
    });

    return { connected, disconnected };
  }
}
