/**
 * API Headers Utility
 * Centralized function to get authentication and tenant headers
 * Used by composables that make direct fetch calls
 */

/**
 * Get CSRF token from cookies
 */
function getCSRFToken(): string | null {
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : null;
}

export function getAuthHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "Cache-Control": "no-cache",
  };

  // Add JWT token
  const token = localStorage.getItem("authToken");
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  // Add tenant ID - CRITICAL for multi-tenant isolation
  const tenantId = localStorage.getItem("tenantId");
  if (tenantId) {
    headers["x-tenant-id"] = tenantId;
  }

  // Add CSRF token for mutating requests
  const csrfToken = getCSRFToken();
  if (csrfToken) {
    headers["x-csrf-token"] = csrfToken;
  }

  return headers;
}

/**
 * Get API base URL dynamically
 */
export function getApiBaseUrl(path: string = ""): string {
  const host = window.location.hostname;
  const protocol = window.location.protocol;

  // In production (non-localhost), use same origin without port
  // In development (localhost), use port 3000 for backend
  const isLocalhost = host === "localhost" || host === "127.0.0.1";
  const baseUrl = isLocalhost
    ? `${protocol}//${host}:3000/api${path}`
    : `${protocol}//${host}/api${path}`;

  return baseUrl;
}

export default { getAuthHeaders, getApiBaseUrl };
