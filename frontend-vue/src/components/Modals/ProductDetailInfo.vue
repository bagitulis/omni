<template>
  <div class="info-section">
    <h4 class="section-title">
      <i class="pi pi-box"></i> Product Information
    </h4>
    <div class="info-grid">
      <div class="info-item">
        <label>Item ID</label>
        <span class="value">{{ product.item_id }}</span>
      </div>
      <div class="info-item">
        <label>Item Name</label>
        <span class="value">{{ product.item_name || "-" }}</span>
      </div>
      <div class="info-item">
        <label>Item SKU</label>
        <span class="value sku">{{ product.item_sku || "-" }}</span>
      </div>
      <div class="info-item">
        <label>Shop ID</label>
        <span class="value">{{ product.shop_id || 0 }}</span>
      </div>
      <div class="info-item">
        <label>Current Price</label>
        <span class="value price">Rp {{ formatPrice(product.current_price) }}</span>
      </div>
      <div class="info-item">
        <label>Original Price</label>
        <span class="value price">Rp {{ formatPrice(product.original_price) }}</span>
      </div>
      <div class="info-item">
        <label>Seller Stock</label>
        <span class="value stock" :class="{ low: (product.seller_stock || 0) < 10 }">
          {{ product.seller_stock || 0 }}
        </span>
      </div>
      <div class="info-item">
        <label>Shopee Stock</label>
        <span class="value stock">{{ product.shopee_stock || 0 }}</span>
      </div>
      <div class="info-item">
        <label>Category ID</label>
        <span class="value">{{ product.category_id || "-" }}</span>
      </div>
      <div class="info-item">
        <label>Brand</label>
        <span class="value">{{ product.brand || "-" }}</span>
      </div>
      <div class="info-item">
        <label>Has Models/Variants</label>
        <span :class="['badge', product.has_model ? 'badge-success' : 'badge-secondary']">
          {{ product.has_model ? "Yes" : "No" }}
        </span>
      </div>
      <div class="info-item">
        <label>Created At</label>
        <span class="value">{{ formatDate(product.created_at) }}</span>
      </div>
      <div class="info-item">
        <label>Updated At</label>
        <span class="value">{{ formatDate(product.updated_at) }}</span>
      </div>
    </div>

    <div v-if="product.description" class="description-block">
      <label>Description</label>
      <div class="description-content">{{ product.description }}</div>
    </div>

    <div v-if="images.length > 0" class="images-section">
      <label>Product Images</label>
      <div class="images-grid">
        <div v-for="(image, index) in images" :key="index" class="image-item">
          <img :src="image" :alt="`Product image ${index + 1}`" />
        </div>
      </div>
    </div>

    <div v-if="product.attribute_list" class="attributes-section">
      <label>Attributes</label>
      <div class="attributes-list">
        <div v-for="(attr, index) in attributes" :key="index" class="attribute-item">
          <span class="attr-name">{{ attr.name }}</span>
          <span class="attr-value">{{ attr.value }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useProductFormatters } from "@/composables/useProductFormatters";

interface Props {
  product: any;
  images: string[];
}

const props = defineProps<Props>();
const { formatDate, formatPrice, parseAttributes } = useProductFormatters();

const attributes = computed(() => parseAttributes(props.product?.attribute_list));
</script>

<style scoped>
.info-section {
  margin-bottom: 30px;
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

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.info-item label {
  font-weight: 600;
  color: #7f8c8d;
  font-size: 0.9rem;
  text-transform: uppercase;
}

.value {
  color: #2c3e50;
  font-size: 1rem;
  word-break: break-word;
}

.value.sku {
  font-family: "Monaco", "Menlo", monospace;
  background: #f8f9fa;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9rem;
}

.value.price {
  color: #27ae60;
  font-weight: 600;
}

.value.stock {
  font-weight: 600;
  color: #2c3e50;
}

.value.stock.low {
  color: #e74c3c;
  background: #fadbd8;
  padding: 4px 8px;
  border-radius: 4px;
}

.badge {
  display: inline-block;
  padding: 6px 12px;
  border-radius: 12px;
  font-size: 0.85rem;
  font-weight: 500;
  width: fit-content;
}

.badge-success {
  background: #d4edda;
  color: #155724;
}

.badge-secondary {
  background: #e2e3e5;
  color: #383d41;
}

.description-block {
  margin-bottom: 20px;
}

.description-block label {
  display: block;
  font-weight: 600;
  color: #7f8c8d;
  font-size: 0.9rem;
  text-transform: uppercase;
  margin-bottom: 8px;
}

.description-content {
  background: #f8f9fa;
  padding: 15px;
  border-radius: 6px;
  color: #34495e;
  line-height: 1.6;
  max-height: 200px;
  overflow-y: auto;
}

.images-section {
  margin-bottom: 20px;
}

.images-section label {
  display: block;
  font-weight: 600;
  color: #7f8c8d;
  font-size: 0.9rem;
  text-transform: uppercase;
  margin-bottom: 12px;
}

.images-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 10px;
}

.image-item {
  aspect-ratio: 1;
  border-radius: 6px;
  overflow: hidden;
  border: 1px solid #ecf0f1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8f9fa;
}

.image-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.attributes-section {
  margin-bottom: 20px;
}

.attributes-section label {
  display: block;
  font-weight: 600;
  color: #7f8c8d;
  font-size: 0.9rem;
  text-transform: uppercase;
  margin-bottom: 12px;
}

.attributes-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.attribute-item {
  background: #f8f9fa;
  padding: 10px 12px;
  border-radius: 6px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}

.attr-name {
  font-weight: 600;
  color: #7f8c8d;
  font-size: 0.9rem;
}

.attr-value {
  color: #2c3e50;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }

  .images-grid {
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  }
}
</style>
