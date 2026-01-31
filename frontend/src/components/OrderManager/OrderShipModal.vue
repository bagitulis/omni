<template>
  <Teleport to="body">
    <div v-if="visible" class="o" @click.self="close">
      <div class="om">
        <div class="h">
          <h3>Arrange Shipment</h3>
          <button @click="close"><i class="pi pi-times"></i></button>
        </div>
        <div class="b">
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
              <label for="sp">Shipping Provider</label
              ><select id="sp" v-model="f.sp" class="s" :disabled="loading">
                <option value="">Select Provider...</option>
                <option v-for="p in providers" :key="p.value" :value="p.value">
                  {{ p.label }}
                </option>
              </select>
            </div>
            <div>
              <label for="tn"
                >Tracking Number
                <span style="font-weight: 400; color: var(--om-text-secondary)"
                  >(Optional)</span
                ></label
              ><input
                id="tn"
                v-model="f.tn"
                type="text"
                class="x"
                placeholder="Enter tracking number..."
                :disabled="loading"
              />
            </div>
            <div
              v-if="
                order?.platform?.toLowerCase() === 'shopee' &&
                pickupAddresses.length
              "
            >
              <label for="pa">Pickup Address</label
              ><select id="pa" v-model="f.aid" class="s" :disabled="loading">
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
            Cancel</button
          ><button
            @click="submit"
            class="om-btn om-btn-primary"
            :disabled="loading || !valid"
          >
            <i v-if="loading" class="pi pi-spin pi-spinner"></i
            ><i v-else class="pi pi-truck"></i
            ><span>{{ loading ? "Processing..." : "Confirm Shipment" }}</span>
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
interface O {
  order_no: string;
  platform: string;
  buyer_username: string;
  [k: string]: any;
}
interface P {
  value: string;
  label: string;
}
interface A {
  address_id: number;
  address: string;
}
const props = defineProps<{ visible: boolean; order: O | null }>();
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
const f = ref({ sp: "", tn: "", aid: "" });
const loading = ref(false);
const error = ref("");
const providers = ref<P[]>([
  { value: "jne", label: "JNE" },
  { value: "jnt", label: "J&T Express" },
  { value: "sicepat", label: "SiCepat" },
  { value: "anteraja", label: "AnterAja" },
  { value: "ninja", label: "Ninja Van" },
  { value: "shopee_express", label: "Shopee Express" },
  { value: "lazada_logistics", label: "Lazada Logistics" },
]);
const pickupAddresses = ref<A[]>([]);
const valid = computed(() => f.value.sp !== "");
const fmt = (p?: string): string => {
  const m: Record<string, string> = {
    shopee: "Shopee",
    lazada: "Lazada",
    tiktok: "TikTok",
  };
  return m[p?.toLowerCase() || ""] || p || "";
};
const close = () => {
  if (!loading.value) {
    f.value = { sp: "", tn: "", aid: "" };
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
      shipping_provider: f.value.sp,
      tracking_number: f.value.tn || undefined,
      address_id: f.value.aid ? Number(f.value.aid) : undefined,
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
      f.value = { sp: "", tn: "", aid: "" };
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
.x {
  padding: var(--om-spacing-sm) var(--om-spacing-md);
  border: 1px solid var(--om-border);
  border-radius: var(--om-radius-sm);
  font-size: var(--om-font-sm);
  transition: border-color var(--om-transition-fast);
}
.s:focus,
.x:focus {
  outline: none;
  border-color: var(--om-primary);
}
.s:disabled,
.x:disabled {
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
@media (max-width: 768px) {
  .om {
    margin: var(--om-spacing-md);
    max-height: calc(100vh - 2rem);
  }
  .ft {
    flex-direction: column;
  }
  .ft .om-btn {
    width: 100%;
  }
}
</style>
