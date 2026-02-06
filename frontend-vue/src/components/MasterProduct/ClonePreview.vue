<template>
  <Teleport to="body">
    <div v-if="modelValue" class="clone-preview-overlay" @click.self="close">
      <div class="clone-preview-modal">
        <!-- Header -->
        <div class="modal-header">
          <h2>Preview Clone</h2>
          <button @click="close" class="btn-close">&times;</button>
        </div>

        <!-- Status (Loading/Error) -->
        <ClonePreviewStatus
          :loading="loading"
          :error="error"
          @retry="loadPreview"
        />

        <!-- Preview Content -->
        <div v-if="!loading && !error && previewData" class="preview-content">
          <ClonePreviewInfo
            :has-conflict="previewData.has_conflict"
            :target-platform-label="targetPlatformLabel"
            :adjustments="previewData.adjustments"
          />

          <!-- Two Column Comparison -->
          <div class="comparison-grid">
            <ClonePreviewPanel
              :platform="sourcePlatform"
              :platform-label="sourcePlatformLabel"
              panel-label="Sumber"
              :product="previewData.source_product"
            />

            <!-- Arrow -->
            <div class="arrow-container">
              <span class="arrow">&rarr;</span>
            </div>

            <ClonePreviewPanel
              :platform="targetPlatform"
              :platform-label="targetPlatformLabel"
              :panel-label="previewData.has_conflict ? 'Sudah Ada' : 'Target'"
              :product="targetProductToDisplay"
              :is-conflict="previewData.has_conflict"
              :adjustments="
                previewData.has_conflict ? undefined : previewData.adjustments
              "
              :differences="previewData.differences"
            />
          </div>

          <ClonePreviewDifferences :differences="previewData.differences" />
        </div>

        <!-- Footer Actions -->
        <ClonePreviewActions
          v-if="!loading && !error"
          :has-conflict="!!previewData?.has_conflict"
          :cloning="cloning"
          :target-platform-label="targetPlatformLabel"
          @cancel="close"
          @clone="handleCloneAction"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import apiService from "@/services/api";
import type { PreviewData } from "./ClonePreview.types";
import ClonePreviewInfo from "./ClonePreviewInfo.vue";
import ClonePreviewPanel from "./ClonePreviewPanel.vue";
import ClonePreviewDifferences from "./ClonePreviewDifferences.vue";
import ClonePreviewActions from "./ClonePreviewActions.vue";
import ClonePreviewStatus from "./ClonePreviewStatus.vue";

const props = defineProps<{
  modelValue: boolean;
  sourcePlatform: string;
  sourceItemId: string;
  targetPlatform: string;
  sku?: string;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "clone", data: { updateExisting: boolean }): void;
}>();

const loading = ref(false);
const error = ref("");
const previewData = ref<PreviewData | null>(null);
const cloning = ref(false);

const getPlatformLabel = (platform: string) => {
  const labels: Record<string, string> = {
    shopee: "Shopee",
    lazada: "Lazada",
    tiktok: "TikTok",
  };
  return labels[platform] || platform;
};

const sourcePlatformLabel = computed(() =>
  getPlatformLabel(props.sourcePlatform),
);
const targetPlatformLabel = computed(() =>
  getPlatformLabel(props.targetPlatform),
);

const targetProductToDisplay = computed(() => {
  if (previewData.value?.has_conflict && previewData.value?.target_product) {
    return previewData.value.target_product;
  }
  return previewData.value?.source_product;
});

const loadPreview = async () => {
  loading.value = true;
  error.value = "";

  try {
    const params = new URLSearchParams({
      source_platform: props.sourcePlatform,
      target_platform: props.targetPlatform,
      source_item_id: props.sourceItemId,
    });
    if (props.sku) {
      params.append("sku", props.sku);
    }

    const response = await apiService.get<{
      success: boolean;
      data: PreviewData;
    }>(`/clone/preview?${params.toString()}`);

    if (response.success) {
      previewData.value = response.data;
    } else {
      throw new Error("Failed to load preview");
    }
  } catch (err: any) {
    console.error("Error loading preview:", err);
    error.value = err.message || "Gagal memuat preview";
  } finally {
    loading.value = false;
  }
};

const close = () => {
  emit("update:modelValue", false);
};

const handleCloneAction = (updateExisting: boolean) => {
  cloning.value = true;
  emit("clone", { updateExisting });
};

watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal) {
      loadPreview();
    } else {
      previewData.value = null;
      error.value = "";
      cloning.value = false;
    }
  },
);
</script>

<style scoped>
.clone-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.clone-preview-modal {
  background: white;
  border-radius: 1rem;
  width: 100%;
  max-width: 900px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
  color: white;
  border-radius: 1rem 1rem 0 0;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 700;
}

.btn-close {
  width: 36px;
  height: 36px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border-radius: 50%;
  font-size: 1.5rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: rotate(90deg);
}

.preview-content {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.comparison-grid {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 1rem;
  align-items: start;
}

.arrow-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 100px;
}

.arrow {
  font-size: 2rem;
  color: #9ca3af;
}

@media (max-width: 768px) {
  .comparison-grid {
    grid-template-columns: 1fr;
  }

  .arrow-container {
    padding: 0.5rem 0;
  }

  .arrow {
    transform: rotate(90deg);
  }
}
</style>
