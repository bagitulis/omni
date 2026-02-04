<template>
  <form @submit.prevent="$emit('submit')" class="form">
    <div v-if="showShopeeNote" class="info-note">
      Shipping is handled by Shopee. Select pickup or dropoff options if
      required.
    </div>

    <div v-if="showShippingModeSelect">
      <label for="mode">Shipping Method</label>
      <select
        id="mode"
        :value="modelValue.mode"
        @input="onInput('mode', $event)"
        class="select"
        :disabled="loading"
      >
        <option value="pickup">Pickup</option>
        <option value="dropoff">Dropoff</option>
      </select>
    </div>

    <div v-if="showLazadaProvider">
      <label for="sp">Shipping Provider</label>
      <input
        id="sp"
        :value="modelValue.sp"
        @input="onInput('sp', $event)"
        type="text"
        class="input"
        placeholder="Use the provider from Lazada shipment providers"
        :disabled="loading || lazadaProviderLocked"
      />
      <p v-if="showLazadaProviderHelp" class="helper-text">
        Provider must match the shipment providers for this order.
      </p>
    </div>

    <div v-if="showTiktokProvider">
      <label for="sp">Shipping Provider ID</label>
      <input
        id="sp"
        :value="modelValue.sp"
        @input="onInput('sp', $event)"
        type="text"
        class="input"
        placeholder="Enter TikTok shipping provider ID..."
        :disabled="loading"
      />
      <p v-if="showTiktokProviderHelp" class="helper-text">
        Use the shipping_provider_id from the TikTok Shipping Provider API.
      </p>
    </div>

    <div v-if="showTrackingInput">
      <label for="tn">
        Tracking Number
        <span style="font-weight: 400; color: var(--om-text-secondary)">
          ({{ trackingLabel }})</span
        >
      </label>
      <input
        id="tn"
        :value="modelValue.tn"
        @input="onInput('tn', $event)"
        type="text"
        class="input"
        placeholder="Enter tracking number..."
        :disabled="loading"
      />
    </div>

    <div v-if="showPackageId">
      <label for="pid">
        Package ID
        <span style="font-weight: 400; color: var(--om-text-secondary)">
          (Required)</span
        >
      </label>
      <input
        id="pid"
        :value="modelValue.pid"
        @input="onInput('pid', $event)"
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
        @input="onInput('aid', $event)"
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
        @input="onInput('ptid', $event)"
        class="select"
        :disabled="loading"
      >
        <option value="">Select Pickup Time...</option>
        <option v-for="t in pickupTimes" :key="t.value" :value="t.value">
          {{ t.label }}
        </option>
      </select>
    </div>

    <div v-if="showDropoffSelect">
      <label for="bd">Dropoff Branch</label>
      <select
        id="bd"
        :value="modelValue.bid"
        @input="onInput('bid', $event)"
        class="select"
        :disabled="loading"
      >
        <option value="">Select Dropoff Branch...</option>
        <option
          v-for="b in dropoffBranches"
          :key="b.branch_id"
          :value="b.branch_id"
        >
          {{ b.address }}
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
    bid?: string | number;
    mode?: string;
  };
  loading: boolean;
  error: string;
  pickupAddresses: any[];
  dropoffBranches: any[];
  order: any;
  providers: ShippingProvider[];
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "update:modelValue", value: any): void;
  (e: "submit"): void;
}>();

const platform = computed(() => props.order?.platform?.toLowerCase() || "");

const hasPickup = computed(() => props.pickupAddresses.length > 0);
const hasDropoff = computed(() => props.dropoffBranches.length > 0);

const showShopeeNote = computed(
  () => platform.value === "shopee" && (hasPickup.value || hasDropoff.value),
);

const showLazadaProvider = computed(() => platform.value === "lazada");
const lazadaProviderLocked = computed(
  () => platform.value === "lazada" && !!props.modelValue.sp,
);
const showLazadaProviderHelp = computed(
  () => showLazadaProvider.value && !props.modelValue.sp,
);

const showTiktokProvider = computed(() => platform.value === "tiktok");
const showTiktokProviderHelp = computed(() => platform.value === "tiktok");

const showShippingModeSelect = computed(
  () => platform.value === "shopee" && hasPickup.value && hasDropoff.value,
);

const showPickupSelect = computed(() => {
  if (platform.value !== "shopee") return false;
  if (!hasPickup.value) return false;
  if (hasDropoff.value && props.modelValue.mode === "dropoff") return false;
  return true;
});

const pickupTimes = computed(() => {
  const selected = props.pickupAddresses.find(
    (address) => address.address_id === Number(props.modelValue.aid),
  );
  const timeSlots = selected?.time_slots || selected?.time_slot_list;
  if (!timeSlots) {
    return [] as { value: string; label: string }[];
  }
  return timeSlots.map((slot: any) => ({
    value: slot.pickup_time_id || slot.time_slot || "",
    label: slot.pickup_time || slot.time || slot.time_slot || "",
  }));
});

const showPickupTimeSelect = computed(
  () => showPickupSelect.value && pickupTimes.value.length > 0,
);

const showDropoffSelect = computed(() => {
  if (platform.value !== "shopee") return false;
  if (!hasDropoff.value) return false;
  if (hasPickup.value && props.modelValue.mode !== "dropoff") return false;
  return true;
});

const trackingLabel = computed(() => {
  if (platform.value === "tiktok") return "Required";
  if (platform.value === "shopee" && !hasPickup.value && !hasDropoff.value) {
    return "Required";
  }
  return "Optional";
});

const showPackageId = computed(() => platform.value === "tiktok");

const showTrackingInput = computed(() => {
  if (platform.value === "tiktok" || platform.value === "lazada") return true;
  if (platform.value === "shopee" && !hasPickup.value && !hasDropoff.value) {
    return true;
  }
  return false;
});

const update = (key: string, value: any) => {
  emit("update:modelValue", { ...props.modelValue, [key]: value });
};

const onInput = (key: string, event: Event) => {
  const target = event.target as HTMLInputElement | HTMLSelectElement | null;
  update(key, target?.value ?? "");
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

.info-note {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border-radius: var(--om-radius-sm);
  background: var(--om-bg-secondary);
  border: 1px dashed var(--om-border);
  font-size: var(--om-font-sm);
  color: var(--om-text-secondary);
}
</style>
