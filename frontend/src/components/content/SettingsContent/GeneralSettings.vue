<template>
  <div class="general-settings">
    <h2>⚙️ General Settings</h2>
    <div class="settings-placeholder">
      <p>General settings will be displayed here</p>
    </div>
  </div>
</template>

<script setup lang="ts">
// General settings logic can be added here
</script>

<style scoped>
.general-settings {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.general-settings h2 {
  margin: 0;
  color: #333;
}

.settings-placeholder {
  padding: 48px 24px;
  text-align: center;
  background: #f9fafb;
  border-radius: 8px;
  color: #6b7280;
}

.settings-placeholder p {
  margin: 0;
  font-size: 1em;
}
</style>
