interface RouteConfig {
  id: string;
  route_path: string;
  method: string;
  cache_ttl: number;
  description: string;
}

export const API_ROUTES: RouteConfig[] = [
  {
    id: "1",
    route_path: "/api/orders",
    method: "GET",
    cache_ttl: 300,
    description: "Fetch order list",
  },
  {
    id: "2",
    route_path: "/api/orders",
    method: "POST",
    cache_ttl: 0,
    description: "Create order",
  },
  {
    id: "3",
    route_path: "/api/products",
    method: "GET",
    cache_ttl: 600,
    description: "Fetch products",
  },
  {
    id: "4",
    route_path: "/api/products/:id",
    method: "GET",
    cache_ttl: 300,
    description: "Product detail",
  },
  {
    id: "5",
    route_path: "/api/products",
    method: "PUT",
    cache_ttl: 0,
    description: "Update product",
  },
  {
    id: "6",
    route_path: "/api/inventory",
    method: "GET",
    cache_ttl: 120,
    description: "Fetch inventory",
  },
];

export type { RouteConfig };
