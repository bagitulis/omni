<template>
  <Teleport to="body">
    <div v-if="isOpen" class="modal-overlay" @click.self="$emit('close')">
      <div
        class="modal-content"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
      >
        <h4 :id="titleId">Link SKU to {{ platform }}</h4>

        <div class="form-group">
          <label>Platform Item ID</label>
          <input
            v-model="form.platformItemId"
            placeholder="e.g., 12345678"
            ref="itemIdInput"
          />
        </div>

        <div class="form-group">
          <label>Platform SKU ID (optional)</label>
          <input v-model="form.platformSkuId" placeholder="e.g., sku_12345" />
        </div>

        <div class="modal-actions">
          <button @click="$emit('close')" class="btn-cancel">Cancel</button>
          <button
            @click="handleSubmit"
            class="btn-submit"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? "Linking..." : "Link" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from "vue";

interface Props {
  isOpen: boolean;
  platform: string;
  isSubmitting: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  (e: "close"): void;
  (
    e: "submit",
    payload: { platformItemId: string; platformSkuId: string },
  ): void;
}>();

const form = ref({
  platformItemId: "",
  platformSkuId: "",
});

const itemIdInput = ref<HTMLInputElement | null>(null);
const titleId = `modal-title-${Math.random().toString(36).substr(2, 9)}`;

watch(
  () => props.isOpen,
  (newVal) => {
    if (newVal) {
      document.body.style.overflow = "hidden";
      form.value = { platformItemId: "", platformSkuId: "" };
      nextTick(() => {
        itemIdInput.value?.focus();
      });
    } else {
      document.body.style.overflow = "";
    }
  },
);

const handleSubmit = () => {
  if (!form.value.platformItemId) return;
  emit("submit", { ...form.value });
};

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === "Escape" && props.isOpen) {
    emit("close");
  }
};

onMounted(() => {
  document.addEventListener("keydown", handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", handleKeydown);
  document.body.style.overflow = "";
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  padding: 1.5rem;
  border-radius: 0.5rem;
  width: 100%;
  max-width: 400px;
}

.modal-content h4 {
  margin: 0 0 1.5rem;
  text-transform: capitalize;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
  margin-top: 1.5rem;
}

.btn-cancel {
  padding: 0.5rem 1rem;
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 0.375rem;
  cursor: pointer;
}

.btn-submit {
  padding: 0.5rem 1rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
}

.btn-submit:disabled {
  opacity: 0.6;
}
</style>
