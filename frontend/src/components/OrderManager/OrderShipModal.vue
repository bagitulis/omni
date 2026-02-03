<template>
  <Teleport to="body">
    <div v-if="visible" class="o" @click.self="close">
      <div class="om">
        <div class="h">
          <h3>Arrange Shipment</h3>
          <button @click="close"><Icon name="close" size="md" /></button>
        </div>
        <div class="b">
          <OrderShipInfo :order="order" />
          <OrderShipForm
            v-model="form"
            :loading="loading"
            :error="error"
            :pickupAddresses="pickupAddresses"
            :order="order"
            :providers="providers"
            @submit="submit"
          />
        </div>
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
            <Icon v-if="loading" name="spinner" size="sm" spin />
            <Icon v-else name="truck" size="sm" />
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
  getShippingProviders,
  type ShippingProvider,
} from "./composables/useModalForm";
import OrderShipInfo from "./partials/OrderShipInfo.vue";
import OrderShipForm from "./partials/OrderShipForm.vue";
import Icon from "@/components/ui/Icon.vue";

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
