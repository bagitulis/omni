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
            :dropoffBranches="dropoffBranches"
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

interface DropoffBranch {
  branch_id: number;
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
      pickup_time_id?: string;
      branch_id?: number;
      package_id?: string;
    },
  ];
}>();

const { form, loading, error, resetForm } = useModalForm({
  sp: "",
  tn: "",
  aid: "",
  ptid: "",
  pid: props.order?.package_id || "",
  bid: "",
  mode: "",
});

const providers = computed<ShippingProvider[]>(() => getShippingProviders());

const pickupAddresses = ref<PickupAddress[]>([]);
const dropoffBranches = ref<DropoffBranch[]>([]);

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
      time_slots: option.time_slots || option.time_slot_list || [],
    }));
    dropoffBranches.value = [];
  } else {
    pickupAddresses.value = data?.pickup_addresses || data?.pickup || [];
    dropoffBranches.value = data?.dropoff || data?.dropoff_branches || [];
  }
  if (pickupAddresses.value.length > 0 && dropoffBranches.value.length > 0) {
    form.value.mode = form.value.mode || "pickup";
  } else if (pickupAddresses.value.length > 0) {
    form.value.mode = "pickup";
  } else if (dropoffBranches.value.length > 0) {
    form.value.mode = "dropoff";
  } else {
    form.value.mode = "";
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
    return (
      form.value.sp.trim() !== "" &&
      form.value.tn.trim() !== "" &&
      form.value.pid.trim() !== ""
    );
  }
  if (platform === "shopee") {
    if (
      pickupAddresses.value.length === 0 &&
      dropoffBranches.value.length === 0
    ) {
      return form.value.tn.trim() !== "";
    }
    if (form.value.mode === "pickup") {
      if (!form.value.aid) return false;
      const selected = pickupAddresses.value.find(
        (address) => address.address_id === Number(form.value.aid),
      );
      if (selected?.time_slots?.length) {
        return form.value.ptid.trim() !== "";
      }
      return true;
    }
    if (form.value.mode === "dropoff") {
      return form.value.bid.toString().trim() !== "";
    }
    return false;
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
  emit("confirm", {
    order_no: props.order.order_no,
    shipping_provider: form.value.sp.trim(),
    tracking_number: form.value.tn.trim() || undefined,
    address_id: form.value.aid ? Number(form.value.aid) : undefined,
    pickup_time_id: form.value.ptid || undefined,
    branch_id: form.value.bid ? Number(form.value.bid) : undefined,
    package_id: form.value.pid.trim() || undefined,
  });
};

watch(
  () => props.visible,
  (v) => {
    if (v) {
      resetForm();
      if (props.order?.package_id) {
        form.value.pid = props.order.package_id;
      }
      if (props.order?.platform?.toLowerCase() === "lazada") {
        form.value.sp =
          props.order.shipping_carrier || props.order.courier || "";
      }
      loadShippingParams();
    }
  },
);
</script>

<style scoped>
/* OrderShipModal - Unified Tokopedia Style */

.o {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 16px;
  animation: om-fadeIn 200ms ease-out;
}

.om {
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  animation: om-slideUp 250ms cubic-bezier(0.4, 0, 0.2, 1);
}

.h {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

.h h3 {
  font-size: 1.125rem;
  font-weight: 700;
  color: #212121;
  margin: 0;
}

.h button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
  color: #6c727c;
  border-radius: 6px;
  transition: all 150ms ease;
}

.h button:hover {
  background: #f3f4f5;
  color: #212121;
}

.h button:focus-visible {
  outline: 2px solid #03ac0e;
  outline-offset: 2px;
}

.b {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
  background: #ffffff;
}

.footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #e0e0e0;
  background: #ffffff;
  flex-shrink: 0;
}

.om-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 10px 20px;
  font-size: 0.875rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 150ms ease;
  border: 1px solid #e0e0e0;
  background: #ffffff;
  color: #212121;
}

.om-btn:hover:not(:disabled) {
  background: #f5f5f5;
  border-color: #bdbdbd;
}

.om-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.om-btn:focus-visible {
  outline: 2px solid #03ac0e;
  outline-offset: 2px;
}

.om-btn-secondary {
  background: #ffffff;
  border-color: #e0e0e0;
  color: #212121;
}

.om-btn-primary {
  background: #03ac0e;
  border-color: #03ac0e;
  color: #ffffff;
}

.om-btn-primary:hover:not(:disabled) {
  background: #029a0c;
  border-color: #029a0c;
}

/* Responsive */
@media (max-width: 640px) {
  .o {
    padding: 0;
  }

  .om {
    max-width: 100%;
    max-height: 100%;
    height: 100%;
    border-radius: 0;
  }

  .h {
    padding: 12px 16px;
  }

  .b {
    padding: 16px;
  }

  .footer {
    padding: 12px 16px;
    flex-direction: column;
  }

  .footer .om-btn {
    width: 100%;
  }
}

/* Accessibility */
@media (prefers-reduced-motion: reduce) {
  .o,
  .om {
    animation: none;
  }
}
</style>
