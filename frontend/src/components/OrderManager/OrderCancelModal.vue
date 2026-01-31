<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="close">
      <div class="modal-container">
        <!-- Modal Header -->
        <div class="modal-header">
          <h3 class="modal-title">Cancel Order</h3>
          <button @click="close" class="close-btn">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <!-- Modal Body -->
        <div class="modal-body">
          <!-- Warning Banner -->
          <div class="warning-banner">
            <i class="pi pi-exclamation-triangle"></i>
            <span
              >This action cannot be undone. Please confirm you want to cancel
              this order.</span
            >
          </div>

          <!-- Order Info -->
          <div class="order-info">
            <div class="info-row">
              <span class="info-label">Order No.</span>
              <span class="info-value">{{ order?.order_no }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Platform</span>
              <span
                :class="[
                  'om-platform-badge',
                  `om-platform-${order?.platform?.toLowerCase()}`,
                ]"
              >
                {{ formatPlatform(order?.platform) }}
              </span>
            </div>
            <div class="info-row">
              <span class="info-label">Buyer</span>
              <span class="info-value">{{ order?.buyer_username }}</span>
            </div>
          </div>

          <!-- Cancellation Form -->
          <form @submit.prevent="submitCancellation" class="cancel-form">
            <!-- Cancellation Reason -->
            <div class="form-group">
              <label for="cancelReason" class="form-label">
                Cancellation Reason <span class="required">*</span>
              </label>
              <select
                id="cancelReason"
                v-model="formData.cancel_reason"
                class="form-select"
                :disabled="loading"
                required
              >
                <option value="">Select a reason...</option>
                <option
                  v-for="reason in cancellationReasons"
                  :key="reason.value"
                  :value="reason.value"
                >
                  {{ reason.label }}
                </option>
              </select>
            </div>

            <!-- Reason Detail -->
            <div class="form-group">
              <label for="reasonDetail" class="form-label">
                Additional Details
                <span class="optional-text">(Optional)</span>
              </label>
              <textarea
                id="reasonDetail"
                v-model="formData.reason_detail"
                class="form-textarea"
                placeholder="Provide additional details..."
                rows="3"
                :disabled="loading"
              ></textarea>
            </div>

            <!-- Confirmation Checkbox -->
            <div class="form-group checkbox-group">
              <label class="checkbox-label">
                <input
                  type="checkbox"
                  v-model="formData.confirmed"
                  :disabled="loading"
                />
                <span
                  >I confirm that I want to cancel order
                  <strong>{{ order?.order_no }}</strong></span
                >
              </label>
            </div>

            <!-- Error Message -->
            <div v-if="error" class="error-message">
              <i class="pi pi-exclamation-circle"></i>
              <span>{{ error }}</span>
            </div>
          </form>
        </div>

        <!-- Modal Footer -->
        <div class="modal-footer">
          <button
            @click="close"
            class="om-btn om-btn-secondary"
            :disabled="loading"
          >
            Go Back
          </button>
          <button
            @click="submitCancellation"
            class="om-btn cancel-confirm-btn"
            :disabled="loading || !isFormValid"
          >
            <i v-if="loading" class="pi pi-spin pi-spinner"></i>
            <i v-else class="pi pi-times"></i>
            <span>{{ loading ? "Processing..." : "Cancel Order" }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";

interface Order {
  order_no: string;
  platform: string;
  buyer_username: string;
  [key: string]: any;
}

interface CancellationReason {
  value: string;
  label: string;
}

interface Props {
  visible: boolean;
  order: Order | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  confirm: [
    data: { order_no: string; cancel_reason: string; reason_detail?: string },
  ];
}>();

// Form state
const formData = ref({
  cancel_reason: "",
  reason_detail: "",
  confirmed: false,
});
const loading = ref(false);
const error = ref("");

// Platform-specific cancellation reasons
const cancellationReasons = computed<CancellationReason[]>(() => {
  const platform = props.order?.platform?.toLowerCase();

  // Common reasons for all platforms
  const commonReasons = [
    { value: "OUT_OF_STOCK", label: "Out of stock" },
    { value: "BUYER_REQUEST", label: "Buyer requested cancellation" },
    { value: "WRONG_PRICE", label: "Wrong price/listing error" },
    { value: "DUPLICATE_ORDER", label: "Duplicate order" },
    { value: "OTHER", label: "Other reason" },
  ];

  // Platform-specific reasons
  if (platform === "shopee") {
    return [
      { value: "CUSTOMER_REQUEST", label: "Customer requested cancellation" },
      { value: "OUT_OF_STOCK", label: "Out of stock" },
      { value: "UNDELIVERABLE_AREA", label: "Undeliverable area" },
      { value: "COD_NOT_SUPPORTED", label: "COD not supported" },
      ...commonReasons.filter(
        (r) => !["OUT_OF_STOCK", "BUYER_REQUEST"].includes(r.value),
      ),
    ];
  }

  if (platform === "lazada") {
    return [
      { value: "customer_request", label: "Customer requested cancellation" },
      { value: "out_of_stock", label: "Out of stock" },
      { value: "sourcing_failed", label: "Sourcing failed" },
      ...commonReasons.filter(
        (r) => !["OUT_OF_STOCK", "BUYER_REQUEST"].includes(r.value),
      ),
    ];
  }

  return commonReasons;
});

// Computed
const isFormValid = computed(() => {
  return formData.value.cancel_reason !== "" && formData.value.confirmed;
});

// Methods
const formatPlatform = (platform?: string): string => {
  if (!platform) return "";
  const names: Record<string, string> = {
    shopee: "Shopee",
    lazada: "Lazada",
    tiktok: "TikTok",
  };
  return names[platform.toLowerCase()] || platform;
};

const close = () => {
  if (!loading.value) {
    resetForm();
    emit("close");
  }
};

const resetForm = () => {
  formData.value = {
    cancel_reason: "",
    reason_detail: "",
    confirmed: false,
  };
  error.value = "";
};

const submitCancellation = async () => {
  if (!isFormValid.value || !props.order) return;

  loading.value = true;
  error.value = "";

  try {
    emit("confirm", {
      order_no: props.order.order_no,
      cancel_reason: formData.value.cancel_reason,
      reason_detail: formData.value.reason_detail || undefined,
    });
  } catch (err: any) {
    error.value = err.message || "Failed to cancel order";
  } finally {
    loading.value = false;
  }
};

// Watch for modal open to reset form
watch(
  () => props.visible,
  (newVal) => {
    if (newVal) {
      resetForm();
    }
  },
);
</script>

<style scoped>
@import "./OrderManager.theme.css";

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: om-fadeIn var(--om-transition-fast);
}

.modal-container {
  background: var(--om-bg-primary);
  border-radius: var(--om-radius-lg);
  box-shadow: var(--om-shadow-lg);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  animation: om-slideUp var(--om-transition-normal);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-bottom: 1px solid var(--om-border);
}

.modal-title {
  font-size: var(--om-font-lg);
  font-weight: 600;
  color: var(--om-text-primary);
  margin: 0;
}

.close-btn {
  background: none;
  border: none;
  padding: var(--om-spacing-xs);
  cursor: pointer;
  color: var(--om-text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--om-radius-sm);
  transition: all var(--om-transition-fast);
}

.close-btn:hover {
  background: var(--om-bg-secondary);
  color: var(--om-text-primary);
}

.modal-body {
  padding: var(--om-spacing-lg);
  overflow-y: auto;
  max-height: calc(90vh - 140px);
}

.warning-banner {
  display: flex;
  align-items: flex-start;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-md);
  background: #fff3e0;
  border: 1px solid #ffb74d;
  border-radius: var(--om-radius-md);
  margin-bottom: var(--om-spacing-lg);
  color: #e65100;
  font-size: var(--om-font-sm);
}

.warning-banner i {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.order-info {
  background: var(--om-bg-secondary);
  border-radius: var(--om-radius-md);
  padding: var(--om-spacing-md);
  margin-bottom: var(--om-spacing-lg);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-xs) 0;
}

.info-row:not(:last-child) {
  border-bottom: 1px solid var(--om-border);
  padding-bottom: var(--om-spacing-sm);
  margin-bottom: var(--om-spacing-sm);
}

.info-label {
  font-size: var(--om-font-sm);
  color: var(--om-text-secondary);
}

.info-value {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
}

.cancel-form {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-md);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-xs);
}

.form-label {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
}

.required {
  color: var(--om-status-cancelled);
}

.optional-text {
  font-weight: 400;
  color: var(--om-text-secondary);
}

.form-select,
.form-textarea {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  font-family: inherit;
  transition: border-color var(--om-transition-fast);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.form-select:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--om-primary);
}

.form-select:disabled,
.form-textarea:disabled {
  background: var(--om-bg-secondary);
  cursor: not-allowed;
}

.checkbox-group {
  margin-top: var(--om-spacing-sm);
}

.checkbox-label {
  display: flex;
  align-items: flex-start;
  gap: var(--om-spacing-sm);
  cursor: pointer;
  font-size: var(--om-font-sm);
  color: var(--om-text-secondary);
}

.checkbox-label input[type="checkbox"] {
  margin-top: 2px;
  accent-color: var(--om-primary);
}

.error-message {
  display: flex;
  align-items: center;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-sm);
  background: #ffebee;
  color: var(--om-status-cancelled);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-top: 1px solid var(--om-border);
}

.cancel-confirm-btn {
  background: var(--om-status-cancelled);
  color: white;
}

.cancel-confirm-btn:hover:not(:disabled) {
  background: #c62828;
}

.cancel-confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .modal-container {
    margin: var(--om-spacing-md);
    max-height: calc(100vh - 2rem);
  }

  .modal-footer {
    flex-direction: column;
  }

  .modal-footer .om-btn {
    width: 100%;
  }
}
</style>
