import cacheService from "@/services/cacheService";

export type SortDirection = "asc" | "desc" | null;
export interface SortState {
  column: string | null;
  direction: SortDirection;
}

const STORAGE_KEY = "inventory_sort_state";
const CACHE_TTL = 24 * 60 * 60 * 1000;

const defaultState: SortState = { column: null, direction: null };

export function loadSortState(): SortState {
  try {
    const stored = cacheService.get(STORAGE_KEY, { ttl: CACHE_TTL });
    if (stored && typeof stored === "object") {
      return {
        column: stored.column || null,
        direction: stored.direction ?? null,
      };
    }
  } catch (error) {
    console.warn("Failed to load sort state", error);
  }
  return { ...defaultState };
}

export function saveSortState(state: SortState) {
  try {
    cacheService.set(STORAGE_KEY, state, { ttl: CACHE_TTL });
  } catch (error) {
    console.warn("Failed to save sort state", error);
  }
}

export function clearSortState() {
  saveSortState({ ...defaultState });
}
