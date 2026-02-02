<template>
  <Teleport to="body"
    ><div v-if="visible" class="o" @click.self="close">
      <div class="om">
        <div class="h">
          <h3>Cancel Order</h3>
          <button @click="close"><i class="pi pi-times"></i></button>
        </div>
        <div class="b">
          <!-- Warning Alert -->
          <div class="warning-box">
            <i class="pi pi-exclamation-triangle"></i>
            <span
              >This action cannot be undone. Please confirm you want to cancel
              this order.</span
            >
          </div>

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
              <label for="cr"
                >Cancellation Reason
                <span style="color: var(--om-status-cancelled)">*</span></label
              ><select
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
              <label for="rd"
                >Additional Details
                <span style="font-weight: 400; color: var(--om-text-secondary)"
                  >(Optional)</span
                ></label
              ><textarea
                id="rd"
                v-model="form.rd"
                class="textarea"
                placeholder="Provide additional details..."
                rows="3"
                :disabled="loading"
              ></textarea>
            </div>
            <div class="confirm-checkbox">
              <label
                ><input
                  type="checkbox"
                  v-model="form.ok"
                  :disabled="loading"
                /><span
                  >I confirm that I want to cancel order
                  <strong>{{ order?.order_no }}</strong></span
                ></label
              >
            </div>

            <!-- Error Message -->
            <div v-if="error" class="error-box">
              <i class="pi pi-exclamation-circle"></i><span>{{ error }}</span>
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
    </div></Teleport
  >
</template>

<script setup lang="ts">
import { computed, watch } from "vue";
import {
  useModalForm,
  platformFormatter,
  getCancellationReasons,
  type CancellationReason,
} from "./composables/useModalForm";

interface Order {
  order_no: string;
  platform: string;
  buyer_username: string;
  [k: string]: any;
}

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
  background: 0;
  border: 0;
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

.warning-box {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 1rem;
  background: #fff3e0;
  border: 1px solid #ffb74d;
  border-radius: 8px;
  margin-bottom: 1rem;
  color: #e65100;
  font-size: 0.875rem;
}

.warning-box i {
  font-size: 1.25rem;
  flex-shrink: 0;
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
.textarea {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  font-family: inherit;
  transition: border-color var(--om-transition-fast);
}

.textarea {
  resize: vertical;
  min-height: 80px;
}

.select:focus,
.textarea:focus {
  outline: 0;
  border-color: var(--om-primary);
}

.select:disabled,
.textarea:disabled {
  background: var(--om-bg-secondary);
  cursor: not-allowed;
}

.confirm-checkbox {
  margin-top: 0.5rem;
}

.confirm-checkbox label {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--om-text-secondary);
  font-weight: normal;
}

.confirm-checkbox input {
  margin-top: 2px;
  accent-color: var(--om-primary);
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

.cancel-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  font-size: var(--om-font-sm);
  font-weight: 500;
  border-radius: var(--om-radius-sm);
  cursor: pointer;
  transition: all var(--om-transition-fast);
  border: 0;
  background: var(--om-status-cancelled);
  color: #fff;
}

.cancel-btn:hover:not(:disabled) {
  background: #c62828;
}

.cancel-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .om {
    margin: var(--om-spacing-md);
    max-height: calc(100vh - 2rem);
  }
  .footer {
    flex-direction: column;
  }
  .footer .om-btn,
  .footer .cancel-btn {
    width: 100%;
  }
}
</style>
