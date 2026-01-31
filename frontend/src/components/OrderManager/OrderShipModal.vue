<template>
  <Teleport to="body">
    <div v-if="visible" class="modal-overlay" @click.self="close">
      <div class="modal-container">
        <!-- Modal Header -->
        <div class="modal-header">
          <h3 class="modal-title">Arrange Shipment</h3>
          <button @click="close" class="close-btn">
            <i class="pi pi-times"></i>
          </button>
        </div>

        <!-- Modal Body -->
        <div class="modal-body">
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

          <!-- Shipping Form -->
          <form @submit.prevent="submitShipment" class="shipping-form">
            <!-- Shipping Provider (for non-integrated) -->
            <div class="form-group">
              <label for="shippingProvider" class="form-label"
                >Shipping Provider</label
              >
              <select
                id="shippingProvider"
                v-model="formData.shipping_provider"
                class="form-select"
                :disabled="loading"
              >
                <option value="">Select Provider...</option>
                <option
                  v-for="provider in shippingProviders"
                  :key="provider.value"
                  :value="provider.value"
                >
                  {{ provider.label }}
                </option>
              </select>
            </div>

            <!-- Tracking Number (optional for some platforms) -->
            <div class="form-group">
              <label for="trackingNumber" class="form-label">
                Tracking Number
                <span class="optional-text">(Optional)</span>
              </label>
              <input
                id="trackingNumber"
                v-model="formData.tracking_number"
                type="text"
                class="form-input"
                placeholder="Enter tracking number..."
                :disabled="loading"
              />
            </div>

            <!-- Address Selection for Pickup (Shopee) -->
            <div
              v-if="
                order?.platform?.toLowerCase() === 'shopee' &&
                pickupAddresses.length > 0
              "
              class="form-group"
            >
              <label for="pickupAddress" class="form-label"
                >Pickup Address</label
              >
              <select
                id="pickupAddress"
                v-model="formData.address_id"
                class="form-select"
                :disabled="loading"
              >
                <option value="">Select Pickup Address...</option>
                <option
                  v-for="addr in pickupAddresses"
                  :key="addr.address_id"
                  :value="addr.address_id"
                >
                  {{ addr.address }}
                </option>
              </select>
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
            Cancel
          </button>
          <button
            @click="submitShipment"
            class="om-btn om-btn-primary"
            :disabled="loading || !isFormValid"
          >
            <i v-if="loading" class="pi pi-spin pi-spinner"></i>
            <i v-else class="pi pi-truck"></i>
            <span>{{ loading ? "Processing..." : "Confirm Shipment" }}</span>
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

interface ShippingProvider {
  value: string;
  label: string;
}

interface PickupAddress {
  address_id: number;
  address: string;
}

interface Props {
  visible: boolean;
  order: Order | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  confirm: [
    data: {
      order_no: string;
      shipping_provider: string;
      tracking_number?: string;
      address_id?: number;
    },
  ];
}>();

// Form state
const formData = ref({
  shipping_provider: "",
  tracking_number: "",
  address_id: "",
});
const loading = ref(false);
const error = ref("");

// Mock data - in real implementation, fetch from API
const shippingProviders = ref<ShippingProvider[]>([
  { value: "jne", label: "JNE" },
  { value: "jnt", label: "J&T Express" },
  { value: "sicepat", label: "SiCepat" },
  { value: "anteraja", label: "AnterAja" },
  { value: "ninja", label: "Ninja Van" },
  { value: "shopee_express", label: "Shopee Express" },
  { value: "lazada_logistics", label: "Lazada Logistics" },
]);

const pickupAddresses = ref<PickupAddress[]>([]);

// Computed
const isFormValid = computed(() => {
  return formData.value.shipping_provider !== "";
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
    shipping_provider: "",
    tracking_number: "",
    address_id: "",
  };
  error.value = "";
};

const submitShipment = async () => {
  if (!isFormValid.value || !props.order) return;

  loading.value = true;
  error.value = "";

  try {
    emit("confirm", {
      order_no: props.order.order_no,
      shipping_provider: formData.value.shipping_provider,
      tracking_number: formData.value.tracking_number || undefined,
      address_id: formData.value.address_id
        ? Number(formData.value.address_id)
        : undefined,
    });
  } catch (err: any) {
    error.value = err.message || "Failed to process shipment";
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
      // TODO: Fetch shipping parameters from API based on order
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

.shipping-form {
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

.optional-text {
  font-weight: 400;
  color: var(--om-text-secondary);
}

.form-input,
.form-select {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  transition: border-color var(--om-transition-fast);
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: var(--om-primary);
}

.form-input:disabled,
.form-select:disabled {
  background: var(--om-bg-secondary);
  cursor: not-allowed;
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
