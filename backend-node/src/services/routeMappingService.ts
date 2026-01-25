/**
 * Route Mapping Service - Facade/Orchestrator
 * Delegates analysis to RouteAnalysisEngine
 */

import * as path from "path";
import { RouteAnalysisEngine } from "./routeAnalysisEngine";
import { ComponentScanner } from "./routeScannerService";
import {
  CATEGORY_LABELS,
  formatComponents,
  formatFrontendOnly,
  formatRouteArray,
  normalizeCategoryKey,
  toCategoryRecord,
} from "./routeMappingFormatter";
import {
  AnalysisResult,
  RouteCategories,
  RouteInfo,
  FrontendOnlyRoute,
  ComponentInfo,
} from "./routeTypes";

interface RouteReport {
  success: boolean;
  totalRoutes: number;
  totalComponents: number;
  totalCategories: number;
  totalDynamicRoutes: number;
  totalCalledRoutes: number;
  totalDisconnectedRoutes: number;
  totalUnusedRoutes: number;
  connectionRate: string;
  byCategory: Record<string, RouteInfo[] | FrontendOnlyRoute[]>;
  categoryLabels: typeof CATEGORY_LABELS;
  categories: RouteCategories;
  components: Record<string, ComponentInfo>;
  disconnectedRoutes: Record<string, any>;
  backendOnlyRoutes: Record<string, any>;
  unusedRoutes: Record<string, any>;
  buttonToEndpoints: Record<string, string[]>;
  timestamp: string;
}

export class RouteMappingAnalyzer {
  private engine: RouteAnalysisEngine;
  private buttonToEndpoints: Map<string, string[]> = new Map();

  constructor() {
    this.engine = new RouteAnalysisEngine();
  }

  public analyze(): RouteReport {
    const analysis: AnalysisResult = this.engine.analyze();
    const { routes, components, connectedRoutes, categories } = analysis;

    const byCategory = toCategoryRecord(categories);
    const totalRoutes = routes.size;
    const totalConnected = connectedRoutes.size;
    const totalDisconnected = categories.frontendOnly.length;
    const totalUnused = categories.unused.length;
    const connectionRate =
      totalRoutes > 0
        ? `${((totalConnected / totalRoutes) * 100).toFixed(1)}%`
        : "0%";

    const totalDynamicRoutes = Array.from(routes.values()).filter(
      (r: RouteInfo) => Boolean(r.isDynamic)
    ).length;

    return {
      success: true,
      totalRoutes,
      totalComponents: components.size,
      totalCategories: Object.keys(byCategory).length,
      totalDynamicRoutes,
      totalCalledRoutes: totalConnected,
      totalDisconnectedRoutes: totalDisconnected,
      totalUnusedRoutes: totalUnused,
      connectionRate,
      byCategory,
      categoryLabels: CATEGORY_LABELS,
      categories,
      components: formatComponents(components),
      disconnectedRoutes: formatFrontendOnly(categories.frontendOnly),
      backendOnlyRoutes: formatRouteArray(
        categories.backendOnly,
        "BACKEND_ONLY"
      ),
      unusedRoutes: formatRouteArray(categories.unused, "UNUSED"),
      buttonToEndpoints: Object.fromEntries(this.buttonToEndpoints),
      timestamp: new Date().toISOString(),
    };
  }

  public getByCategory(category?: string): Record<string, any> {
    const report = this.analyze();
    const normalized = normalizeCategoryKey(category);

    if (normalized && report.byCategory[normalized]) {
      return {
        success: true,
        category: normalized,
        label: CATEGORY_LABELS[normalized],
        routes: report.byCategory[normalized],
        total: report.byCategory[normalized].length,
      };
    }

    const categories = Object.keys(report.byCategory);
    return {
      success: true,
      categories,
      labels: CATEGORY_LABELS,
      by_category: report.byCategory,
      total_categories: categories.length,
    };
  }

  public getByComponent(componentName?: string): Record<string, any> {
    const report = this.analyze();

    if (componentName && report.components[componentName]) {
      return {
        success: true,
        component: componentName,
        data: report.components[componentName],
      };
    }

    return {
      success: true,
      components: report.components,
      total_components: Object.keys(report.components).length,
    };
  }

  public getDetailedMapping(): RouteReport {
    return this.analyze();
  }

  public getStatistics(): Record<string, any> {
    const report = this.analyze();

    const routesByMethod: Record<string, number> = {};
    const routedEntries = [
      ...report.categories.connected,
      ...report.categories.backendOnly,
      ...report.categories.unused,
    ];

    routedEntries.forEach((route) => {
      if (!route.method) return;
      routesByMethod[route.method] = (routesByMethod[route.method] || 0) + 1;
    });

    const componentsByCategory: Record<string, number> = {};
    Object.values(report.components).forEach((comp) => {
      componentsByCategory[comp.category] =
        (componentsByCategory[comp.category] || 0) + 1;
    });

    const dynamicRoutePercentage =
      report.totalRoutes > 0
        ? `${(
            (report.totalDynamicRoutes / Math.max(report.totalRoutes, 1)) *
            100
          ).toFixed(1)}%`
        : "0%";

    return {
      success: true,
      routesByMethod,
      componentsByCategory,
      averageRoutesPerComponent:
        report.totalRoutes > 0
          ? (report.totalRoutes / report.totalComponents).toFixed(2)
          : 0,
      dynamicRoutePercentage,
      componentConnectionRate: report.connectionRate,
      totalUnusedRoutes: report.totalUnusedRoutes,
      timestamp: report.timestamp,
    };
  }

  public enableAutoDetection(): void {
    try {
      const routeScannerService = require("./routeScannerService");
      const RouteScanner = routeScannerService.RouteScanner;

      const routesPath = path.join(__dirname, "../routes");
      RouteScanner.scanRoutesFromFolder(routesPath);

      const frontendPath = path.join(
        __dirname,
        "../../..",
        "frontend",
        "src",
        "components"
      );
      ComponentScanner.scanComponentsFromFolder(frontendPath);
    } catch (error) {
      // Auto-detection failed silently
    }
  }
}

// Singleton instance
let instance: RouteMappingAnalyzer | null = null;

export function getRouteMappingService(): RouteMappingAnalyzer {
  if (!instance) {
    instance = new RouteMappingAnalyzer();
  }
  return instance;
}
