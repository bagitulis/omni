import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
export default defineConfig({
    plugins: [react()],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        port: 5173,
        host: "0.0.0.0",
        proxy: {
            "/api": {
                target: "http://localhost:3000",
                changeOrigin: true,
            },
        },
    },
    build: {
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
