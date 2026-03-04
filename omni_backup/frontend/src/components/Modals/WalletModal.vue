<template>
  <Modal
    :is-open="visible"
    title="Wallet Transactions"
    @close="$emit('update:visible', false)"
  >
    <div class="space-y-5">
      <!-- Info Card -->
      <div class="alert alert-info shadow-sm">
        <div class="flex items-start gap-3">
          <span class="text-xl">💰</span>
          <p class="text-sm leading-tight">
            Download wallet transaction reports for the selected period.
          </p>
        </div>
      </div>

      <!-- Form Controls -->
      <div class="space-y-4">
        <!-- Month Select -->
        <div class="form-control">
          <label class="label" for="wallet-month-select">
            <span class="label-text font-medium">Month</span>
          </label>
          <select
            id="wallet-month-select"
            v-model="internalWalletParams.month"
            @change="updateWalletParams"
            class="select select-bordered select-sm w-full"
          >
            <option
              v-for="month in months"
              :key="month.value"
              :value="month.value"
            >
              {{ month.name }}
            </option>
          </select>
        </div>

        <!-- Year Select -->
        <div class="form-control">
          <label class="label" for="wallet-year-select">
            <span class="label-text font-medium">Year</span>
          </label>
          <select
            id="wallet-year-select"
            v-model="internalWalletParams.year"
            @change="updateWalletParams"
            class="select select-bordered select-sm w-full"
          >
            <option v-for="year in years" :key="year.value" :value="year.value">
              {{ year.name }}
            </option>
          </select>
        </div>

        <!-- Transaction Type Select -->
        <div class="form-control">
          <label class="label" for="wallet-transaction-type-select">
            <span class="label-text font-medium">Transaction Type</span>
          </label>
          <select
            id="wallet-transaction-type-select"
            v-model="internalWalletParams.transaction_type"
            @change="updateWalletParams"
            class="select select-bordered select-sm w-full"
          >
            <option
              v-for="type in transactionTypes"
              :key="type.value"
              :value="type.value"
            >
              {{ type.name }}
            </option>
          </select>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex gap-2 justify-end">
        <Button
          label="Cancel"
          color="ghost"
          size="sm"
          @click="$emit('close')"
        />
        <Button
          label="📊 Export to Google Sheets"
          color="primary"
          size="sm"
          @click="$emit('export-wallet-to-sheets')"
        />
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import Modal from "@/components/Modal.vue";
import Button from "@/components/Button.vue";

interface WalletParams {
  month: number;
  year: number;
  transaction_type: string;
}

interface SelectOption {
  name: string;
  value: string | number;
}

interface Props {
  visible?: boolean;
  walletParams?: WalletParams;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  walletParams: () => ({
    month: new Date().getMonth() + 1,
    year: new Date().getFullYear(),
    transaction_type: "wallet_order_income",
  }),
});

const emit = defineEmits<{
  close: [];
  "update-wallet-params": [params: WalletParams];
  "export-wallet-to-sheets": [];
}>();

const internalWalletParams = ref<WalletParams>({ ...props.walletParams });

const years = ref<SelectOption[]>(
  Array.from({ length: 11 }, (_, i) => ({
    name: (2020 + i).toString(),
    value: 2020 + i,
  }))
);

const months = ref<SelectOption[]>([
  { name: "January", value: 1 },
  { name: "February", value: 2 },
  { name: "March", value: 3 },
  { name: "April", value: 4 },
  { name: "May", value: 5 },
  { name: "June", value: 6 },
  { name: "July", value: 7 },
  { name: "August", value: 8 },
  { name: "September", value: 9 },
  { name: "October", value: 10 },
  { name: "November", value: 11 },
  { name: "December", value: 12 },
]);

const transactionTypes = ref<SelectOption[]>([
  { name: "Order Income", value: "wallet_order_income" },
  { name: "Adjustment", value: "wallet_adjustment_filter" },
  { name: "Wallet Payment", value: "wallet_wallet_payment" },
  { name: "Refund from Order", value: "wallet_refund_from_order" },
  { name: "Withdrawals", value: "wallet_withdrawals" },
]);

const updateWalletParams = (): void => {
  emit("update-wallet-params", internalWalletParams.value);
};

watch(
  () => props.walletParams,
  (newParams) => {
    if (newParams) {
      internalWalletParams.value = { ...newParams };
    }
  }
);
</script>
