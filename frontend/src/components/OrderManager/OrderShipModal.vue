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
import { useOrderActions } from "./composables/useOrderActions";
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
  time_slots?: Array<{ pickup_time_id?: string; pickup_time?: string }>;
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
      pickup_time_id?: string;
      package_id?: string;
    },
  ];
}>();

const { form, loading, error, resetForm } = useModalForm({
  sp: "",
  tn: "",
  aid: "",
  ptid: "",
  pid: "",
});

const providers = computed<ShippingProvider[]>(() => getShippingProviders());

const pickupAddresses = ref<PickupAddress[]>([]);

const { getShippingParameters } = useOrderActions();

const loadShippingParams = async () => {
  if (!props.order?.order_no || !props.order.platform) {
    pickupAddresses.value = [];
    return;
  }

  const data = await getShippingParameters(
    props.order.order_no,
    props.order.platform,
  );

  if (Array.isArray(data)) {
    pickupAddresses.value = data.map((option) => ({
      address_id: option.address_id ?? option.logistic_id,
      address: option.address ?? option.logistic_name,
      time_slots: option.time_slots || [],
    }));
  } else {
    pickupAddresses.value = data?.pickup_addresses || data?.pickup || [];
  }
  if (form.value.aid && pickupAddresses.value.length > 0) {
    const match = pickupAddresses.value.find(
      (address) => address.address_id === Number(form.value.aid),
    );
    if (!match) {
      form.value.aid = "";
      form.value.ptid = "";
    }
  }
};

const valid = computed(() => {
  if (!props.order?.platform) return false;
  const platform = props.order.platform.toLowerCase();
  if (platform === "tiktok") {
    return form.value.tn.trim() !== "" && form.value.pid.trim() !== "";
  }
  return form.value.sp.trim() !== "";
});

const close = () => {
  if (!loading.value) {
    resetForm();
    emit("close");
  }
};

const submit = async () => {
  if (!valid.value || !props.order) return;
  const platform = props.order.platform?.toLowerCase();
  const provider = form.value.sp.trim();
  emit("confirm", {
    order_no: props.order.order_no,
    shipping_provider: provider,
    tracking_number: form.value.tn.trim() || undefined,
    address_id: form.value.aid ? Number(form.value.aid) : undefined,
    pickup_time_id: form.value.ptid || undefined,
    package_id: form.value.pid.trim() || undefined,
  });
};

watch(
  () => props.visible,
  (v) => {
    if (v) {
      resetForm();
      loadShippingParams();
    }
  },
);
</script>

<style scoped>
@import "./OrderManager.theme.css";

.o {
  position: fixed;
  inset: 0;
  background: var(--om-bg-overlay);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: var(--om-z-modal);
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
