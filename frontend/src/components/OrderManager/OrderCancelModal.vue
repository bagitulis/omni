<template>
  <Teleport to="body"
    ><div v-if="visible" class="o" @click.self="close">
      <div class="om">
        <div class="h">
          <h3>Cancel Order</h3>
          <button @click="close"><i class="pi pi-times"></i></button>
        </div>
        <div class="b">
          <div
            style="
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
            "
          >
            <i
              class="pi pi-exclamation-triangle"
              style="font-size: 1.25rem; flex-shrink: 0"
            ></i
            ><span
              >This action cannot be undone. Please confirm you want to cancel
              this order.</span
            >
          </div>
          <div class="i">
            <div class="r">
              <span>Order No.</span><span>{{ order?.order_no }}</span>
            </div>
            <div class="r">
              <span>Platform</span
              ><span
                :class="[
                  'om-platform-badge',
                  `om-platform-${order?.platform?.toLowerCase()}`,
                ]"
                >{{ fmt(order?.platform) }}</span
              >
            </div>
            <div class="r">
              <span>Buyer</span><span>{{ order?.buyer_username }}</span>
            </div>
          </div>
          <form @submit.prevent="submit" class="f">
            <div>
              <label for="cr"
                >Cancellation Reason
                <span style="color: var(--om-status-cancelled)">*</span></label
              ><select
                id="cr"
                v-model="f.cr"
                class="s"
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
                v-model="f.rd"
                class="ta"
                placeholder="Provide additional details..."
                rows="3"
                :disabled="loading"
              ></textarea>
            </div>
            <div style="margin-top: 0.5rem">
              <label
                style="
                  display: flex;
                  align-items: flex-start;
                  gap: 0.5rem;
                  cursor: pointer;
                  font-size: 0.875rem;
                  color: var(--om-text-secondary);
                "
                ><input
                  type="checkbox"
                  v-model="f.ok"
                  :disabled="loading"
                  style="margin-top: 2px; accent-color: var(--om-primary)"
                /><span
                  >I confirm that I want to cancel order
                  <strong>{{ order?.order_no }}</strong></span
                ></label
              >
            </div>
            <div
              v-if="error"
              style="
                display: flex;
                align-items: center;
                gap: 0.5rem;
                padding: 0.5rem;
                background: #ffebee;
                color: var(--om-status-cancelled);
                border-radius: 4px;
                font-size: 0.875rem;
              "
            >
              <i class="pi pi-exclamation-circle"></i><span>{{ error }}</span>
            </div>
          </form>
        </div>
        <div class="ft">
          <button
            @click="close"
            class="om-btn om-btn-secondary"
            :disabled="loading"
          >
            Go Back</button
          ><button @click="submit" class="cb" :disabled="loading || !valid">
            <i v-if="loading" class="pi pi-spin pi-spinner"></i
            ><i v-else class="pi pi-times"></i
            ><span>{{ loading ? "Processing..." : "Cancel Order" }}</span>
          </button>
        </div>
      </div>
    </div></Teleport
  >
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
interface O {
  order_no: string;
  platform: string;
  buyer_username: string;
  [k: string]: any;
}
interface R {
  value: string;
  label: string;
}
const props = defineProps<{ visible: boolean; order: O | null }>();
const emit = defineEmits<{
  close: [];
  confirm: [
    { order_no: string; cancel_reason: string; reason_detail?: string },
  ];
}>();
const f = ref({ cr: "", rd: "", ok: false }),
  loading = ref(false),
  error = ref("");
const reasons = computed<R[]>(() => {
    const p = props.order?.platform?.toLowerCase(),
      c = [
        { value: "OUT_OF_STOCK", label: "Out of stock" },
        { value: "BUYER_REQUEST", label: "Buyer requested cancellation" },
        { value: "WRONG_PRICE", label: "Wrong price/listing error" },
        { value: "DUPLICATE_ORDER", label: "Duplicate order" },
        { value: "OTHER", label: "Other reason" },
      ];
    if (p === "shopee")
      return [
        { value: "CUSTOMER_REQUEST", label: "Customer requested cancellation" },
        { value: "OUT_OF_STOCK", label: "Out of stock" },
        { value: "UNDELIVERABLE_AREA", label: "Undeliverable area" },
        { value: "COD_NOT_SUPPORTED", label: "COD not supported" },
        ...c.filter(
          (r) => !["OUT_OF_STOCK", "BUYER_REQUEST"].includes(r.value),
        ),
      ];
    if (p === "lazada")
      return [
        { value: "customer_request", label: "Customer requested cancellation" },
        { value: "out_of_stock", label: "Out of stock" },
        { value: "sourcing_failed", label: "Sourcing failed" },
        ...c.filter(
          (r) => !["OUT_OF_STOCK", "BUYER_REQUEST"].includes(r.value),
        ),
      ];
    return c;
  }),
  valid = computed(() => f.value.cr !== "" && f.value.ok),
  fmt = (p?: string): string => {
    const m: Record<string, string> = {
      shopee: "Shopee",
      lazada: "Lazada",
      tiktok: "TikTok",
    };
    return m[p?.toLowerCase() || ""] || p || "";
  };
const close = () => {
  if (!loading.value) {
    f.value = { cr: "", rd: "", ok: false };
    error.value = "";
    emit("close");
  }
};
const submit = async () => {
  if (!valid.value || !props.order) return;
  loading.value = true;
  error.value = "";
  try {
    emit("confirm", {
      order_no: props.order.order_no,
      cancel_reason: f.value.cr,
      reason_detail: f.value.rd || undefined,
    });
  } catch (e: any) {
    error.value = e.message || "Failed";
  } finally {
    loading.value = false;
  }
};
watch(
  () => props.visible,
  (v) => {
    if (v) {
      f.value = { cr: "", rd: "", ok: false };
      error.value = "";
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
.i {
  background: var(--om-bg-secondary);
  border-radius: var(--om-radius-md);
  padding: var(--om-spacing-md);
  margin-bottom: var(--om-spacing-lg);
}
.r {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--om-spacing-xs) 0;
  font-size: var(--om-font-sm);
}
.r:not(:last-child) {
  border-bottom: 1px solid var(--om-border);
  padding-bottom: var(--om-spacing-sm);
  margin-bottom: var(--om-spacing-sm);
}
.r span:first-child {
  color: var(--om-text-secondary);
}
.r span:last-child {
  font-weight: 500;
  color: var(--om-text-primary);
}
.f {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-md);
}
.f > div {
  display: flex;
  flex-direction: column;
  gap: var(--om-spacing-xs);
}
.f label {
  font-size: var(--om-font-sm);
  font-weight: 500;
  color: var(--om-text-primary);
}
.s,
.ta {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  font-family: inherit;
  transition: border-color var(--om-transition-fast);
}
.ta {
  resize: vertical;
  min-height: 80px;
}
.s:focus,
.ta:focus {
  outline: 0;
  border-color: var(--om-primary);
}
.s:disabled,
.ta:disabled {
  background: var(--om-bg-secondary);
  cursor: not-allowed;
}
.ft {
  display: flex;
  justify-content: flex-end;
  gap: var(--om-spacing-sm);
  padding: var(--om-spacing-md) var(--om-spacing-lg);
  border-top: 1px solid var(--om-border);
}
.cb {
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
.cb:hover:not(:disabled) {
  background: #c62828;
}
.cb:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
@media (max-width: 768px) {
  .om {
    margin: var(--om-spacing-md);
    max-height: calc(100vh - 2rem);
  }
  .ft {
    flex-direction: column;
  }
  .ft .om-btn,
  .ft .cb {
    width: 100%;
  }
}
</style>
