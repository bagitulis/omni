<template>
  <div class="pagination">
    <button
      @click="$emit('previous')"
      :disabled="currentOffset === 0"
      class="btn-page"
    >
      ◀ Prev
    </button>
    <span class="page-info">
      Halaman {{ currentPage }} ({{ currentOffset }}-{{ endRecord }} dari
      {{ totalRecords }})
    </span>
    <button
      @click="$emit('next')"
      :disabled="currentOffset + pageSize >= totalRecords"
      class="btn-page"
    >
      Next ▶
    </button>
    <div class="rows-per-page">
      <label for="rows-selector">Baris per halaman:</label>
      <select
        id="rows-selector"
        :value="pageSize"
        @change="handlePageSizeChange"
        class="rows-selector"
      >
        <option :value="50">50</option>
        <option :value="100">100</option>
        <option :value="150">150</option>
      </select>
    </div>
  </div>
</template>

<script>
import { defineComponent, computed } from "vue";

export default defineComponent({
  name: "InventoryPagination",
  props: {
    currentOffset: {
      type: Number,
      required: true,
    },
    pageSize: {
      type: Number,
      required: true,
    },
    totalRecords: {
      type: Number,
      required: true,
    },
  },
  emits: ["previous", "next", "page-size-changed"],

  setup(props, { emit }) {
    const currentPage = computed(() => {
      return Math.floor(props.currentOffset / props.pageSize) + 1;
    });

    const endRecord = computed(() => {
      return Math.min(props.currentOffset + props.pageSize, props.totalRecords);
    });

    const handlePageSizeChange = (event) => {
      const newSize = parseInt(event.target.value, 10);
      emit("page-size-changed", newSize);
    };

    return {
      currentPage,
      endRecord,
      handlePageSizeChange,
    };
  },
});
</script>

<style scoped>
.pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  margin-top: 1rem;
  padding: 0.5rem;
}

.btn-page {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
  border-radius: 4px;
}

.btn-page:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-page:hover:not(:disabled) {
  background: #f5f5f5;
}

.page-info {
  font-size: 0.9rem;
  color: #666;
}

.rows-per-page {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-left: 1rem;
}

.rows-per-page label {
  font-size: 0.85rem;
  color: #666;
}

.rows-selector {
  padding: 0.25rem 0.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
}
</style>
