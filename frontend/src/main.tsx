import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { initPerformanceMonitoring } from "./lib/webVitals";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);

// Initialize Web Vitals monitoring after app mount
initPerformanceMonitoring();
