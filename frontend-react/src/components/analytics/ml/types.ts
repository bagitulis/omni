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

export const mockProducts: Product[] = [
  {
    id: "1",
    sku: "SKU-001",
    name: "Premium Wireless Headphones",
    score: 92,
    recommendation: "Increase stock - High demand detected",
    last_updated: "2024-02-06",
  },
  {
    id: "2",
    sku: "SKU-002",
    name: "USB-C Phone Charger",
    score: 88,
    recommendation: "Stable performer - Continue current strategy",
    last_updated: "2024-02-06",
  },
  {
    id: "3",
    sku: "SKU-003",
    name: "Portable Phone Stand",
    score: 75,
    recommendation: "Monitor closely - Seasonal trend detected",
    last_updated: "2024-02-06",
  },
  {
    id: "4",
    sku: "SKU-004",
    name: "Bluetooth Speaker",
    score: 68,
    recommendation: "Consider promotional campaign",
    last_updated: "2024-02-06",
  },
  {
    id: "5",
    sku: "SKU-005",
    name: "Phone Screen Protector",
    score: 45,
    recommendation: "Review pricing strategy - High competition detected",
    last_updated: "2024-02-05",
  },
  {
    id: "6",
    sku: "SKU-006",
    name: "Laptop Cooling Pad",
    score: 38,
    recommendation: "Critical: Consider discontinuing or rebrand",
    last_updated: "2024-02-05",
  },
  {
    id: "7",
    sku: "SKU-007",
    name: "Mechanical Keyboard",
    score: 85,
    recommendation: "Strong performer - Expand marketing reach",
    last_updated: "2024-02-06",
  },
  {
    id: "8",
    sku: "SKU-008",
    name: "Monitor Arm Stand",
    score: 72,
    recommendation: "Growing interest - Ready for scaling",
    last_updated: "2024-02-06",
  },
];

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
