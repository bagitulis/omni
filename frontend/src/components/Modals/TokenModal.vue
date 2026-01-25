<template>
  <Modal
    :is-open="visible"
    :title="`${
      platform ? platform.charAt(0).toUpperCase() + platform.slice(1) : ''
    } Token Management`"
    @close="$emit('update:visible', false)"
  >
    <div class="space-y-5">
      <!-- Info Alert -->
      <div class="alert alert-info shadow-sm">
        <span class="text-sm leading-tight">
          Manage your {{ platform }} API tokens and authorization codes.
        </span>
      </div>

      <!-- Action Buttons -->
      <div class="space-y-2 flex flex-col">
        <Button
          label="📝 Update Authorization Code"
          color="secondary"
          fullWidth
          size="sm"
          @click="$emit('handle-token-operation', 'update_code')"
        />
        <Button
          label="🔑 Get Access Token"
          color="secondary"
          fullWidth
          size="sm"
          @click="$emit('handle-token-operation', 'get_token')"
        />
        <Button
          label="🔄 Refresh Token"
          color="secondary"
          fullWidth
          size="sm"
          @click="$emit('handle-token-operation', 'refresh_token')"
        />
      </div>
    </div>

    <template #footer>
      <div class="flex gap-2 justify-end">
        <Button label="Close" color="ghost" size="sm" @click="$emit('close')" />
      </div>
    </template>
  </Modal>
</template>

<script setup lang="ts">
import Modal from "@/components/Modal.vue";
import Button from "@/components/Button.vue";

defineProps<{
  visible: boolean;
  platform: string | null | undefined;
}>();

defineEmits<{
  close: [];
  "handle-token-operation": [operation: string];
  "update:visible": [value: boolean];
}>();
</script>
