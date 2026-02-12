import type { MetricType } from "web-vitals";
import { logger } from "@/lib/logger";

type MetricCallback = (metric: MetricType) => void;

const defaultCallback: MetricCallback = (metric) => {
  if (import.meta.env.DEV) {
    logger.info(
      `[WebVitals] ${metric.name}: ${metric.value.toFixed(2)} (${metric.rating})`,
    );
  }
};

export function initPerformanceMonitoring(
  callback: MetricCallback = defaultCallback,
): void {
  // Dynamic import to avoid blocking app render
  import("web-vitals").then(({ onCLS, onINP, onFCP, onLCP, onTTFB }) => {
    onCLS(callback);
    onINP(callback);
    onFCP(callback);
    onLCP(callback);
    onTTFB(callback);
  });
}
