import { useState, useEffect, useMemo } from "react";

interface UseVirtualScrollOptions {
  /**
   * Explicit container height. If not provided, will calculate from viewport.
   */
  containerHeight?: number;
  /**
   * Row height for virtual scrolling calculation.
   * Default: 54 (Standard AntD table row height)
   */
  rowHeight?: number;
  /**
   * Whether virtual scrolling is enabled.
   * Default: true
   */
  enabled?: boolean;
  /**
   * Space to subtract from viewport height if calculating automatically.
   * Include header (48px), padding, page headers, filters, pagination (approx 64px), etc.
   * Default: 300 (Safety margin for headers + filters + padding)
   */
  offsetBottom?: number;
}

interface UseVirtualScrollResult {
  scrollConfig: { y: number; x?: number | string | true };
  virtual: boolean;
  tableProps: {
    virtual: boolean;
    scroll: { y: number; x?: number | string | true };
  };
}

/**
 * Hook to configure virtual scrolling for Ant Design Tables.
 * Calculates available height and returns props to enable virtual mode.
 */
export function useVirtualScroll({
  containerHeight,
  enabled = true,
  offsetBottom = 300,
}: UseVirtualScrollOptions = {}): UseVirtualScrollResult {
  const [scrollY, setScrollY] = useState<number>(500);

  useEffect(() => {
    if (!enabled && !containerHeight) return;

    const calculateHeight = () => {
      if (containerHeight) {
        setScrollY(containerHeight);
        return;
      }

      // Calculate based on viewport
      const windowHeight = window.innerHeight;
      // Subtract offset (headers, padding, etc.)
      // Ensure minimum height of 400px
      const calculatedHeight = Math.max(windowHeight - offsetBottom, 400);
      setScrollY(calculatedHeight);
    };

    calculateHeight();

    const handleResize = () => {
      calculateHeight();
    };

    window.addEventListener("resize", handleResize);
    return () => window.removeEventListener("resize", handleResize);
  }, [containerHeight, enabled, offsetBottom]);

  const result = useMemo(() => {
    if (!enabled) {
      return {
        scrollConfig: { y: scrollY },
        virtual: false,
        tableProps: {
          virtual: false,
          scroll: { y: scrollY },
        },
      };
    }

    return {
      scrollConfig: { y: scrollY },
      virtual: true,
      tableProps: {
        virtual: true,
        scroll: { y: scrollY },
      },
    };
  }, [enabled, scrollY]);

  return result;
}
