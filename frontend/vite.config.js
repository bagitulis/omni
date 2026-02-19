var __spreadArray = (this && this.__spreadArray) || function (to, from, pack) {
    if (pack || arguments.length === 2) for (var i = 0, l = from.length, ar; i < l; i++) {
        if (ar || !(i in from)) {
            if (!ar) ar = Array.prototype.slice.call(from, 0, i);
            ar[i] = from[i];
        }
    }
    return to.concat(ar || Array.prototype.slice.call(from));
};
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { visualizer } from "rollup-plugin-visualizer";
import path from "path";
export default defineConfig({
    plugins: __spreadArray([
        react()
    ], (process.env.ANALYZE === "true"
        ? [
            visualizer({
                filename: "dist/stats.html",
                open: false,
                gzipSize: true,
                brotliSize: true,
            }),
        ]
        : []), true),
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        port: 5174,
        host: "0.0.0.0",
        proxy: {
            "/api": {
                target: "http://localhost:3000",
                changeOrigin: true,
            },
        },
    },
    build: {
        target: "es2020",
        // Optimize chunk size
        chunkSizeWarningLimit: 600,
        rollupOptions: {
            output: {
                // Manual chunks for better caching
                manualChunks: {
                    // Vendor chunk - React core
                    "vendor-react": ["react", "react-dom", "react-router-dom"],
                    // Ant Design - large UI library
                    "vendor-antd": ["antd", "@ant-design/icons"],
                    // Charts library
                    "vendor-charts": ["apexcharts", "react-apexcharts"],
                    // Data fetching libraries
                    "vendor-query": ["@tanstack/react-query", "axios"],
                    // State management
                    "vendor-state": ["zustand"],
                },
            },
        },
        // Enable minification
        minify: "terser",
        terserOptions: {
            compress: {
                drop_console: true, // Remove console.log in production
                drop_debugger: true,
            },
        },
    },
});
