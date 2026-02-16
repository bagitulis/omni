import { useEffect, useMemo, useState } from "react";
import type { InventoryFilterPreferences } from "@/api/inventoryFilterPreferences";
import {
  buildInventoryFilterPreferencesSignature,
  normalizeInventoryFilterPreferencesForSync,
} from "../utils/inventoryFilterPreferenceSync";

interface UseInventoryFilterPreferenceSyncParams {
  preferencesLoaded: boolean;
  visibleColumns: string[];
  lockedColumns: string[];
  columnFilters: Record<string, string>;
  searchQuery: string;
  saveFilterPreferences: (
    preferences: InventoryFilterPreferences,
    options?: { onSuccess?: () => void },
  ) => void;
  debounceMs?: number;
}

export function useInventoryFilterPreferenceSync({
  preferencesLoaded,
  visibleColumns,
  lockedColumns,
  columnFilters,
  searchQuery,
  saveFilterPreferences,
  debounceMs = 300,
}: UseInventoryFilterPreferenceSyncParams) {
  const [isInitialized, setIsInitialized] = useState(false);
  const [lastSavedSignature, setLastSavedSignature] = useState<string | null>(
    null,
  );

  const payload = useMemo(
    () =>
      normalizeInventoryFilterPreferencesForSync({
        visible_columns: visibleColumns,
        locked_columns: lockedColumns,
        column_filters: columnFilters,
        search_query: searchQuery,
      }),
    [columnFilters, lockedColumns, searchQuery, visibleColumns],
  );

  const signature = useMemo(
    () => buildInventoryFilterPreferencesSignature(payload),
    [payload],
  );

  useEffect(() => {
    if (!preferencesLoaded || isInitialized) {
      return;
    }

    setLastSavedSignature(signature);
    setIsInitialized(true);
  }, [isInitialized, preferencesLoaded, signature]);

  useEffect(() => {
    if (!preferencesLoaded || !isInitialized) {
      return;
    }

    if (lastSavedSignature === signature) {
      return;
    }

    const syncTimer = window.setTimeout(() => {
      const requestSignature = signature;
      saveFilterPreferences(payload, {
        onSuccess: () => {
          setLastSavedSignature(requestSignature);
        },
      });
    }, debounceMs);

    return () => {
      window.clearTimeout(syncTimer);
    };
  }, [
    debounceMs,
    isInitialized,
    lastSavedSignature,
    payload,
    preferencesLoaded,
    saveFilterPreferences,
    signature,
  ]);
}
