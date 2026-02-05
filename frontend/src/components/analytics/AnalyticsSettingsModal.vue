<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="modal-overlay"
      @click.self="$emit('close')"
      role="dialog"
      aria-modal="true"
      aria-labelledby="analytics-settings-title"
    >
      <div class="modal-container">
        <div class="modal-header">
          <h2 id="analytics-settings-title" class="modal-title">
            <Icon name="settings" size="md" /> Analytics Settings
          </h2>
          <button
            class="modal-close-btn"
            type="button"
            aria-label="Close modal"
            @click="$emit('close')"
          >
            <svg
              width="20"
              height="20"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">Price Column Name</label>
            <select v-model="localSettings.price_column" class="form-control">
              <option value="">-- Select Column --</option>
              <option v-for="col in priceColumns" :key="col" :value="col">
                {{ col }}
              </option>
            </select>
            <p class="help-text">
              Column name in inventory that contains marketplace price
            </p>
          </div>

          <div class="form-group">
            <label class="form-label">Formula: Expected Income</label>
            <div class="formula-box">
              <code
                >(Harga Marketplace - {{ localSettings.formula_deduction }}) ×
                {{ localSettings.formula_multiplier }}</code
              >
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">Admin Fee (Rp)</label>
              <input
                v-model.number="localSettings.formula_deduction"
                type="number"
                min="0"
                class="form-control"
              />
              <p class="help-text">Shopee admin fee deduction</p>
            </div>
            <div class="form-group">
              <label class="form-label">Multiplier</label>
              <input
                v-model.number="localSettings.formula_multiplier"
                type="number"
                step="0.01"
                min="0"
                max="1"
                class="form-control"
              />
              <p class="help-text">After-fee percentage (0.84 = 84%)</p>
            </div>
          </div>

          <div class="example-box">
            <h4 class="example-title">
              <Icon name="document" size="sm" /> Example Calculation
            </h4>
            <div class="example-content">
              <p>If Harga Marketplace = <strong>Rp 100,000</strong></p>
              <p>
                Expected = (100,000 - {{ localSettings.formula_deduction }}) ×
                {{ localSettings.formula_multiplier }}
              </p>
              <p class="result">
                = <strong>{{ formatPrice(calculateExample()) }}</strong>
              </p>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="modal-btn modal-btn-secondary" @click="$emit('close')">
            Cancel
          </button>
          <button
            class="modal-btn modal-btn-primary"
            aria-label="Save settings"
            @click="handleSave"
          >
            <Icon name="download" size="sm" /> Save Settings
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, computed } from "vue";
import type { AnalyticsSettings } from "@/composables/useAnalytics";
import { getAuthHeaders, getApiBaseUrl } from "@/utils/apiHeaders";
import Icon from "@/components/ui/Icon.vue";

const props = defineProps<{ settings: AnalyticsSettings; visible: boolean }>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", settings: AnalyticsSettings): void;
}>();

const localSettings = ref<AnalyticsSettings>({ ...props.settings });
const inventoryColumns = ref<string[]>([]);

// ESC key to close modal
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.visible) {
    emit("close");
  }
};

// Body scroll lock
watch(
  () => props.visible,
  (visible) => {
    document.body.style.overflow = visible ? "hidden" : "";
  },
);

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
  fetchInventoryColumns();
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});

async function fetchInventoryColumns() {
  try {
    // Use centralized API URL function (supports localhost dev + production nginx proxy)
    const apiUrl = getApiBaseUrl("/inventory/stats");
    const response = await fetch(apiUrl, {
      headers: getAuthHeaders(),
    });
    if (response.ok) {
      const data = await response.json();
      if (data.success && data.data?.columns) {
        inventoryColumns.value = data.data.columns.map(
          (c: any) => c.column_name || c.name || c,
        );
      }
    }
  } catch (error) {
    console.error("Failed to fetch inventory columns:", error);
  }
}

const priceColumns = computed(() => {
  const suggested = inventoryColumns.value.filter((name) => {
    const lower = name.toLowerCase();
    return (
      lower.includes("harga") ||
      lower.includes("price") ||
      lower.includes("cost") ||
      lower.includes("nilai")
    );
  });
  const others = inventoryColumns.value.filter(
    (name) => !suggested.includes(name),
  );
  return [...suggested, ...others];
});

watch(
  () => props.settings,
  (newVal) => {
    localSettings.value = { ...newVal };
  },
);

function calculateExample(): number {
  const price = 100000;
  return (
    (price - localSettings.value.formula_deduction) *
    localSettings.value.formula_multiplier
  );
}

function formatPrice(price: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(price);
}

function handleSave() {
  emit("save", localSettings.value);
}
</script>

<style scoped>
/* Modal Overlay - Tokopedia Style */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
  animation: fadeIn 200ms ease-out;
}

/* Modal Container */
.modal-container {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: slideUp 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Modal Header */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

.modal-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 1.125rem;
  font-weight: 700;
  color: #212121;
}

.modal-close-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: #5f656e;
  cursor: pointer;
  transition: all 150ms ease;
}

.modal-close-btn:hover {
  background: #f3f4f5;
  color: #212121;
}

.modal-close-btn:focus-visible {
  outline: 2px solid #027a0a;
  outline-offset: 2px;
}

/* Modal Body */
.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: #ffffff;
}

.modal-body::-webkit-scrollbar {
  width: 6px;
}

.modal-body::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 3px;
}

/* Form Elements */
.form-group {
  margin-bottom: 16px;
}

.form-label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  color: #212121;
  font-size: 0.875rem;
}

.form-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #ffffff;
  color: #212121;
  font-size: 0.875rem;
  transition:
    border-color 150ms ease,
    box-shadow 150ms ease;
}

.form-control:focus {
  outline: none;
  border-color: #027a0a;
  box-shadow: 0 0 0 3px rgba(2, 122, 10, 0.1);
}

.help-text {
  margin: 6px 0 0;
  font-size: 0.75rem;
  color: #5f656e;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

/* Formula Box */
.formula-box {
  padding: 12px 16px;
  background: #f3f4f5;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

.formula-box code {
  font-size: 0.875rem;
  color: #027a0a;
  font-family: "Consolas", "Monaco", monospace;
}

/* Example Box */
.example-box {
  margin-top: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

.example-title {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 12px;
  font-size: 0.875rem;
  font-weight: 600;
  color: #5f656e;
}

.example-content {
  font-size: 0.875rem;
  color: #212121;
}

.example-content p {
  margin: 6px 0;
}

.example-content .result {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #e0e0e0;
  font-size: 1rem;
  color: #027a0a;
}

/* Modal Footer */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

/* Buttons */
.modal-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 150ms ease;
  border: 1px solid transparent;
}

.modal-btn-secondary {
  background: #ffffff;
  color: #212121;
  border-color: #e0e0e0;
}

.modal-btn-secondary:hover {
  background: #f3f4f5;
  border-color: #bdbdbd;
}

.modal-btn-primary {
  background: #027a0a;
  color: white;
  border-color: #027a0a;
}

.modal-btn-primary:hover {
  background: #026208;
  border-color: #026208;
}

.modal-btn:focus-visible {
  outline: 2px solid #027a0a;
  outline-offset: 2px;
}

/* Animations */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Responsive */
@media (max-width: 640px) {
  .modal-overlay {
    padding: 0;
  }

  .modal-container {
    max-width: 100%;
    max-height: 100%;
    height: 100%;
    border-radius: 0;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .modal-footer {
    flex-direction: column;
  }

  .modal-btn {
    width: 100%;
  }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .modal-overlay,
  .modal-container {
    animation: none;
  }
}
</style>
