/**
 * TikTok Category Service
 * Handles category tree fetching and caching for product creation
 */

import { TiktokAPIClient } from "../api/clients/tiktokAPIClient";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";

interface Category {
  id: string;
  name: string;
  parentId?: string;
  isLeaf: boolean;
  children?: Category[];
}

interface CategoryAttribute {
  id: string;
  name: string;
  type: string;
  isRequired: boolean;
  values?: { id: string; name: string }[];
}

export class TiktokCategoryService {
  private apiClient: TiktokAPIClient;
  private logger: Logger;
  private categoryCache: Map<string, { data: Category[]; timestamp: number }> =
    new Map();
  private attributeCache: Map<
    string,
    { data: CategoryAttribute[]; timestamp: number }
  > = new Map();
  private CACHE_TTL = 24 * 60 * 60 * 1000; // 24 hours

  constructor(apiClient: TiktokAPIClient) {
    this.apiClient = apiClient;
    this.logger = getLogger("TiktokCategoryService");
  }

  /**
   * Get category tree from TikTok API
   * API: GET /product/202309/categories (categories endpoint is only in 202309)
   * For SEA (ID, SG, MY, VN, TH, PH) and US shops, use category_version=v2
   */
  async getCategoryTree(parentId?: string): Promise<Category[]> {
    const cacheKey = `tree_${parentId || "root"}`;
    const cached = this.categoryCache.get(cacheKey);

    if (cached && Date.now() - cached.timestamp < this.CACHE_TTL) {
      this.logger.info(`📦 Using cached category tree for ${cacheKey}`);
      return cached.data;
    }

    this.logger.info(
      `🔄 Fetching TikTok category tree ${parentId ? `for parent ${parentId}` : "(root)"}`,
    );

    try {
      const endpoint = "/product/202309/categories";
      const params: Record<string, any> = {
        category_version: "v2", // V2 required for SEA (ID, SG, MY, etc.) and US shops
      };

      if (parentId) {
        params.category_id = parentId;
      }

      const result = await this.apiClient.request(endpoint, "GET", params);

      if (!result?.data?.categories) {
        this.logger.warn("No categories found in response");
        return [];
      }

      const categories: Category[] = result.data.categories.map((cat: any) => ({
        id: cat.id,
        name: cat.local_name || cat.name,
        parentId: cat.parent_id,
        isLeaf: cat.is_leaf === true,
      }));

      this.categoryCache.set(cacheKey, {
        data: categories,
        timestamp: Date.now(),
      });
      this.logger.info(`✅ Fetched ${categories.length} categories`);

      return categories;
    } catch (error: any) {
      this.logger.error(`❌ Error fetching categories: ${error.message}`);
      throw error;
    }
  }

  /**
   * Get category attributes (required fields for product in this category)
   * API: GET /product/202309/categories/{category_id}/attributes
   */
  async getCategoryAttributes(
    categoryId: string,
  ): Promise<CategoryAttribute[]> {
    const cacheKey = `attrs_${categoryId}`;
    const cached = this.attributeCache.get(cacheKey);

    if (cached && Date.now() - cached.timestamp < this.CACHE_TTL) {
      this.logger.info(`📦 Using cached attributes for category ${categoryId}`);
      return cached.data;
    }

    this.logger.info(`🔄 Fetching attributes for category ${categoryId}`);

    try {
      const endpoint = `/product/202309/categories/${categoryId}/attributes`;
      const result = await this.apiClient.request(endpoint, "GET", {});

      if (!result?.data?.attributes) {
        this.logger.warn("No attributes found in response");
        return [];
      }

      const attributes: CategoryAttribute[] = result.data.attributes.map(
        (attr: any) => ({
          id: attr.id,
          name: attr.name,
          type: attr.input_type || "TEXT",
          isRequired: attr.is_required === true,
          values:
            attr.values?.map((v: any) => ({ id: v.id, name: v.name })) || [],
        }),
      );

      this.attributeCache.set(cacheKey, {
        data: attributes,
        timestamp: Date.now(),
      });
      this.logger.info(
        `✅ Fetched ${attributes.length} attributes for category ${categoryId}`,
      );

      return attributes;
    } catch (error: any) {
      this.logger.error(
        `❌ Error fetching category attributes: ${error.message}`,
      );
      throw error;
    }
  }

  /**
   * Search categories by keyword
   * For SEA (ID, SG, MY, VN, TH, PH) and US shops, use category_version=v2
   */
  async searchCategories(keyword: string): Promise<Category[]> {
    this.logger.info(`🔍 Searching categories for "${keyword}"`);

    try {
      const endpoint = "/product/202309/categories";
      const result = await this.apiClient.request(endpoint, "GET", {
        keyword,
        category_version: "v2", // V2 required for SEA (ID, SG, MY, etc.) and US shops
      });

      if (!result?.data?.categories) {
        return [];
      }

      const categories: Category[] = result.data.categories
        .filter((cat: any) => cat.is_leaf === true)
        .map((cat: any) => ({
          id: cat.id,
          name: cat.local_name || cat.name,
          parentId: cat.parent_id,
          isLeaf: true,
        }));

      this.logger.info(`✅ Found ${categories.length} matching categories`);
      return categories;
    } catch (error: any) {
      this.logger.error(`❌ Error searching categories: ${error.message}`);
      throw error;
    }
  }

  /**
   * Clear category cache
   */
  clearCache(): void {
    this.categoryCache.clear();
    this.attributeCache.clear();
    this.logger.info("🧹 Category cache cleared");
  }

  /**
   * Recommend category based on product title and images
   * API: POST /product/202309/categories/recommend
   * Requires at least one image URI from TikTok
   * For SEA (ID, SG, MY, VN, TH, PH) and US shops, use category_version=v2
   */
  async recommendCategory(
    title: string,
    imageUris: string[],
    categoryVersion: "v1" | "v2" = "v2", // Default V2 for SEA/US shops
  ): Promise<Category | null> {
    if (!title || imageUris.length === 0) {
      this.logger.warn("Cannot recommend category without title and images");
      return null;
    }

    this.logger.info(`🔮 Getting category recommendation for "${title}"`);

    try {
      const endpoint = "/product/202309/categories/recommend";
      const result = await this.apiClient.request(
        endpoint,
        "POST",
        {},
        {
          product_title: title,
          images: imageUris.map((uri) => ({ uri })),
          category_version: categoryVersion,
        },
      );

      if (result?.data?.leaf_category_id) {
        const categoryId = result.data.leaf_category_id;

        // Fetch category details to get the name
        const categories = await this.getCategoryTree();
        const findCategory = (
          cats: Category[],
          id: string,
        ): Category | undefined => {
          for (const cat of cats) {
            if (cat.id === id) return cat;
            if (cat.children) {
              const found = findCategory(cat.children, id);
              if (found) return found;
            }
          }
          return undefined;
        };

        const category = findCategory(categories, categoryId);
        if (category) {
          this.logger.info(
            `✅ Recommended category: ${category.name} (${categoryId})`,
          );
          return category;
        }

        // Return basic category if not found in tree
        return {
          id: categoryId,
          name:
            result.data.category_chain?.join(" > ") || `Category ${categoryId}`,
          isLeaf: true,
        };
      }

      this.logger.warn("No category recommendation received");
      return null;
    } catch (error: any) {
      this.logger.error(
        `❌ Error getting category recommendation: ${error.message}`,
      );
      return null;
    }
  }

  /**
   * Get popular/frequently used categories
   * Returns cached root-level leaf categories
   */
  async getPopularCategories(limit = 10): Promise<Category[]> {
    try {
      const categories = await this.getCategoryTree();
      // Return leaf categories from the first levels
      return categories.filter((cat) => cat.isLeaf).slice(0, limit);
    } catch {
      return [];
    }
  }
}
