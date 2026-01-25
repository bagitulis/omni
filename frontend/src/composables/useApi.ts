/**
 * Composable for API operations
 * Provides reactive access to the API service
 */

import api from "@/services/api";

export function useApi() {
  return api;
}
