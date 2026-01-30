<template>
  <div class="master-product-import">
    <div class="page-header">
      <router-link to="/master-products" class="back-link">
        ← Kembali ke Daftar
      </router-link>
      <h1>Import Produk dari Shopee</h1>
    </div>

    <div class="import-container">
      <ImportSelector />
    </div>
  </div>
</template>

<script setup lang="ts">
import ImportSelector from "@/components/MasterProduct/ImportSelector.vue";
</script>

<style scoped>
.master-product-import {
  padding: 1.5rem;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.page-header {
  margin-bottom: 1.5rem;
}

.back-link {
  display: inline-block;
  color: var(--text-secondary, #6b7280);
  text-decoration: none;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  transition: color 0.2s;
}

.back-link:hover {
  color: var(--primary, #3b82f6);
}

.page-header h1 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary, #1a1a1a);
  margin: 0;
}

.import-container {
  flex: 1;
  min-height: 0;
  background: var(--surface, white);
  border-radius: 0.5rem;
  border: 1px solid var(--border, #e5e7eb);
  padding: 1.5rem;
  overflow: hidden;
}
</style>
