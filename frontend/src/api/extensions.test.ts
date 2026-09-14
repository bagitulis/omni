// Tests for the extensions API module's pure logic.
//
// The network calls are thin wrappers over apiClient, but the source summary is
// real logic and it is load-bearing: it is how an operator notices that the DOM
// fallback has taken over from the preferred API capture.

import { describe, expect, it } from "vitest";
import { summariseSources, type ScrapedProduct } from "./extensions";

/**
 * Mirror of the defensive normalisation in useScrapedProducts.
 *
 * The job id arrives from a URL query parameter, so it can be undefined; calling
 * `.length` on it would throw and take the page down.
 */
function normaliseJobId(jobId: string | undefined | null): string {
  return typeof jobId === "string" ? jobId.trim() : "";
}

describe("job id normalisation", () => {
  it("tolerates undefined and null without throwing", () => {
    expect(() => normaliseJobId(undefined)).not.toThrow();
    expect(normaliseJobId(undefined)).toBe("");
    expect(normaliseJobId(null)).toBe("");
  });

  it("trims whitespace so a padded id still matches", () => {
    expect(normaliseJobId("  job-1  ")).toBe("job-1");
    expect(normaliseJobId("   ")).toBe("");
  });

  it("passes a normal id through", () => {
    expect(normaliseJobId("job-abc")).toBe("job-abc");
  });
});

function product(source: string): ScrapedProduct {
  return {
    id: 1,
    job_id: "job-1",
    platform: "shopee",
    scrape_mode: "search",
    product_name: "P",
    price: "1",
    sold: "1",
    link: "https://shopee.co.id/product/1/2",
    image_url: "",
    page_number: 1,
    source,
    created_at: "2026-09-14T12:00:00Z",
  };
}

describe("summariseSources", () => {
  it("reports an empty result set", () => {
    expect(summariseSources([])).toBe("no results");
  });

  it("reports an all-network run", () => {
    const products = [product("network"), product("network")];
    expect(summariseSources(products)).toBe("all via network API");
  });

  it("reports an all-fallback run", () => {
    // This is the case worth surfacing: the preferred capture path has stopped
    // working and the fallback is carrying the traffic.
    const products = [product("dom"), product("dom")];
    expect(summariseSources(products)).toBe("all via DOM fallback");
  });

  it("reports a mixed run with both counts", () => {
    const products = [product("network"), product("dom"), product("network")];
    expect(summariseSources(products)).toBe(
      "2 via network API, 1 via DOM fallback",
    );
  });

  it("reports unknown sources separately instead of hiding them", () => {
    // With an unrecognised source present, the summary must not claim the run
    // was entirely network: that would hide the very condition it exists to
    // reveal.
    const products = [product("network"), product("")];
    expect(summariseSources(products)).toBe(
      "1 via network API, 0 via DOM fallback, 1 unknown source",
    );
  });

  it("does not claim an all-network run when an unknown row is present", () => {
    const products = [product("network"), product("network"), product("")];
    expect(summariseSources(products)).not.toBe("all via network API");
    expect(summariseSources(products)).toContain("1 unknown source");
  });
});
