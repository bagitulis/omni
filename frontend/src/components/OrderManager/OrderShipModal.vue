<template>
  <Teleport to="body">
    <div v-if="visible" class="o" @click.self="close">
      <div class="om">
        <div class="h">
          <h3>Arrange Shipment</h3>
          <button @click="close"><i class="pi pi-times"></i></button>
        </div>
        <div class="b">
          <!-- Order Info -->
          <div class="info-section">
            <div class="info-row">
              <span>Order No.</span><span>{{ order?.order_no }}</span>
            </div>
            <div class="info-row">
              <span>Platform</span
              ><span
                :class="[
                  'om-platform-badge',
                  `om-platform-${order?.platform?.toLowerCase()}`,
                ]"
                >{{ fmt(order?.platform) }}</span
              >
            </div>
            <div class="info-row">
              <span>Buyer</span><span>{{ order?.buyer_username }}</span>
            </div>
          </div>

          <!-- Form -->
          <form @submit.prevent="submit" class="form">
            <div>
              <label for="sp">Shipping Provider</label>
              <select
                id="sp"
                v-model="form.sp"
                class="select"
                :disabled="loading"
              >
                <option value="">Select Provider...</option>
                <option v-for="p in providers" :key="p.value" :value="p.value">
                  {{ p.label }}
                </option>
              </select>
            </div>

            <div>
              <label for="tn"
                >Tracking Number
                <span style="font-weight: 400; color: var(--om-text-secondary)"
                  >(Optional)</span
                ></label
              >
              <input
                id="tn"
                v-model="form.tn"
                type="text"
                class="input"
                placeholder="Enter tracking number..."
                :disabled="loading"
              />
            </div>

            <div
              v-if="
                order?.platform?.toLowerCase() === 'shopee' &&
                pickupAddresses.length
              "
            >
              <label for="pa">Pickup Address</label>
              <select
                id="pa"
                v-model="form.aid"
                class="select"
                :disabled="loading"
              >
                <option value="">Select Pickup Address...</option>
                <option
                  v-for="a in pickupAddresses"
                  :key="a.address_id"
                  :value="a.address_id"
                >
                  {{ a.address }}
                </option>
              </select>
            </div>

            <!-- Error Message -->
            <div v-if="error" class="error-box">
              <i class="pi pi-exclamation-circle"></i>
              <span>{{ error }}</span>
            </div>
          </form>
        </div>

        <!-- Footer -->
        <div class="footer">
          <button
            @click="close"
            class="om-btn om-btn-secondary"
            :disabled="loading"
          >
            Cancel
          </button>
          <button
            @click="submit"
            class="om-btn om-btn-primary"
            :disabled="loading || !valid"
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
import { computed, watch, ref } from "vue";
import {
  useModalForm,
  platformFormatter,
  getShippingProviders,
  type ShippingProvider,
} from "./composables/useModalForm";

interface Order {
  order_no: string;
  platform: string;
  buyer_username: string;
  [k: string]: any;
}

interface PickupAddress {
  address_id: number;
  address: string;
}

const props = defineProps<{ visible: boolean; order: Order | null }>();
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

const { form, loading, error, resetForm } = useModalForm({
  sp: "",
  tn: "",
  aid: "",
});

const providers = computed<ShippingProvider[]>(() => getShippingProviders());

const pickupAddresses = ref<PickupAddress[]>([]);

const valid = computed(() => form.value.sp !== "");

const fmt = (platform?: string): string => platformFormatter.format(platform);

const close = () => {
  if (!loading.value) {
    resetForm();
    emit("close");
  }
};

const submit = async () => {
  if (!valid.value || !props.order) return;
  emit("confirm", {
    order_no: props.order.order_no,
    shipping_provider: form.value.sp,
    tracking_number: form.value.tn || undefined,
    address_id: form.value.aid ? Number(form.value.aid) : undefined,
  });
};

watch(
  () => props.visible,
  (v) => {
    if (v) {
      resetForm();
    }
  },
);
</script>

<style scoped>
@import "./OrderManager.theme.css";

.o {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: om-fadeIn var(--om-transition-fast);
}

.om {
  background: var(--om-bg-primary);
  border-radius: var(--om-radius-lg);
  box-shadow: var(--om-shadow-lg);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  animation: om-slideUp var(--om-transition-normal);
}

.h {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-bottom: 1px solid var(--om-border);
}

.h h3 {
  font-size: var(--om-font-lg);
  font-weight: 600;
  color: var(--om-text-primary);
  margin: 0;
}

.h button {
  background: none;
  border: none;
  padding: var(--om-spacing-xs);
  cursor: pointer;
  color: var(--om-text-secondary);
  display: flex;
  border-radius: var(--om-radius-sm);
  transition: all var(--om-transition-fast);
}

.h button:hover {
  background: var(--om-bg-secondary);
  color: var(--om-text-primary);
}

.b {
  padding: var(--om-spacing-lg);
  overflow-y: auto;
  max-height: calc(90vh - 140px);
}

.info-section {
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
  font-size: var(--om-font-sm);
}

.info-row:not(:last-child) {
  border-bottom: 1px solid var(--om-border);
  padding-bottom: var(--om-spacing-sm);
  margin-bottom: var(--om-spacing-sm);
}

.info-row span:first-child {
  color: var(--om-text-secondary);
}

.info-row span:last-child {
  font-weight: 500;
  color: var(--om-text-primary);
}

.form {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-md);
}

.form > div {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-xs);
}

.form label {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
}

.select,
.input {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  transition: border-color var(--om-transition-fast);
}

.select:focus,
.input:focus {
  outline: none;
  border-color: var(--om-primary);
}

.select:disabled,
.input:disabled {
  background: var(--om-bg-secondary);
  cursor: not-allowed;
}

.error-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  background: #ffebee;
  color: var(--om-status-cancelled);
  border-radius: 4px;
  font-size: 0.875rem;
}

.footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-top: 1px solid var(--om-border);
}

@media (max-width: 768px) {
  .om {
    margin: var(--om-spacing-md);
    max-height: calc(100vh - 2rem);
  }
  .footer {
    flex-direction: column;
  }
  .footer .om-btn {
    width: 100%;
  }
}
</style>
