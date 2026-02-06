<template>
  <div class="options-container">
    <div class="options-intro">
      <h3 class="text-lg font-bold text-gray-800 mb-2">Select an Option</h3>
      <p class="text-sm text-gray-600">Choose how to process shipping fees</p>
    </div>

    <div class="space-y-3">
      <button
        @click="$emit('select', 1)"
        class="option-button option-wallet"
      >
        <div class="option-icon">📤</div>
        <div class="option-content">
          <h4 class="option-title">Get Order Numbers from Wallet</h4>
          <p class="option-description">
            Automatically extract order numbers from wallet transactions
          </p>
        </div>
        <div class="option-arrow">→</div>
      </button>

      <button
        @click="$emit('select', 2)"
        class="option-button option-files"
      >
        <div class="option-icon">📥</div>
        <div class="option-content">
          <h4 class="option-title">Process Existing Files</h4>
          <p class="option-description">
            Process existing order number files in the system
          </p>
        </div>
        <div class="option-arrow">→</div>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  select: [option: number];
}>();
</script>

<style scoped>
.options-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.options-intro {
  padding: 1rem;
  background: linear-gradient(135deg, #f0f4f8 0%, #f8fafc 100%);
  border-radius: 0.75rem;
  border-left: 4px solid #3b82f6;
}

.options-intro h3 {
  margin: 0;
  color: #1f2937;
  font-size: 1.125rem;
}

.options-intro p {
  margin: 0;
  color: #6b7280;
  font-size: 0.875rem;
}

.option-button {
  display: flex;
  align-items: center;
  gap: 1rem;
  width: 100%;
  padding: 1.25rem;
  background: white;
  border: 2px solid #e5e7eb;
  border-radius: 0.75rem;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: left;
}

.option-button:hover {
  border-color: #3b82f6;
  background: #f0f9ff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  transform: translateY(-2px);
}

.option-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 3rem;
  height: 3rem;
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  border-radius: 0.5rem;
  font-size: 1.5rem;
  flex-shrink: 0;
}

.option-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.option-title {
  margin: 0;
  color: #1f2937;
  font-weight: 600;
  font-size: 0.95rem;
}

.option-description {
  margin: 0;
  color: #6b7280;
  font-size: 0.8rem;
}

.option-arrow {
  color: #9ca3af;
  font-size: 1.25rem;
  transition: color 0.3s ease;
}

.option-button:hover .option-arrow {
  color: #3b82f6;
}
</style>
