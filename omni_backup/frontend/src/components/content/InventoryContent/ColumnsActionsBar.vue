<template>
  <div class="actions-group">
    <div class="action-section">
      <label class="section-label">Visibility:</label>
      <div class="button-group">
        <button @click="$emit('select-all')" class="btn-small">
          Select All
        </button>
        <button @click="$emit('clear-all')" class="btn-small btn-secondary">
          Clear All
        </button>
      </div>
    </div>

    <div class="action-section">
      <label class="section-label">Lock:</label>
      <div class="button-group">
        <button @click="$emit('lock-all')" class="btn-lock">Lock All</button>
        <button @click="$emit('unlock-all')" class="btn-unlock">
          Unlock All
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  "select-all": [];
  "clear-all": [];
  "lock-all": [];
  "unlock-all": [];
}>();
</script>

<style scoped>
.actions-group {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  padding: 12px;
  background-color: #f9f9f9;
  border-radius: 4px;
  border: 1px solid #e8e8e8;
  width: 100%;
  box-sizing: border-box;
}

.action-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.section-label {
  font-size: 11px;
  font-weight: 600;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.button-group {
  display: flex;
  gap: 6px;
}

.btn-small {
  padding: 6px 10px;
  font-size: 11px;
  border: 1px solid #ddd;
  border-radius: 4px;
  background-color: white;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-small:hover {
  background-color: #f5f5f5;
  border-color: #6b7280;
}

.btn-secondary {
  color: #6b7280;
}

.btn-lock,
.btn-unlock {
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid #ddd;
  background-color: white;
  white-space: nowrap;
}

.btn-lock {
  border-color: #e74c3c;
  color: #e74c3c;
}

.btn-lock:hover {
  background-color: #ffe0e0;
}

.btn-unlock {
  border-color: #27ae60;
  color: #27ae60;
}

.btn-unlock:hover {
  background-color: #e8f8f5;
}
</style>
