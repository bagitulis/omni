/**
 * test-accounts-helpers.js - Shared utilities for account testing
 *
 * Profile merge logic (JS equivalent of ai_profiles.py) and
 * error classification for opencode test runs.
 */

/**
 * Transform model names for plugin mode
 * google/claude-* -> google/antigravity-claude-*
 * google/gemini-* -> google/antigravity-gemini-*
 *
 * Single source of truth: same rules as ai_profiles.transform_for_plugin()
 */
function transformForPlugin(content) {
  return content
    .replace(/google\/claude-/g, "google/antigravity-claude-")
    .replace(/google\/gemini-/g, "google/antigravity-gemini-");
}

/**
 * Deep-merge shared config with profile overrides.
 * JS equivalent of ai_profiles.merge_profile() — same merge logic.
 */
function mergeProfile(shared, profile) {
  const result = {};

  for (const key of ["$schema", "google_auth", "browser_automation_engine"]) {
    if (key in shared) result[key] = JSON.parse(JSON.stringify(shared[key]));
  }

  for (const key of ["default_model", "variant"]) {
    if (key in profile) result[key] = profile[key];
  }

  result.agents = mergeSection(shared.agents || {}, profile.agents || {});
  result.categories = mergeSection(
    shared.categories || {},
    profile.categories || {},
  );

  return result;
}

/**
 * Merge a shared section with profile overrides.
 * JS equivalent of ai_profiles._merge_section().
 */
function mergeSection(sharedSection, profileSection) {
  const allKeys = new Set([
    ...Object.keys(sharedSection),
    ...Object.keys(profileSection),
  ]);
  const merged = {};

  for (const key of allKeys) {
    const entry = JSON.parse(JSON.stringify(sharedSection[key] || {}));
    Object.assign(entry, profileSection[key] || {});
    merged[key] = entry;
  }

  return merged;
}

// Error classification rules: [patterns, status, message]
const ERROR_RULES = [
  [
    ["invalid_grant", "Token has been expired", "Token has been revoked"],
    "TOKEN_EXPIRED",
    "Refresh token expired or revoked",
  ],
  [
    ["quota", "rate limit", "429", "RESOURCE_EXHAUSTED"],
    "RATE_LIMITED",
    "Quota exhausted",
  ],
  [
    ["unauthorized", "401", "UNAUTHENTICATED"],
    "UNAUTHORIZED",
    "Authentication failed",
  ],
  [
    ["permission", "403", "PERMISSION_DENIED"],
    "PERMISSION_DENIED",
    "Permission denied - check project settings",
  ],
];

const STATUS_ICONS = {
  OK: "✅",
  TOKEN_EXPIRED: "🔴",
  RATE_LIMITED: "⚠️",
  UNAUTHORIZED: "🔒",
  PERMISSION_DENIED: "🚫",
  PROJECT_ERROR: "📁",
  TIMEOUT: "⏱️",
  ERROR: "❌",
  UNKNOWN: "❓",
};

/**
 * Classify error output into a status and message
 */
function classifyError(output, err) {
  for (const [patterns, status, message] of ERROR_RULES) {
    if (patterns.some((p) => output.includes(p)))
      return { status, error: message };
  }
  if (output.includes("project") && output.includes("not found")) {
    return { status: "PROJECT_ERROR", error: "GCP project not found" };
  }
  if (err.killed || err.signal === "SIGTERM") {
    return { status: "TIMEOUT", error: "Request timed out" };
  }
  return {
    status: "ERROR",
    error: (err.message || "Unknown error").substring(0, 200),
  };
}

module.exports = {
  transformForPlugin,
  mergeProfile,
  classifyError,
  STATUS_ICONS,
};
