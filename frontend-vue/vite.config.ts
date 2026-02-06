import { defineConfig, Plugin } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";
import { fileURLToPath } from "url";
import compression from "vite-plugin-compression";
import { visualizer } from "rollup-plugin-visualizer";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

/**
 * Custom plugin to make CSS non-render-blocking
 * Converts <link rel="stylesheet"> to async loading pattern
 */
function asyncCssPlugin(): Plugin {
  return {
    name: "async-css",
    enforce: "post",
    transformIndexHtml(html) {
      // Convert CSS links to async loading with print/all media trick
      return html.replace(
        /<link rel="stylesheet"([^>]*) href="([^"]+)"([^>]*)>/g,
        (match, before, href, after) => {
          // Keep critical inline styles, only async external CSS
          if (href.includes("critical")) return match;
          return `<link rel="stylesheet"${before} href="${href}"${after} media="print" onload="this.media='all'">
<noscript><link rel="stylesheet" href="${href}"></noscript>`;
        }
      );
    },
  };
}

/**
 * Plugin to add preload hints for critical chunks
 */
function preloadPlugin(): Plugin {
  return {
    name: "preload-chunks",
    enforce: "post",
    transformIndexHtml(html, ctx) {
      if (!ctx.bundle) return html;
      
      // Find critical JS chunks to preload
      const criticalChunks = ["vue-core", "index", "app"];
      const preloadTags: string[] = [];
      
      const bundleEntries = Object.entries(ctx.bundle);
      for (const [fileName] of bundleEntries) {
        if (fileName.endsWith(".js")) {
          const isCritical = criticalChunks.some(c => fileName.includes(c));
          if (isCritical && !fileName.includes("chunk")) {
            preloadTags.push(`<link rel="modulepreload" href="/${fileName}">`);
          }
        }
      }
      
      // Insert preload tags before </head>
      return html.replace("</head>", `${preloadTags.join("\n")}\n</head>`);
    },
  };
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // Async CSS loading for non-render-blocking
    asyncCssPlugin(),
    // Preload critical chunks
    preloadPlugin(),
    compression({
      ext: ".gz",
      algorithm: "gzip",
      deleteOriginFile: false,
    }),
    visualizer({
      open: false,
      gzipSize: true,
      brotliSize: true,
      filename: "bundle-analysis.html",
      title: "Bundle Analysis",
    }),
  ],
  // Optimize dependencies - pre-bundle vendors for faster load
  optimizeDeps: {
    include: ["vue", "vue-router", "pinia", "axios"],
    exclude: ["@vite/client"],
    // Disable auto discovery for better control
    entries: ["./src/main.ts"],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@components": path.resolve(__dirname, "./src/components"),
      "@services": path.resolve(__dirname, "./src/services"),
      "@store": path.resolve(__dirname, "./src/store"),
      "@types": path.resolve(__dirname, "./src/types"),
      "@utils": path.resolve(__dirname, "./src/utils"),
      "@views": path.resolve(__dirname, "./src/views"),
    },
  },
  build: {
    // Split code aggressively
    rollupOptions: {
      // Optimize tree-shaking
      treeshake: {
        moduleSideEffects: false,
        propertyReadSideEffects: false,
        tryCatchDeoptimization: false,
        preset: "recommended",
      },
      output: {
        manualChunks: {
          // Vue core - must load first
          "vue-core": ["vue", "vue-router"],
          // State management
          state: ["pinia"],
          // API client
          api: ["axios"],
        },
        compact: true,
        // Optimize for HTTP/2 server push
        entryFileNames: "js/[name]-[hash].js",
        chunkFileNames: "js/[name]-[hash].js",
        assetFileNames: ({ name }) => {
          if (/\.css$/.test(name ?? "")) {
            return "css/[name]-[hash][extname]";
          }
          if (/\.(png|jpg|jpeg|gif|webp|svg)$/.test(name ?? "")) {
            return "images/[name]-[hash][extname]";
          }
          return "assets/[name]-[hash][extname]";
        },
      },
    },
    // Aggressive production minification for max compression
    minify: "terser",
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
        passes: 2,
        pure_funcs: ["console.log", "console.debug", "console.info"],
        pure_getters: true,
        // Removed unsafe options that can break third-party libraries
        unused: true,
        toplevel: true,
        arrows: true,
        hoist_funs: true,
        hoist_vars: false,
        inline: 2,
        join_vars: true,
        reduce_vars: true,
        side_effects: true,
        typeofs: false,
        evaluate: true,
        booleans: true,
        sequences: true,
        properties: true,
        dead_code: true,
      },
      mangle: {
        toplevel: true,
        // Removed property mangling that can break libraries
        keep_fnames: false,
      },
      format: {
        comments: false,
        beautify: false,
        max_line_len: 80000,
        ascii_only: false,
      },
      ecma: 2020,
      toplevel: true,
    },
    // Target modern browsers only - smaller output
    target: "es2020",
    // CSS code splitting
    cssCodeSplit: true,
    // Report compressed size
    reportCompressedSize: true,
    // Chunk size warnings - lower threshold to catch issues
    chunkSizeWarningLimit: 300,
    // Reduce main thread work
    sourcemap: false,
    // Enable compression
    commonjsOptions: {
      transformMixedEsModules: true,
      esmExternals: true,
    },
    // Inline small imports
    assetsInlineLimit: 4096,
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:3000",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path,
      },
    },
    port: 5173,
    host: "0.0.0.0", // Izinkan akses dari jaringan
    cors: true,
  },
  preview: {
    proxy: {
      "/api": {
        target: "http://localhost:3000",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path,
      },
    },
    port: 4173,
    host: "0.0.0.0",
    cors: true,
  },
  define: {
    __APP_VERSION__: JSON.stringify("1.0.0"),
  },
});
