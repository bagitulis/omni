<template>
  <form @submit.prevent="$emit('submit')" class="form">
    <div v-if="showProviderSelect">
      <label for="sp">Shipping Provider</label>
      <select
        id="sp"
        :value="modelValue.sp"
        @input="update('sp', ($event.target as HTMLSelectElement).value)"
        class="select"
        :disabled="loading"
      >
        <option value="">Select Provider...</option>
        <option v-for="p in providers" :key="p.value" :value="p.value">
          {{ p.label }}
        </option>
      </select>
    </div>

    <div v-else>
      <label for="sp">Shipping Provider ID</label>
      <input
        id="sp"
        :value="modelValue.sp"
        @input="update('sp', ($event.target as HTMLInputElement).value)"
        type="text"
        class="input"
        placeholder="Enter TikTok shipping provider ID..."
        :disabled="loading"
      />
      <p v-if="showProviderHelp" class="helper-text">
        For TikTok use the shipping_provider_id from the Shipping Provider API.
      </p>
    </div>

    <div>
      <label for="tn">
        Tracking Number
        <span style="font-weight: 400; color: var(--om-text-secondary)"
          >({{ trackingLabel }})</span
        >
      </label>
      <input
        id="tn"
        :value="modelValue.tn"
        @input="update('tn', ($event.target as HTMLInputElement).value)"
        type="text"
        class="input"
        placeholder="Enter tracking number..."
        :disabled="loading"
      />
    </div>

    <div v-if="showPackageId">
      <label for="pid">
        Package ID
        <span style="font-weight: 400; color: var(--om-text-secondary)"
          >(Required)</span
        >
      </label>
      <input
        id="pid"
        :value="modelValue.pid"
        @input="update('pid', ($event.target as HTMLInputElement).value)"
        type="text"
        class="input"
        placeholder="Enter package ID..."
        :disabled="loading"
      />
    </div>

    <div v-if="showPickupSelect">
      <label for="pa">Pickup Address</label>
      <select
        id="pa"
        :value="modelValue.aid"
        @input="update('aid', ($event.target as HTMLSelectElement).value)"
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

    <div v-if="showPickupTimeSelect">
      <label for="pt">Pickup Time</label>
      <select
        id="pt"
        :value="modelValue.ptid"
        @input="update('ptid', ($event.target as HTMLSelectElement).value)"
        class="select"
        :disabled="loading"
      >
        <option value="">Select Pickup Time...</option>
        <option v-for="t in pickupTimes" :key="t.value" :value="t.value">
          {{ t.label }}
        </option>
      </select>
    </div>

    <div v-if="error" class="error-box">
      <Icon name="warning" size="sm" />
      <span>{{ error }}</span>
    </div>
  </form>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ShippingProvider } from "../composables/useModalForm";
import Icon from "@/components/ui/Icon.vue";

interface Props {
  modelValue: {
    sp: string;
    tn: string;
    aid: string | number;
    ptid?: string;
    pid?: string;
  };
  loading: boolean;
  error: string;
  pickupAddresses: any[];
  order: any;
  providers: ShippingProvider[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "update:modelValue", value: any): void;
  (e: "submit"): void;
}>();

const showProviderSelect = computed(() => {
  const platform = props.order?.platform?.toLowerCase();
  return platform !== "tiktok";
});

const showProviderHelp = computed(() => {
  const platform = props.order?.platform?.toLowerCase();
  return platform === "tiktok";
});

const showPickupSelect = computed(() => {
  const platform = props.order?.platform?.toLowerCase();
  return platform === "shopee" && props.pickupAddresses.length > 0;
});

const pickupTimes = computed(() => {
  const selected = props.pickupAddresses.find(
    (address) => address.address_id === Number(props.modelValue.aid),
  );
  if (!selected?.time_slots) {
    return [] as { value: string; label: string }[];
  }
  return selected.time_slots.map((slot: any) => ({
    value: slot.pickup_time_id || slot.time_slot || "",
    label: slot.pickup_time || slot.time || slot.time_slot || "",
  }));
});

const showPickupTimeSelect = computed(
  () => showPickupSelect.value && pickupTimes.value.length > 0,
);

const trackingLabel = computed(() => {
  const platform = props.order?.platform?.toLowerCase();
  return platform === "tiktok" ? "Required" : "Optional";
});

const showPackageId = computed(() => {
  const platform = props.order?.platform?.toLowerCase();
  return platform === "tiktok";
});

const update = (key: string, value: any) => {
  emit("update:modelValue", { ...props.modelValue, [key]: value });
};
</script>

<style scoped>
@import "../OrderManager.theme.css";

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

.helper-text {
  margin-top: 0.35rem;
  font-size: 0.75rem;
  color: var(--om-text-secondary);
}
</style>
