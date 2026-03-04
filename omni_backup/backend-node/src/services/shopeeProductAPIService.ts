import { ShopeeAPIClient } from "../api/clients/shopeeAPIClient";
import { Logger } from "winston";
import { getLogger } from "../utils/logger";

export class ShopeeProductAPIService {
  private apiClient: ShopeeAPIClient;
  private logger: Logger;

  constructor(apiClient: ShopeeAPIClient) {
    this.apiClient = apiClient;
    this.logger = getLogger("ShopeeProductAPIService");
  }

  /**
   * Get item list from Shopee API (single page)
   * Returns items + pagination info (has_next_page, next_offset, total_count)
   */
  async getItemList(
    itemStatus: string,
    offset: number,
    limit: number
  ): Promise<any> {
    const response = await this.apiClient.request(
      "/api/v2/product/get_item_list",
      "GET",
      { item_status: itemStatus, offset, page_size: Math.min(limit, 100) }
    );

    if (!response || response.error) {
      const errorMsg = response?.message || response?.error || "Unknown error";
      throw new Error(`Shopee API Error: ${errorMsg}`);
    }

    return {
      items: response?.response?.item || [],
      totalCount: response?.response?.total_count || 0,
      hasNextPage: response?.response?.has_next_page || false,
      nextOffset: response?.response?.next_offset || 0,
    };
  }

  /**
   * Get ALL items from Shopee API with automatic pagination
   * Loops through all pages until has_next_page is false
   */
  async getAllItems(itemStatus: string, maxItems?: number): Promise<any> {
    this.logger.info(`🔄 Fetching ALL Shopee items | Status: ${itemStatus}${maxItems ? `, Max: ${maxItems}` : ''}`);

    const allItems: any[] = [];
    let offset = 0;
    const pageSize = 100; // Max allowed by Shopee API
    let pageCount = 0;
    let totalCount = 0;

    while (true) {
      this.logger.info(`   📄 Fetching page ${pageCount + 1} | Offset: ${offset}`);

      const result = await this.getItemList(itemStatus, offset, pageSize);

      if (!result.items || result.items.length === 0) {
        this.logger.info(`   ✅ No more items found`);
        break;
      }

      allItems.push(...result.items);
      totalCount = result.totalCount;
      pageCount++;

      this.logger.info(`   📦 Page ${pageCount}: ${result.items.length} items (total so far: ${allItems.length}/${totalCount})`);

      // Check if we've reached max items limit
      if (maxItems && allItems.length >= maxItems) {
        this.logger.info(`   ✅ Reached max items limit: ${maxItems}`);
        allItems.splice(maxItems); // Trim to max
        break;
      }

      // Check if there's more pages
      if (!result.hasNextPage) {
        this.logger.info(`   ✅ Reached last page`);
        break;
      }

      // Move to next page
      offset = result.nextOffset;
    }

    this.logger.info(`✅ Fetched ALL items | Pages: ${pageCount}, Total: ${allItems.length}`);

    return {
      items: allItems,
      totalCount: totalCount,
      pageCount,
    };
  }

  async getBaseInfo(itemIdList: number[]): Promise<any> {
    const response = await this.apiClient.request(
      "/api/v2/product/get_item_base_info",
      "GET",
      { item_id_list: itemIdList.join(",") }
    );

    if (!response || response.error) {
      this.logger.warn(
        `Failed to get base info: ${response?.message || "Unknown error"}`
      );
    }

    return response?.response?.item_list || [];
  }

  async getExtraInfo(itemIdList: number[]): Promise<any> {
    const response = await this.apiClient.request(
      "/api/v2/product/get_item_extra_info",
      "GET",
      { item_id_list: itemIdList.join(",") }
    );

    if (!response || response.error) {
      this.logger.warn(
        `Failed to get extra info: ${response?.message || "Unknown error"}`
      );
    }

    return response?.response?.item_list || [];
  }

  async getModelList(itemId: number): Promise<any> {
    const response = await this.apiClient.request(
      "/api/v2/product/get_model_list",
      "GET",
      { item_id: itemId, page_size: 100 }
    );

    if (!response || response.error) {
      this.logger.warn(
        `Failed to get model list for item ${itemId}: ${response?.message || "Unknown error"}`
      );
      return [];
    }

    return response?.response?.model || [];
  }

  async getModelBaseInfo(itemId: number, modelIdList: number[]): Promise<any> {
    const response = await this.apiClient.request(
      "/api/v2/product/get_model_base_info",
      "GET",
      { item_id: itemId, model_id_list: modelIdList.join(",") }
    );

    if (!response || response.error) {
      this.logger.warn(
        `Failed to get model base info: ${response?.message || "Unknown error"}`
      );
    }

    return response?.response?.model_list || [];
  }

  /**
   * Fetch and merge all product data (list + base + extra + models)
   */
  async fetchAndMergeProducts(
    itemStatus: string,
    offset: number,
    limit: number
  ): Promise<any[]> {
    try {
      // Step 1: Get item list
      const { items } = await this.getItemList(itemStatus, offset, limit);

      if (items.length === 0) {
        return [];
      }

      const itemIdList = items.map((item: any) => item.item_id);

      // Step 2: Get base info
      const baseItems = await this.getBaseInfo(itemIdList);

      // Step 3: Get extra info
      const extraItems = await this.getExtraInfo(itemIdList);

      // Step 4: Get model list for items with models
      const productsWithModels = baseItems.filter(
        (item: any) => item.has_model
      );
      const modelDataMap: Record<number, any> = {};

      for (const item of productsWithModels) {
        try {
          const modelResponse = await this.apiClient.request(
            "/api/v2/product/get_model_list",
            "GET",
            {
              item_id: item.item_id,
            }
          );
          if (modelResponse.response) {
            modelDataMap[item.item_id] = modelResponse.response;
          }
        } catch (err: any) {
          this.logger.warn(
            `Failed to fetch models for item ${item.item_id}: ${err.message}`
          );
        }
      }

      // Merge all data
      return baseItems.map((baseItem: any) => {
        const extraItem =
          extraItems.find((e: any) => e.item_id === baseItem.item_id) || {};
        const modelData = modelDataMap[baseItem.item_id] || {};

        return {
          ...baseItem,
          ...extraItem,
          model_list: modelData.model || [],
          tier_variation: modelData.tier_variation || [],
        };
      });
    } catch (error: any) {
      this.logger.error(`Failed to fetch and merge products: ${error.message}`);
      throw error;
    }
  }

  /**
   * Fetch ALL products with pagination and merge all data
   * This handles the case where total products > 100 (Shopee's max page_size)
   */
  async fetchAllProducts(itemStatus: string, maxItems?: number): Promise<{
    products: any[];
    totalCount: number;
    pageCount: number;
  }> {
    try {
      this.logger.info(`🔄 Fetching ALL Shopee products with pagination | Status: ${itemStatus}`);

      // Step 1: Get ALL item IDs with pagination
      const { items, totalCount, pageCount } = await this.getAllItems(itemStatus, maxItems);

      if (items.length === 0) {
        return { products: [], totalCount: 0, pageCount: 0 };
      }

      this.logger.info(`📦 Total items to process: ${items.length}`);

      // Step 2: Process items in batches of 50 (Shopee API limit for get_item_base_info)
      const batchSize = 50;
      const allMergedProducts: any[] = [];

      for (let i = 0; i < items.length; i += batchSize) {
        const batch = items.slice(i, i + batchSize);
        const itemIdList = batch.map((item: any) => item.item_id);

        this.logger.info(`   🔄 Processing batch ${Math.floor(i / batchSize) + 1}/${Math.ceil(items.length / batchSize)} (${itemIdList.length} items)`);

        // Get base info for batch
        const baseItems = await this.getBaseInfo(itemIdList);

        // Get extra info for batch
        const extraItems = await this.getExtraInfo(itemIdList);

        // Get model list for items with models
        const productsWithModels = baseItems.filter((item: any) => item.has_model);
        const modelDataMap: Record<number, any> = {};

        for (const item of productsWithModels) {
          try {
            const modelResponse = await this.apiClient.request(
              "/api/v2/product/get_model_list",
              "GET",
              { item_id: item.item_id }
            );
            if (modelResponse.response) {
              modelDataMap[item.item_id] = modelResponse.response;
            }
          } catch (err: any) {
            this.logger.warn(`Failed to fetch models for item ${item.item_id}: ${err.message}`);
          }
        }

        // Merge batch data
        const mergedBatch = baseItems.map((baseItem: any) => {
          const extraItem = extraItems.find((e: any) => e.item_id === baseItem.item_id) || {};
          const modelData = modelDataMap[baseItem.item_id] || {};

          return {
            ...baseItem,
            ...extraItem,
            model_list: modelData.model || [],
            tier_variation: modelData.tier_variation || [],
          };
        });

        allMergedProducts.push(...mergedBatch);
      }

      this.logger.info(`✅ Fetched and merged ALL products | Total: ${allMergedProducts.length}`);

      return {
        products: allMergedProducts,
        totalCount,
        pageCount,
      };
    } catch (error: any) {
      this.logger.error(`Failed to fetch all products: ${error.message}`);
      throw error;
    }
  }
}
