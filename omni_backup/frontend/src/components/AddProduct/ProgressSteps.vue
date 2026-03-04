<template>
  <div class="progress-steps">
    <div 
      v-for="(step, index) in steps" 
      :key="step.id"
      :class="['step', { active: currentStep === index, completed: currentStep > index }]"
    >
      <div class="step-number">{{ index + 1 }}</div>
      <div class="step-label">{{ step.label }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Step {
  id: string;
  label: string;
}

defineProps<{
  steps: Step[];
  currentStep: number;
}>();
</script>

<style scoped>
.progress-steps {
  display: flex;
  justify-content: center;
  padding: 16px;
  background: #f9fafb;
  gap: 40px;
  border-bottom: 1px solid #e5e7eb;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  opacity: 0.5;
  transition: opacity 0.3s;
}

.step.active,
.step.completed {
  opacity: 1;
}

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #e5e7eb;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  transition: all 0.3s;
}

.step.active .step-number {
  background: #1d4ed8;
  color: white;
}

.step.completed .step-number {
  background: #10b981;
  color: white;
}

.step-label {
  font-size: 12px;
  color: #6b7280;
  font-weight: 500;
}

@media (max-width: 640px) {
  .progress-steps {
    gap: 20px;
    padding: 12px;
  }

  .step-label {
    font-size: 10px;
  }
}
</style>
