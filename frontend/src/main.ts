import { createApp } from "vue";
import { createPinia } from "pinia";
import VueApexCharts from "vue3-apexcharts";

import App from "./App.vue";
import router from "./router";

// Import Tailwind styles
import "./styles/tailwind.css";
import "./styles/theme-unified.css";
import "./styles/icons.css";
import "./styles/compact-scale.css";

const app = createApp(App);

app.use(createPinia());
app.use(router);
app.use(VueApexCharts);

// Mount app
app.mount("#app");

// Lazy load web-vitals after critical rendering and ServiceWorker
if (import.meta.env.PROD) {
  if ("serviceWorker" in navigator) {
    window.addEventListener("load", () => {
      navigator.serviceWorker.register("/sw.js").catch(() => {});
    });
  }

  // Lazy load web-vitals for performance monitoring
  window.addEventListener(
    "load",
    () => {
      import("web-vitals")
        .then(({ getCLS, getFID, getFCP, getLCP, getTTFB }) => {
          getCLS(console.debug);
          getFID(console.debug);
          getFCP(console.debug);
          getLCP(console.debug);
          getTTFB(console.debug);
        })
        .catch(() => {});
    },
    { once: true },
  );
}
