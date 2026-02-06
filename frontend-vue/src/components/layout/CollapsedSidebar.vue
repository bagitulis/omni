<template>
  <div class="collapsed-view">
    <button
      @click="$emit('toggle')"
      title="Expand Sidebar"
      aria-label="Expand Sidebar"
      class="collapsed-item"
      type="button"
    >
      <span aria-hidden="true">📊</span>
    </button>
    <button
      @click="$emit('load-status')"
      title="Refresh Status"
      aria-label="Refresh Status"
      class="collapsed-item"
      type="button"
    >
      <span aria-hidden="true">🔄</span>
    </button>
    <button
      @click="$emit('show-token-modal', 'shopee')"
      title="Token Management"
      aria-label="Token Management"
      class="collapsed-item"
      type="button"
    >
      <span aria-hidden="true">🔑</span>
    </button>
  </div>
</template>

<script lang="ts">
export default {
  name: "CollapsedSidebar",
  emits: ["toggle", "load-status", "show-token-modal"],
};
</script>

<style scoped>
.collapsed-view {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
  gap: 6px;
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}

.collapsed-item {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-radius: 6px;
  color: hsl(var(--bc));
  transition: all 0.2s ease;
  position: relative;
  text-decoration: none;
  border: none;
  background: linear-gradient(
    135deg,
    hsl(var(--b2)) 0%,
    hsl(var(--b2) / 0.8) 100%
  );
  box-shadow:
    0 2px 6px rgba(0, 0, 0, 0.08),
    inset 0 1px 2px rgba(255, 255, 255, 0.1);
  font-size: 1.2rem;
}

.collapsed-item:hover {
  background: linear-gradient(
    135deg,
    hsl(var(--pf) / 0.15) 0%,
    hsl(var(--p) / 0.1) 100%
  );
  color: hsl(var(--p));
  transform: translateY(-2px);
  box-shadow:
    0 4px 12px rgba(59, 130, 246, 0.15),
    inset 0 1px 2px rgba(255, 255, 255, 0.1);
}

.collapsed-item:active {
  transform: translateY(0);
}
</style>
