<template>
  <div class="dev-preview">
    <!-- Header -->
    <div class="preview-header">
      <h1>🔧 Dev Preview Mode</h1>
      <p>Preview komponen tanpa login - <strong>Development Only</strong></p>
    </div>

    <!-- Component Selector -->
    <div class="preview-selector">
      <label>Pilih Komponen:</label>
      <select v-model="selectedComponent" @change="loadComponent">
        <option value="">-- Pilih --</option>
        <optgroup label="Layout">
          <option value="LeftSidebar">LeftSidebar</option>
          <option value="RightSidebar">RightSidebar</option>
          <option value="Navbar">Navbar</option>
          <option value="DashboardHeader">DashboardHeader</option>
        </optgroup>
        <optgroup label="Common">
          <option value="Button">Button</option>
          <option value="Modal">Modal</option>
          <option value="Toast">Toast</option>
          <option value="Tabs">Tabs (New)</option>
          <option value="Skeleton">Skeleton (New)</option>
        </optgroup>
        <optgroup label="Pages">
          <option value="Dashboard">Dashboard (Full)</option>
          <option value="OrderManager">OrderManager</option>
          <option value="ProductManager">ProductManager</option>
          <option value="Settings">Settings</option>
        </optgroup>
      </select>
    </div>

    <!-- Preview Area -->
    <div class="preview-area">
      <div class="preview-container" :class="previewSize">
        <component
          :is="currentComponent"
          v-if="currentComponent"
          v-bind="componentProps"
        />
        <div v-else class="empty-preview">
          <span>👆 Pilih komponen untuk preview</span>
        </div>
      </div>
    </div>

    <!-- Size Controls -->
    <div class="size-controls">
      <button
        v-for="size in sizes"
        :key="size.value"
        :class="['size-btn', { active: previewSize === size.value }]"
        @click="previewSize = size.value"
      >
        {{ size.label }}
      </button>
    </div>

    <!-- Theme Toggle -->
    <div class="theme-controls">
      <button @click="toggleTheme" class="theme-btn">
        {{ isDark ? "☀️ Light Mode" : "🌙 Dark Mode" }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, computed, watch } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();
const selectedComponent = ref((route.params.component as string) || "");
const currentComponent = shallowRef<any>(null);
const previewSize = ref("desktop");
const isDark = ref(document.documentElement.classList.contains("dark"));

const sizes = [
  { label: "📱 Mobile", value: "mobile" },
  { label: "📱 Tablet", value: "tablet" },
  { label: "🖥️ Desktop", value: "desktop" },
  { label: "🖥️ Wide", value: "wide" },
];

// Component map for lazy loading
const componentMap: Record<string, () => Promise<any>> = {
  // Layout
  LeftSidebar: () => import("@/components/layout/LeftSidebar.vue"),
  RightSidebar: () => import("@/components/layout/sidebar/RightSidebar.vue"),
  Navbar: () => import("@/components/Navbar.vue"),
  DashboardHeader: () => import("@/components/layout/DashboardHeader.vue"),
  // Common
  Button: () => import("@/components/Button.vue"),
  Modal: () => import("@/components/Modal.vue"),
  Toast: () => import("@/components/Toast.vue"),
  // Pages (wrap in error boundary)
  Dashboard: () => import("@/views/Dashboard.vue"),
  OrderManager: () => import("@/components/OrderManager/OrderManager.vue"),
  ProductManager: () =>
    import("@/components/ProductManager/ProductManager.vue"),
  Settings: () => import("@/views/Settings.vue"),
};

// Default props for components that need them
const componentPropsMap: Record<string, any> = {
  LeftSidebar: {
    collapsed: false,
    activeTab: "operation",
    activePlatform: "shopee",
  },
  RightSidebar: { collapsed: false, status: {}, loading: false },
  DashboardHeader: { connectionStatus: "connected", loading: false },
  Button: { label: "Sample Button", class: "btn-primary" },
  Modal: { isOpen: true, title: "Sample Modal" },
  ProductManager: { platform: "shopee" },
};

const componentProps = computed(
  () => componentPropsMap[selectedComponent.value] || {}
);

const loadComponent = async () => {
  if (!selectedComponent.value || !componentMap[selectedComponent.value]) {
    currentComponent.value = null;
    return;
  }

  try {
    const module = await componentMap[selectedComponent.value]();
    currentComponent.value = module.default;
  } catch (error) {
    console.error("Failed to load component:", error);
    currentComponent.value = null;
  }
};

// Watch for URL param changes (browser back/forward, direct URL access)
watch(
  () => route.params.component,
  (newComponent) => {
    const componentValue = newComponent as string;
    if (componentValue !== selectedComponent.value) {
      selectedComponent.value = componentValue || "";
      if (componentValue) {
        loadComponent();
      } else {
        currentComponent.value = null;
      }
    }
  }
);

const toggleTheme = () => {
  isDark.value = !isDark.value;
  document.documentElement.classList.toggle("dark", isDark.value);
};

// Load from URL param if provided
if (selectedComponent.value) {
  loadComponent();
}
</script>

<style scoped>
.dev-preview {
  min-height: 100vh;
  background: #f0f0f0;
  padding: 20px;
}

.dark .dev-preview {
  background: #1a1a2e;
}

.preview-header {
  text-align: center;
  margin-bottom: 20px;
  padding: 15px;
  background: linear-gradient(135deg, #ff6b6b, #feca57);
  border-radius: 8px;
  color: white;
}

.preview-header h1 {
  margin: 0 0 5px 0;
  font-size: 1.5rem;
}

.preview-header p {
  margin: 0;
  font-size: 0.9rem;
  opacity: 0.9;
}

.preview-selector {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  padding: 15px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.dark .preview-selector {
  background: #16213e;
  color: white;
}

.preview-selector select {
  flex: 1;
  padding: 10px;
  border-radius: 6px;
  border: 1px solid #ddd;
  font-size: 1rem;
}

.dark .preview-selector select {
  background: #0f3460;
  border-color: #1a1a2e;
  color: white;
}

.preview-area {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.preview-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  overflow: hidden;
  transition: width 0.3s ease;
}

.dark .preview-container {
  background: #16213e;
}

.preview-container.mobile {
  width: 375px;
  min-height: 667px;
}

.preview-container.tablet {
  width: 768px;
  min-height: 500px;
}

.preview-container.desktop {
  width: 1024px;
  min-height: 600px;
}

.preview-container.wide {
  width: 100%;
  max-width: 1400px;
  min-height: 700px;
}

.empty-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  color: #6b7280;
  font-size: 1.2rem;
}

.size-controls,
.theme-controls {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-bottom: 15px;
}

.size-btn,
.theme-btn {
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  background: white;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.dark .size-btn,
.dark .theme-btn {
  background: #16213e;
  color: white;
}

.size-btn:hover,
.theme-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.size-btn.active {
  background: #667eea;
  color: white;
}
</style>
