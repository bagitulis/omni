<template>
  <Teleport to="body">
    <div v-if="visible" class="o" @click.self="close">
      <div class="om">
        <div class="h">
          <h3>Cancel Order</h3>
          <button @click="close"><i class="pi pi-times"></i></button>
        </div>
        <div class="b">
          <OrderCancelWarning />

          <OrderCancelInfo :order="order" />

          <form @submit.prevent="submit" class="form">
            <div>
              <label for="cr">
                Cancellation Reason
                <span style="color: var(--om-status-cancelled)">*</span>
              </label>
              <select
                id="cr"
                v-model="form.cr"
                class="select"
                :disabled="loading"
                required
              >
                <option value="">Select a reason...</option>
                <option v-for="r in reasons" :key="r.value" :value="r.value">
                  {{ r.label }}
                </option>
              </select>
            </div>
            <div>
              <label for="rd">
                Additional Details
                <span style="font-weight: 400; color: var(--om-text-secondary)">
                  (Optional)
                </span>
              </label>
              <textarea
                id="rd"
                v-model="form.rd"
                class="textarea"
                placeholder="Provide additional details..."
                rows="3"
                :disabled="loading"
              ></textarea>
            </div>
            <div class="confirm-checkbox">
              <label>
                <input type="checkbox" v-model="form.ok" :disabled="loading" />
                <span>
                  I confirm that I want to cancel order
                  <strong>{{ order?.order_no }}</strong>
                </span>
              </label>
            </div>

            <div v-if="error" class="error-box">
              <i class="pi pi-exclamation-circle"></i><span>{{ error }}</span>
            </div>
          </form>
        </div>

        <div class="footer">
          <button
            @click="close"
            class="om-btn om-btn-secondary"
            :disabled="loading"
          >
            Go Back
          </button>
          <button
            @click="submit"
            class="cancel-btn"
            :disabled="loading || !valid"
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
import { computed, watch } from "vue";
import {
  useModalForm,
  getCancellationReasons,
  type CancellationReason,
} from "./composables/useModalForm";
import type { Order } from "./composables/useOrderManager";
import OrderCancelWarning from "./OrderCancelWarning.vue";
import OrderCancelInfo from "./OrderCancelInfo.vue";

const props = defineProps<{ visible: boolean; order: Order | null }>();
const emit = defineEmits<{
  close: [];
  confirm: [
    { order_no: string; cancel_reason: string; reason_detail?: string },
  ];
}>();

const { form, loading, error, resetForm } = useModalForm({
  cr: "",
  rd: "",
  ok: false,
});

const reasons = computed<CancellationReason[]>(() =>
  getCancellationReasons(props.order?.platform),
);

const valid = computed(() => form.value.cr !== "" && form.value.ok);

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
    cancel_reason: form.value.cr,
    reason_detail: form.value.rd || undefined,
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
@import "./OrderCancelModal.styles.css";
</style>
