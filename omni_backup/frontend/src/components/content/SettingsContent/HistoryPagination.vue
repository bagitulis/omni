<template>
  <div class="pagination">
    <div class="pagination-info">
      Showing {{ paginationStart }} - {{ paginationEnd }} of
      {{ totalItems }} items
    </div>
    <div class="pagination-controls">
      <button
        class="btn btn-page"
        :disabled="currentPage <= 1"
        @click="$emit('goToPage', 1)"
        title="First page"
      >
        ⏮️
      </button>
      <button
        class="btn btn-page"
        :disabled="currentPage <= 1"
        @click="$emit('goToPage', currentPage - 1)"
        title="Previous page"
      >
        ◀️ Prev
      </button>

      <div class="page-numbers">
        <button
          v-for="pageNum in visiblePages"
          :key="pageNum"
          :class="['btn', 'btn-page-num', { active: pageNum === currentPage }]"
          @click="$emit('goToPage', pageNum)"
        >
          {{ pageNum }}
        </button>
      </div>

      <button
        class="btn btn-page"
        :disabled="currentPage >= totalPages"
        @click="$emit('goToPage', currentPage + 1)"
        title="Next page"
      >
        Next ▶️
      </button>
      <button
        class="btn btn-page"
        :disabled="currentPage >= totalPages"
        @click="$emit('goToPage', totalPages)"
        title="Last page"
      >
        ⏭️
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  currentPage: number;
  totalPages: number;
  paginationStart: number;
  paginationEnd: number;
  totalItems: number;
  visiblePages: number[];
}>();

defineEmits<{
  goToPage: [page: number];
}>();
</script>

<style src="./HistoryTab.styles.css" scoped></style>
