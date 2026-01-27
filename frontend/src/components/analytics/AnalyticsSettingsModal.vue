<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h2>⚙️ Analytics Settings</h2>
        <button
          class="btn-close"
          type="button"
          aria-label="Close modal"
          @click="$emit('close')"
        >
          ×
        </button>
      </div>

      <div class="modal-body">
        <div class="form-group">
          <label>Price Column Name</label>
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
          <label>Formula: Expected Income</label>
          <div class="formula-box">
            <code
              >(Harga Marketplace - {{ localSettings.formula_deduction }}) ×
              {{ localSettings.formula_multiplier }}</code
            >
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Admin Fee (Rp)</label>
            <input
              v-model.number="localSettings.formula_deduction"
              type="number"
              min="0"
              class="form-control"
            />
            <p class="help-text">Shopee admin fee deduction</p>
          </div>
          <div class="form-group">
            <label>Multiplier</label>
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
          <h4>📝 Example Calculation</h4>
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
        <button class="btn-cancel" @click="$emit('close')">Cancel</button>
        <button class="btn-save" @click="handleSave">💾 Save Settings</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from "vue";
import type { AnalyticsSettings } from "@/composables/useAnalytics";
import { getAuthHeaders } from "@/utils/apiHeaders";

const props = defineProps<{ settings: AnalyticsSettings }>();
const emit = defineEmits<{
  (e: "close"): void;
  (e: "save", settings: AnalyticsSettings): void;
}>();

const localSettings = ref<AnalyticsSettings>({ ...props.settings });
const inventoryColumns = ref<string[]>([]);

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

onMounted(() => {
  fetchInventoryColumns();
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
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--z-modal);
}
.modal-content {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  width: 100%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-xl);
  border: 1px solid var(--color-border);
  color: var(--color-text-primary);
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-md) var(--space-lg);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
}
.modal-header h2 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--color-text-primary);
}
.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--color-text-muted);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  line-height: 1;
}
.btn-close:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}
.modal-body {
  padding: var(--space-lg);
}
.form-group {
  margin-bottom: var(--space-md);
}
.form-group label {
  display: block;
  font-weight: 600;
  margin-bottom: 6px;
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}
.form-control {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  font-size: 0.95rem;
}
.form-control:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
}
.form-control option {
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}
.help-text {
  margin: 4px 0 0;
  font-size: 0.8rem;
  color: var(--color-text-muted);
}
.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.formula-box {
  padding: 12px 16px;
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
}
.formula-box code {
  font-size: 0.95rem;
  color: var(--color-secondary);
  font-family: "Consolas", monospace;
}
.example-box {
  margin-top: var(--space-lg);
  padding: var(--space-md);
  background: var(--color-bg-secondary);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
}
.example-box h4 {
  margin: 0 0 12px;
  font-size: 0.9rem;
  color: var(--color-text-muted);
}
.example-content {
  font-size: 0.9rem;
  color: var(--color-text-secondary);
}
.example-content p {
  margin: 4px 0;
}
.example-content .result {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed var(--color-border);
  font-size: 1.1rem;
  color: var(--color-success);
}
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: var(--space-md) var(--space-lg);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
}
.btn-cancel,
.btn-save {
  padding: 10px 18px;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-weight: 500;
  font-size: 0.9rem;
}
.btn-cancel {
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}
.btn-cancel:hover {
  background: var(--color-neutral);
}
.btn-save {
  background: var(--color-primary);
  color: var(--color-text-inverse);
}
.btn-save:hover {
  opacity: 0.9;
}
</style>
