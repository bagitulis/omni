import { useState, useEffect, useCallback, useRef } from "react";

export interface DriftedSku {
  sku: string;
  platform: string;
  inventory_price: number;
  last_synced_price: number;
  last_sync_at: string;
  inventory_updated_at: string;
}

export interface PriceDriftData {
  drifted_skus: DriftedSku[];
  total_drifted: number;
  total_checked: number;
}

const CACHE_DURATION_MS = 5 * 60 * 1000; // 5 minutes

/**
 * Hook to fetch price drift data from the backend.
 * Caches results for 5 minutes to avoid excessive API calls.
 */
export function usePriceDrift() {
  const [driftData, setDriftData] = useState<PriceDriftData | null>(null);
  const [loading, setLoading] = useState(false);
  const lastFetchRef = useRef<number>(0);
  const driftDataRef = useRef<PriceDriftData | null>(null);

  const fetchDrift = useCallback(async (force = false) => {
    // Skip if cached and not forced
    if (!force && Date.now() - lastFetchRef.current < CACHE_DURATION_MS && driftDataRef.current) {
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("/api/inventory/price-drift", {
        credentials: "include",
      });
      const json = await response.json();
      if (json.success && json.data) {
        setDriftData(json.data);
        driftDataRef.current = json.data;
        lastFetchRef.current = Date.now();
      }
    } catch {
      // Silent failure - drift is non-critical
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch on mount
  useEffect(() => {
    fetchDrift();
  }, [fetchDrift]);

  return { driftData, loading, refetch: () => fetchDrift(true) };
}

/**
 * Get set of drifted SKUs for quick lookup in table rendering.
 */
export function getDriftedSkuSet(driftData: PriceDriftData | null): Set<string> {
  if (!driftData?.drifted_skus) return new Set();
  return new Set(driftData.drifted_skus.map((d) => d.sku));
}

/**
 * Get drift info for a specific SKU (for tooltip display).
 */
export function getDriftInfoForSku(
  driftData: PriceDriftData | null,
  sku: string,
): DriftedSku[] {
  if (!driftData?.drifted_skus) return [];
  return driftData.drifted_skus.filter((d) => d.sku === sku);
}
