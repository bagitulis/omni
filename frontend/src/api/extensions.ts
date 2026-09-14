// API client for the extensions feature.
//
// Types are snake_case to match the Go backend, per the frontend conventions.
// All calls go through the shared apiClient, which attaches auth and unwraps the
// {success, data} envelope.

import apiClient from "./client";

/** A paired Chrome extension install. */
export interface OmniExtension {
  extension_id: string;
  hostname: string;
  browser_info: string;
  chrome_version: string;
  extension_version: string;
  capabilities: string[];
  /** "connected" | "disconnected" */
  status: string;
  last_seen?: string;
  paired_at: string;
}

/** The response of a pairing-code request. */
export interface PairingCodeResponse {
  code: string;
  expires_at: string;
  ttl_seconds: number;
}

/** A product collected by a scrape job. */
export interface ScrapedProduct {
  id: number;
  job_id: string;
  platform: string;
  scrape_mode: string;
  query?: string;
  shop_id?: string;
  product_name: string;
  price: string;
  sold: string;
  link: string;
  image_url: string;
  shopee_item_id?: string;
  page_number: number;
  /** "network" | "dom" — which capture path produced this row. */
  source?: string;
  created_at: string;
}

export interface ScrapedProductsPage {
  products: ScrapedProduct[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface ScrapeRequest {
  mode: "search" | "shop" | "product";
  query?: string;
  shop_url?: string;
  shop_id?: string;
  product_url?: string;
  max_pages?: number;
  max_products?: number;
  extension_id: string;
}

export async function listExtensions(): Promise<OmniExtension[]> {
  const response = await apiClient.get<OmniExtension[]>("/extensions");
  if (!response.success) {
    throw new Error(response.error || "Failed to load extensions");
  }
  return response.data ?? [];
}

export async function generatePairingCode(): Promise<PairingCodeResponse> {
  const response = await apiClient.post<PairingCodeResponse>(
    "/extensions/pairing/generate",
    {},
  );
  if (!response.success || !response.data) {
    throw new Error(response.error || "Failed to generate a pairing code");
  }
  return response.data;
}

export async function unpairExtension(extensionId: string): Promise<void> {
  const response = await apiClient.delete<void>(
    `/extensions/${encodeURIComponent(extensionId)}`,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to unpair the extension");
  }
}

export async function startScrape(
  req: ScrapeRequest,
): Promise<{ summary: string }> {
  const response = await apiClient.post<{ summary: string }>(
    "/extensions/scrape",
    req,
  );
  if (!response.success) {
    throw new Error(response.error || "Failed to start the scrape");
  }
  return response.data ?? { summary: "" };
}

/**
 * Pagination envelope returned by the list endpoint.
 *
 * Declared locally because ApiResponse has no `meta` member. Widening the shared
 * client type for one endpoint would change its contract for every caller, so
 * the extra field is read here instead.
 */
interface PaginatedResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  meta?: {
    total?: number;
    page?: number;
    page_size?: number;
    total_pages?: number;
  };
}

export async function listScrapedProducts(
  jobId: string,
  page = 1,
  pageSize = 50,
): Promise<ScrapedProductsPage> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  });
  if (jobId) params.set("job_id", jobId);

  const response = (await apiClient.get<ScrapedProduct[]>(
    `/extensions/scraped-products?${params.toString()}`,
  )) as PaginatedResponse<ScrapedProduct[]>;

  if (!response.success) {
    throw new Error(response.error || "Failed to load scraped products");
  }

  const products = response.data ?? [];
  return {
    products,
    total: response.meta?.total ?? products.length,
    page: response.meta?.page ?? page,
    page_size: response.meta?.page_size ?? pageSize,
    total_pages: response.meta?.total_pages ?? (products.length > 0 ? 1 : 0),
  };
}

/**
 * Summarise which capture path produced a set of rows.
 *
 * Shown in the results header so it is obvious when the DOM fallback is carrying
 * production traffic, which would otherwise go unnoticed until the fallback
 * breaks too.
 */
export function summariseSources(products: ScrapedProduct[]): string {
  if (products.length === 0) return "no results";

  const network = products.filter((p) => p.source === "network").length;
  const dom = products.filter((p) => p.source === "dom").length;
  // Rows whose source is unrecognised are counted separately rather than folded
  // into a known path: reporting "all via network API" while some rows came from
  // somewhere unknown would hide exactly the condition this summary exists to
  // reveal.
  const other = products.length - network - dom;

  if (other === 0 && network > 0 && dom === 0) return "all via network API";
  if (other === 0 && dom > 0 && network === 0) return "all via DOM fallback";

  const parts = [`${network} via network API`, `${dom} via DOM fallback`];
  if (other > 0) parts.push(`${other} unknown source`);
  return parts.join(", ");
}
