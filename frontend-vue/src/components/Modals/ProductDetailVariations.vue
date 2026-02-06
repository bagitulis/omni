<template>
  <div v-if="variations.length > 0" class="variations-section">
    <h4 class="section-title">
      <i class="pi pi-sliders-h"></i>
      Product Variations ({{ variations.length }})
    </h4>
    <div class="variations-list">
      <div v-for="variation in variations" :key="variation.name" class="variation-group">
        <h5>{{ variation.name }}</h5>
        <div class="option-list">
          <span v-for="option in variation.options" :key="option" class="option-item">
            {{ option }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Variation {
  name: string;
  options: string[];
}

interface Props {
  variations: Variation[];
}

defineProps<Props>();
</script>

<style scoped>
.variations-section {
  margin-bottom: 20px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #2c3e50;
  font-size: 1.1rem;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 2px solid #3498db;
}

.variations-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 15px;
}

.variation-group {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 6px;
  border: 1px solid #ecf0f1;
}

.variation-group h5 {
  margin: 0 0 10px 0;
  color: #2c3e50;
  font-size: 0.95rem;
  font-weight: 600;
}

.option-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.option-item {
  background: white;
  color: #2c3e50;
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid #bdc3c7;
  font-size: 0.85rem;
}

@media (max-width: 768px) {
  .variations-list {
    grid-template-columns: 1fr;
  }
}
</style>
