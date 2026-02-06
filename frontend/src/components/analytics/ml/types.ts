export interface Product {
  id: string;
  sku: string;
  name: string;
  score: number;
  recommendation: string;
  last_updated: string;
}

export interface PortfolioHealth {
  total_products: number;
  high_performers: number;
  needs_attention: number;
  critical: number;
}

export const getScoreColor = (score: number): string => {
  if (score >= 80) return "#16a34a";
  if (score >= 50) return "#f59e0b";
  return "#dc2626";
};

export const getScoreTag = (score: number) => {
  if (score >= 80) return "success";
  if (score >= 50) return "warning";
  return "error";
};
