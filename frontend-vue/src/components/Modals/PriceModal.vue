<template>
  <Modal
    :is-open="visible"
    :title="`Update Prices - ${
      platform ? platform.charAt(0).toUpperCase() + platform.slice(1) : ''
    }`"
    @close="$emit('update:visible', false)"
  >
    <div class="space-y-5">
      <!-- Info Alert -->
      <div class="alert alert-info shadow-sm">
        <span class="text-sm leading-tight"
          >🏷️ Select the type of price update you want to perform.</span
        >
      </div>

      <!-- Price Options -->
      <div class="space-y-3">
        <!-- Regular Price Option -->
        <button
          @click="$emit('select-price-option', 'regular')"
          :class="[
            'w-full p-4 rounded-lg border-2 transition-all text-left',
            currentPriceOption === 'regular'
              ? 'border-primary bg-primary/10'
              : 'border-base-300 hover:border-primary/50',
          ]"
        >
          <div class="flex items-center gap-4">
            <div
              :class="[
                'w-12 h-12 rounded-lg flex items-center justify-center text-xl transition-all',
                currentPriceOption === 'regular'
                  ? 'bg-primary text-primary-content'
                  : 'bg-base-200 text-primary',
              ]"
            >
              🏷️
            </div>
            <div class="flex-1">
              <h4 class="font-semibold text-base-content">Regular Price</h4>
              <p class="text-sm text-base-content/70">
                Update regular product prices
              </p>
            </div>
            <div
              v-if="currentPriceOption === 'regular'"
              class="badge badge-primary"
            >
              ✓
            </div>
          </div>
        </button>

        <!-- Wholesale Price Option -->
        <button
          @click="$emit('select-price-option', 'wholesale')"
          :class="[
            'w-full p-4 rounded-lg border-2 transition-all text-left',
            currentPriceOption === 'wholesale'
              ? 'border-primary bg-primary/10'
              : 'border-base-300 hover:border-primary/50',
          ]"
        >
          <div class="flex items-center gap-4">
            <div
              :class="[
                'w-12 h-12 rounded-lg flex items-center justify-center text-xl transition-all',
                currentPriceOption === 'wholesale'
                  ? 'bg-primary text-primary-content'
                  : 'bg-base-200 text-primary',
              ]"
            >
              📦
            </div>
            <div class="flex-1">
              <h4 class="font-semibold text-base-content">Wholesale Price</h4>
              <p class="text-sm text-base-content/70">
                Update wholesale price tiers
              </p>
            </div>
            <div
              v-if="currentPriceOption === 'wholesale'"
              class="badge badge-primary"
            >
              ✓
            </div>
          </div>
        </button>
      </div>

      <!-- Instructions -->
      <div v-if="currentPriceOption" class="alert alert-warning shadow-sm">
        <div>
          <h4 class="font-semibold text-sm mb-2">Instructions:</h4>
          <p class="text-sm leading-tight">
            {{ priceInstructions }}
          </p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex gap-2 justify-end">
        <Button
          label="Cancel"
          color="ghost"
          size="sm"
          @click="$emit('update:visible', false)"
        />
        <Button
          label="Execute Update"
          color="primary"
          size="sm"
          :disabled="!currentPriceOption"
          @click="$emit('execute-price-update')"
        />
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import Modal from "@/components/Modal.vue";
import Button from "@/components/Button.vue";

interface Props {
  visible?: boolean;
  platform?: string | null;
  currentPriceOption?: string | null;
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  platform: null,
  currentPriceOption: null,
});

defineEmits<{
  close: [];
  "select-price-option": [option: string];
  "execute-price-update": [];
  "update:visible": [value: boolean];
}>();

const currentPriceOption = ref<string | null>(props.currentPriceOption);

const priceInstructions = computed((): string => {
  const instructions: Record<string, string> = {
    regular:
      "Update regular prices for all checked products in the sheet. Make sure the Google Sheet is prepared correctly with product IDs and new prices.",
    wholesale:
      "Update wholesale price tiers. Ensure Min_Order1, Price_Order1, Max_Order1 columns are filled. The system will automatically delete old tiers before adding new ones.",
  };
  return instructions[currentPriceOption.value || ""] || "";
});
</script>
