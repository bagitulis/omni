<template>
  <form @submit.prevent="$emit('submit')" class="form">
    <div>
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

    <div>
      <label for="tn">
        Tracking Number
        <span style="font-weight: 400; color: var(--om-text-secondary)"
          >(Optional)</span
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

    <div
      v-if="
        order?.platform?.toLowerCase() === 'shopee' && pickupAddresses.length
      "
    >
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

    <div v-if="error" class="error-box">
      <Icon name="warning" size="sm" />
      <span>{{ error }}</span>
    </div>
  </form>
</template>

<script setup lang="ts">
import type { ShippingProvider } from "../composables/useModalForm";
import Icon from "@/components/ui/Icon.vue";

interface Props {
  modelValue: { sp: string; tn: string; aid: string | number };
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
</style>
