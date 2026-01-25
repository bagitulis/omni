/**
 * Shopee Category Service
 * Handles fetching and caching Shopee categories for product creation
 * Max 300 lines - AGENTS.MD compliant
 */

import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";

interface ShopeeCategory {
  category_id: number;
  parent_category_id: number;
  original_category_name: string;
  display_category_name: string;
  has_children: boolean;
}

interface CategoryAttribute {
  attribute_id: number;
  original_attribute_name: string;
  display_attribute_name: string;
  is_mandatory: boolean;
  input_validation_type: string;
  format_type: string;
  date_format_type: string;
  input_type: string;
  attribute_unit: string[];
  attribute_value_list: Array<{
    value_id: number;
    original_value_name: string;
    display_value_name: string;
    value_unit: string;
    parent_attribute_list: any[];
    parent_brand_list: any[];
  }>;
}

export class ShopeeCategoryService {
  private apiClient: ShopeeAPIClient;
  private categoryCache: Map<string, ShopeeCategory[]> = new Map();
  private attributeCache: Map<number, CategoryAttribute[]> = new Map();

  constructor(apiClient: ShopeeAPIClient) {
    this.apiClient = apiClient;
  }

  /**
   * Get category list from Shopee
   * GET /api/v2/product/get_category
   */
  async getCategoryList(): Promise<ShopeeCategory[]> {
    const cacheKey = "all";
    if (this.categoryCache.has(cacheKey)) {
      return this.categoryCache.get(cacheKey)!;
    }

    const response = await this.apiClient.request(
      "/api/v2/product/get_category",
      "GET",
      { language: "id" }
    );

    if (response?.response?.category_list) {
      const categories: ShopeeCategory[] = response.response.category_list.map(
        (cat: any) => ({
          category_id: cat.category_id,
          parent_category_id: cat.parent_category_id,
          original_category_name: cat.original_category_name,
          display_category_name: cat.display_category_name,
          has_children: cat.has_children,
        })
      );
      this.categoryCache.set(cacheKey, categories);
      return categories;
    }

    throw new Error(response?.message || "Failed to fetch categories");
  }

  /**
   * Get category tree (hierarchical)
   */
  async getCategoryTree(parentId?: number): Promise<Array<{
    id: string;
    name: string;
    parentId?: string;
    isLeaf: boolean;
  }>> {
    const allCategories = await this.getCategoryList();
    const targetParent = parentId ?? 0;

    return allCategories
      .filter(cat => cat.parent_category_id === targetParent)
      .map(cat => ({
        id: String(cat.category_id),
        name: cat.display_category_name,
        parentId: cat.parent_category_id ? String(cat.parent_category_id) : undefined,
        isLeaf: !cat.has_children,
      }));
  }

  /**
   * Search categories by keyword
   */
  async searchCategories(keyword: string): Promise<Array<{
    id: string;
    name: string;
    isLeaf: boolean;
  }>> {
    const allCategories = await this.getCategoryList();
    const lowercaseKeyword = keyword.toLowerCase();

    return allCategories
      .filter(cat => 
        cat.display_category_name.toLowerCase().includes(lowercaseKeyword) ||
        cat.original_category_name.toLowerCase().includes(lowercaseKeyword)
      )
      .slice(0, 50)
      .map(cat => ({
        id: String(cat.category_id),
        name: cat.display_category_name,
        isLeaf: !cat.has_children,
      }));
  }

  /**
   * Get category attributes
   * GET /api/v2/product/get_attributes
   */
  async getCategoryAttributes(categoryId: number): Promise<CategoryAttribute[]> {
    if (this.attributeCache.has(categoryId)) {
      return this.attributeCache.get(categoryId)!;
    }

    const response = await this.apiClient.request(
      "/api/v2/product/get_attributes",
      "GET",
      { category_id: categoryId, language: "id" }
    );

    if (response?.response?.attribute_list) {
      const attributes: CategoryAttribute[] = response.response.attribute_list;
      this.attributeCache.set(categoryId, attributes);
      return attributes;
    }

    return [];
  }

  /**
   * Clear caches
   */
  clearCache(): void {
    this.categoryCache.clear();
    this.attributeCache.clear();
  }

  /**
   * Recommend category based on product name and optional image
   * API: GET /api/v2/product/category_recommend
   */
  async recommendCategory(
    itemName: string,
    coverImageId?: string
  ): Promise<{ id: string; name: string; isLeaf: boolean } | null> {
    if (!itemName) {
      return null;
    }

    try {
      const params: Record<string, any> = { item_name: itemName };
      if (coverImageId) {
        params.cover_image = coverImageId;
      }

      const response = await this.apiClient.request(
        "/api/v2/product/category_recommend",
        "GET",
        params
      );

      if (response?.response?.category_id?.length > 0) {
        const categoryIds = response.response.category_id;
        const leafCategoryId = categoryIds[categoryIds.length - 1];

        // Find category name from cache
        const allCategories = await this.getCategoryList();
        const found = allCategories.find(c => c.category_id === leafCategoryId);

        return {
          id: String(leafCategoryId),
          name: found?.display_category_name || `Category ${leafCategoryId}`,
          isLeaf: true,
        };
      }

      return null;
    } catch {
      return null;
    }
  }

  /**
   * Get popular categories (top-level leaf categories)
   */
  async getPopularCategories(limit = 10): Promise<Array<{
    id: string;
    name: string;
    isLeaf: boolean;
  }>> {
    try {
      const allCategories = await this.getCategoryList();
      return allCategories
        .filter(cat => !cat.has_children)
        .slice(0, limit)
        .map(cat => ({
          id: String(cat.category_id),
          name: cat.display_category_name,
          isLeaf: true,
        }));
    } catch {
      return [];
    }
  }
}
