/**
 * Lazada Category Service
 * Handles fetching and caching Lazada categories for product creation
 * Max 300 lines - AGENTS.MD compliant
 */

import { LazadaAPIClient } from "../api/clients/lazadaAPIClient";

interface LazadaCategory {
  category_id: number;
  name: string;
  var: boolean; // has variations
  leaf: boolean;
  children?: LazadaCategory[];
}

interface CategoryAttribute {
  name: string;
  input_type: string;
  is_mandatory: boolean;
  attribute_type: string;
  label: string;
  options?: Array<{ name: string }>;
}

export class LazadaCategoryService {
  private apiClient: LazadaAPIClient;
  private categoryCache: LazadaCategory[] | null = null;
  private attributeCache: Map<number, CategoryAttribute[]> = new Map();

  constructor(apiClient: LazadaAPIClient) {
    this.apiClient = apiClient;
  }

  /**
   * Get category tree from Lazada
   * GET /category/tree/get
   */
  async getCategoryTree(): Promise<LazadaCategory[]> {
    if (this.categoryCache) {
      return this.categoryCache;
    }

    const response = await this.apiClient.request(
      "/category/tree/get",
      "GET"
    );

    if (response?.data) {
      this.categoryCache = response.data;
      return response.data;
    }

    throw new Error(response?.message || "Failed to fetch categories");
  }

  /**
   * Get flattened category list for searching
   */
  async getFlatCategories(parentId?: number): Promise<Array<{
    id: string;
    name: string;
    parentId?: string;
    isLeaf: boolean;
  }>> {
    const tree = await this.getCategoryTree();
    const result: Array<{ id: string; name: string; parentId?: string; isLeaf: boolean }> = [];

    const flatten = (categories: LazadaCategory[], parentCategory?: number) => {
      for (const cat of categories) {
        if (!parentId || parentCategory === parentId || !parentCategory) {
          result.push({
            id: String(cat.category_id),
            name: cat.name,
            parentId: parentCategory ? String(parentCategory) : undefined,
            isLeaf: cat.leaf,
          });
        }
        if (cat.children && cat.children.length > 0) {
          flatten(cat.children, cat.category_id);
        }
      }
    };

    flatten(tree);
    return result;
  }

  /**
   * Search categories by keyword
   */
  async searchCategories(keyword: string): Promise<Array<{
    id: string;
    name: string;
    isLeaf: boolean;
  }>> {
    const allCategories = await this.getFlatCategories();
    const lowercaseKeyword = keyword.toLowerCase();

    return allCategories
      .filter(cat => cat.name.toLowerCase().includes(lowercaseKeyword))
      .slice(0, 50)
      .map(cat => ({
        id: cat.id,
        name: cat.name,
        isLeaf: cat.isLeaf,
      }));
  }

  /**
   * Get category attributes
   * GET /category/attributes/get
   */
  async getCategoryAttributes(categoryId: number): Promise<CategoryAttribute[]> {
    if (this.attributeCache.has(categoryId)) {
      return this.attributeCache.get(categoryId)!;
    }

    const response = await this.apiClient.request(
      "/category/attributes/get",
      "GET",
      { primary_category_id: categoryId }
    );

    if (response?.data) {
      const attributes: CategoryAttribute[] = response.data;
      this.attributeCache.set(categoryId, attributes);
      return attributes;
    }

    return [];
  }

  /**
   * Clear caches
   */
  clearCache(): void {
    this.categoryCache = null;
    this.attributeCache.clear();
  }

  /**
   * Recommend category based on product title
   * API: GET /category/suggestion/get
   * Lazada is the simplest - only requires product title
   */
  async recommendCategory(
    productTitle: string
  ): Promise<{ id: string; name: string; isLeaf: boolean } | null> {
    if (!productTitle || productTitle.length < 3) {
      return null;
    }

    try {
      const response = await this.apiClient.request(
        "/category/suggestion/get",
        "GET",
        { product_name: productTitle }
      );

      if (response?.data?.categorySuggestions?.length > 0) {
        const suggestion = response.data.categorySuggestions[0];
        const categoryId = suggestion.category_id || suggestion.categoryId;
        const categoryName = suggestion.category_name || suggestion.categoryPath;

        return {
          id: String(categoryId),
          name: categoryName || `Category ${categoryId}`,
          isLeaf: true,
        };
      }

      return null;
    } catch {
      return null;
    }
  }

  /**
   * Get popular categories (leaf categories from tree)
   */
  async getPopularCategories(limit = 10): Promise<Array<{
    id: string;
    name: string;
    isLeaf: boolean;
  }>> {
    try {
      const allCategories = await this.getFlatCategories();
      return allCategories
        .filter(cat => cat.isLeaf)
        .slice(0, limit);
    } catch {
      return [];
    }
  }
}
