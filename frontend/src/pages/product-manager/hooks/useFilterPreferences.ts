import { useEffect, useState, useRef, useCallback } from "react";
import type { FilterPreferences } from "@/api/productManager";
import {
  getFilterPreferences,
  saveFilterPreferences,
} from "@/api/productManager";

const DEFAULT_PREFERENCES: FilterPreferences = {
  search: "",
  column_visibility: {},
  page_size: 50,
};

export function useFilterPreferences(platform: string) {
  const [preferences, setPreferences] =
    useState<FilterPreferences>(DEFAULT_PREFERENCES);
  const [isLoading, setIsLoading] = useState(true);
  const saveTimeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(
    undefined,
  );

  // Load preferences on mount
  useEffect(() => {
    let cancelled = false;

    const loadPreferences = async () => {
      try {
        setIsLoading(true);
        const data = await getFilterPreferences(platform, "product-manager");
        if (!cancelled && data) {
          setPreferences(data);
        }
      } catch (error) {
        console.error("Failed to load filter preferences:", error);
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    };

    void loadPreferences();

    return () => {
      cancelled = true;
    };
  }, [platform]);

  // Save preferences with debounce (500ms)
  const updatePreference = useCallback(
    (updates: Partial<FilterPreferences>) => {
      const newPreferences = { ...preferences, ...updates };
      setPreferences(newPreferences);

      // Clear existing timeout
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }

      // Debounce save
      saveTimeoutRef.current = setTimeout(() => {
        saveFilterPreferences({
          platform,
          page: "product-manager",
          preferences: newPreferences,
        }).catch((error) => {
          console.error("Failed to save filter preferences:", error);
        });
      }, 500);
    },
    [platform, preferences],
  );

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (saveTimeoutRef.current) {
        clearTimeout(saveTimeoutRef.current);
      }
    };
  }, []);

  return {
    preferences,
    updatePreference,
    isLoading,
  };
}
