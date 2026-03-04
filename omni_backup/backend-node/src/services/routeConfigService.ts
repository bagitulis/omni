/**
 * Route Config Service
 * Business logic untuk route configuration management
 * Single Responsibility: Handle all route config operations
 * All operations are tenant-scoped
 */

import { getPrisma } from "./prismaClient";
import { getLogger } from "../utils/logger";
import { tenantContext } from "../utils/tenantContext";

const logger = getLogger("RouteConfigService");

export interface RouteConfigInput {
  routePath: string;
  routeMethod?: string;
  routeName?: string;
  description?: string;
  category?: string;
  enabled?: boolean;
  cachingEnabled?: boolean;
  cacheTTL?: number;
  cacheStrategy?: string;
  queueEnabled?: boolean;
  queueMaxSize?: number;
  queuePriority?: string;
  maxConcurrent?: number;
  rateLimitEnabled?: boolean;
  rateLimitWindow?: number;
  rateLimitMax?: number;
  minIntervalMs?: number;
  timeout?: number;
  retryEnabled?: boolean;
  maxRetries?: number;
  retryDelayMs?: number;
  customConfig?: any;
}

export class RouteConfigService {
  async getAllRoutes() {
    const tenantId = tenantContext.getTenantId();
    return getPrisma().routeConfig.findMany({
      where: { tenantId },
      orderBy: { category: "asc" },
    });
  }

  async getRouteById(id: string) {
    const tenantId = tenantContext.getTenantId();
    const route = await getPrisma().routeConfig.findUnique({
      where: { id },
    });
    // Verify tenant ownership
    if (route && route.tenantId !== tenantId) {
      return null;
    }
    return route;
  }

  async getRoutesByCategory(category: string) {
    const tenantId = tenantContext.getTenantId();
    return getPrisma().routeConfig.findMany({
      where: { tenantId, category },
      orderBy: { routePath: "asc" },
    });
  }

  async getCategories() {
    const tenantId = tenantContext.getTenantId();
    const configs = await getPrisma().routeConfig.findMany({
      where: { tenantId },
      distinct: ["category"],
      select: { category: true },
    });

    return configs
      .map((c: { category: string | null }) => c.category)
      .filter((c: string | null): c is string => c !== null)
      .sort();
  }

  async createRoute(input: RouteConfigInput) {
    const tenantId = tenantContext.getTenantId();
    const data = {
      tenantId,
      routePath: input.routePath,
      routeMethod: input.routeMethod || "GET",
      routeName: input.routeName,
      description: input.description,
      category: input.category,
      enabled: input.enabled ?? true,
      cachingEnabled: input.cachingEnabled ?? true,
      cacheTTL: input.cacheTTL || 300,
      cacheStrategy: input.cacheStrategy || "standard",
      queueEnabled: input.queueEnabled ?? true,
      queueMaxSize: input.queueMaxSize || 100,
      queuePriority: input.queuePriority || "normal",
      maxConcurrent: input.maxConcurrent || 5,
      rateLimitEnabled: input.rateLimitEnabled || false,
      rateLimitWindow: input.rateLimitWindow || 60,
      rateLimitMax: input.rateLimitMax || 100,
      minIntervalMs: input.minIntervalMs || 0,
      timeout: input.timeout || 30000,
      retryEnabled: input.retryEnabled ?? true,
      maxRetries: input.maxRetries || 3,
      retryDelayMs: input.retryDelayMs || 1000,
      customConfig: input.customConfig ? JSON.stringify(input.customConfig) : null,
    };

    try {
      const route = await getPrisma().routeConfig.create({ data });
      logger.info(`✅ [${tenantId}] Created route: ${input.routePath}`);
      return route;
    } catch (error: any) {
      if (error.code === "P2002") {
        throw new Error("Route path already exists");
      }
      throw error;
    }
  }

  async updateRoute(id: string, input: Partial<RouteConfigInput>) {
    const tenantId = tenantContext.getTenantId();
    // Verify ownership before update
    const existing = await this.getRouteById(id);
    if (!existing) {
      throw new Error("Route not found or access denied");
    }

    try {
      const data = { ...input };
      if (data.customConfig && typeof data.customConfig === "object") {
        data.customConfig = JSON.stringify(data.customConfig) as any;
      }

      const route = await getPrisma().routeConfig.update({
        where: { id },
        data,
      });
      logger.info(`✅ [${tenantId}] Updated route: ${id}`);
      return route;
    } catch (error: any) {
      if (error.code === "P2025") {
        throw new Error("Route not found");
      }
      throw error;
    }
  }

  async deleteRoute(id: string) {
    const tenantId = tenantContext.getTenantId();
    // Verify ownership before delete
    const existing = await this.getRouteById(id);
    if (!existing) {
      throw new Error("Route not found or access denied");
    }

    try {
      const route = await getPrisma().routeConfig.delete({
        where: { id },
      });
      logger.info(`✅ [${tenantId}] Deleted route: ${id}`);
      return route;
    } catch (error: any) {
      if (error.code === "P2025") {
        throw new Error("Route not found");
      }
      throw error;
    }
  }

  async bulkUpdateRoutes(routeIds: string[], updates: Partial<RouteConfigInput>) {
    if (!routeIds || routeIds.length === 0) {
      throw new Error("Route IDs required");
    }
    
    const tenantId = tenantContext.getTenantId();

    const data = { ...updates };
    if (data.customConfig && typeof data.customConfig === "object") {
      data.customConfig = JSON.stringify(data.customConfig) as any;
    }

    // Verify ownership before bulk update
    const routes = await Promise.all(
      routeIds.map(async (id) => {
        const existing = await this.getRouteById(id);
        if (!existing) {
          throw new Error(`Route ${id} not found or access denied`);
        }
        return getPrisma().routeConfig.update({
          where: { id },
          data,
        });
      })
    );

    logger.info(`✅ [${tenantId}] Bulk updated ${routes.length} routes`);
    return routes;
  }

  async applyPreset(
    routeIds: string[],
    preset: "aggressive" | "balanced" | "minimal" | "realtime"
  ) {
    if (!routeIds || routeIds.length === 0) {
      throw new Error("Route IDs required");
    }

    const presets: Record<string, any> = {
      aggressive: {
        cachingEnabled: true,
        cacheTTL: 600,
        cacheStrategy: "aggressive",
        queueEnabled: true,
        maxConcurrent: 10,
        rateLimitEnabled: false,
        timeout: 60000,
        retryEnabled: true,
        maxRetries: 5,
      },
      balanced: {
        cachingEnabled: true,
        cacheTTL: 300,
        cacheStrategy: "standard",
        queueEnabled: true,
        maxConcurrent: 5,
        rateLimitEnabled: true,
        rateLimitWindow: 60,
        rateLimitMax: 100,
        timeout: 30000,
        retryEnabled: true,
        maxRetries: 3,
      },
      minimal: {
        cachingEnabled: false,
        queueEnabled: false,
        rateLimitEnabled: false,
        timeout: 10000,
        retryEnabled: false,
      },
      realtime: {
        cachingEnabled: false,
        queueEnabled: true,
        maxConcurrent: 1,
        minIntervalMs: 0,
        timeout: 5000,
        retryEnabled: false,
      },
    };

    const config = presets[preset];
    if (!config) {
      throw new Error(
        `Unknown preset: ${preset}. Available: aggressive, balanced, minimal, realtime`
      );
    }

    const routes = await this.bulkUpdateRoutes(routeIds, config);
    logger.info(`✅ Applied preset '${preset}' to ${routes.length} routes`);
    return routes;
  }
}

export const routeConfigService = new RouteConfigService();
