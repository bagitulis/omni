/**
 * Lazy load component definitions for Vue Router
 * Extracted for maintainability (agents.md: max 300 lines per file)
 */

// Core views
export const Dashboard = () =>
  import(/* webpackChunkName: "dashboard" */ "../views/Dashboard.vue");
export const Inventory = () =>
  import(/* webpackChunkName: "inventory" */ "../views/Inventory.vue");
export const Settings = () =>
  import(/* webpackChunkName: "settings" */ "../views/Settings.vue");
export const RouteMapping = () =>
  import(/* webpackChunkName: "route-mapping" */ "../views/RouteMapping.vue");
export const ScriptMonitor = () =>
  import(/* webpackChunkName: "script-monitor" */ "../views/ScriptMonitor.vue");
export const LoginView = () =>
  import(/* webpackChunkName: "login" */ "../views/LoginView.vue");
export const DevPreview = () =>
  import(/* webpackChunkName: "dev-preview" */ "../views/DevPreview.vue");

// Analytics views - Report (Escrow/Price/Shipping Fee Analysis)
export const ShopeeAnalytics = () =>
  import(
    /* webpackChunkName: "shopee-analytics" */ "../views/analytics/ShopeeAnalytics.vue"
  );
export const TiktokAnalytics = () =>
  import(
    /* webpackChunkName: "tiktok-analytics" */ "../views/analytics/TiktokAnalytics.vue"
  );

// Analytics views - Ads Performance
export const ShopeeAdsAnalytics = () =>
  import(
    /* webpackChunkName: "shopee-ads-analytics" */ "../views/analytics/ShopeeAdsAnalytics.vue"
  );
export const TiktokAdsAnalytics = () =>
  import(
    /* webpackChunkName: "tiktok-ads-analytics" */ "../views/analytics/TiktokAdsAnalytics.vue"
  );
export const AIReportGallery = () =>
  import(
    /* webpackChunkName: "ai-report-gallery" */ "../views/analytics/AIReportGallery.vue"
  );
