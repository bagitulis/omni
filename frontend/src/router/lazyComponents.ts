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

// Analytics views
export const AdsDashboard = () => import("../views/Analytics/AdsDashboard.vue");

// Redirect legacy analytics routes to the new AdsDashboard
export const ShopeeAnalytics = AdsDashboard;
export const TiktokAnalytics = AdsDashboard;
export const TiktokAdsAnalytics = AdsDashboard;
export const ShopeeAdsAnalytics = AdsDashboard;
