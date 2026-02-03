<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="closeIfNotProcessing">
      <div class="modal-content wholesale-modal">
        <!-- Header -->
        <div class="modal-header">
          <h3><Icon name="shopping-cart" size="sm" /> Update Harga Grosir</h3>
          <button class="close-btn" @click="closeIfNotProcessing">✕</button>
        </div>

        <!-- Body -->
        <div class="modal-body">
          <!-- Tab Navigation -->
          <div class="tab-nav">
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'wholesale' }"
              @click="activeTab = 'wholesale'"
            >
              <Icon name="shopping-cart" size="sm" /> Wholesale
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'mpq' }"
              @click="activeTab = 'mpq'"
            >
              <Icon name="document" size="sm" /> MPQ
            </button>
            <button
              class="tab-btn delete"
              :class="{ active: activeTab === 'delete' }"
              @click="activeTab = 'delete'"
            >
              <Icon name="trash" size="sm" /> Delete
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeTab === 'settings' }"
              @click="activeTab = 'settings'"
            >
              <Icon name="settings" size="sm" /> Settings
            </button>
          </div>

          <!-- WHOLESALE TAB (Shopee Only) -->
          <WholesaleTab
            v-if="activeTab === 'wholesale'"
            :items="shopeeItems"
            :settings="settings"
            :processing="processing"
            @update="handleWholesaleUpdate"
          />

          <!-- MPQ TAB (Shopee + TikTok) -->
          <MpqTab
            v-if="activeTab === 'mpq'"
            :shopee-items="shopeeItems"
            :tiktok-items="tiktokItems"
            :settings="settings"
            :processing="processing"
            @update="handleMpqUpdate"
          />

          <!-- DELETE TAB (Shopee + TikTok) -->
          <DeleteTab
            v-if="activeTab === 'delete'"
            :shopee-items="shopeeItems"
            :tiktok-items="tiktokItems"
            :processing="processing"
            @update="handleDeleteUpdate"
          />

          <!-- SETTINGS TAB -->
          <SettingsTab
            v-if="activeTab === 'settings'"
            :settings="settings"
            @save="handleSettingsSave"
            @update:settings="settings = $event"
          />

          <!-- Platform Summary -->
          <div class="platform-summary">
            <span class="summary-item shopee">
              <Icon name="store" size="sm" /> Shopee: {{ shopeeItems.length }}
              <Icon
                v-if="activeTab === 'wholesale' || activeTab === 'mpq'"
                name="check"
                size="sm"
              />
            </span>
            <span class="summary-item tiktok">
              <Icon name="chart" size="sm" /> TikTok: {{ tiktokItems.length }}
              <Icon v-if="activeTab === 'mpq'" name="check" size="sm" />
              <Icon v-else name="warning" size="sm" />
            </span>
            <span class="summary-item lazada">
              <Icon name="store" size="sm" /> Lazada: {{ lazadaItems.length }}
              <Icon name="warning" size="sm" /> Tidak didukung
            </span>
          </div>

          <!-- Result Section -->
          <ResultSection
            v-if="result"
            :result="result"
            :visible="!!result"
            @close="result = null"
          />

          <!-- Error Message -->
          <div v-if="error" class="error-message">
            <Icon name="warning" size="sm" /> {{ error }}
          </div>
        </div>

        <!-- Footer -->
        <div class="modal-footer">
          <button class="btn-cancel" @click="closeIfNotProcessing">
            {{ result ? "Tutup" : "Batal" }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue";
import Icon from "@/components/ui/Icon.vue";
import WholesaleTab from "./tabs/WholesaleTab.vue";
import MpqTab from "./tabs/MpqTab.vue";
import DeleteTab from "./tabs/DeleteTab.vue";
import SettingsTab from "./tabs/SettingsTab.vue";
import ResultSection from "./ResultSection.vue";
import wholesaleService, {
  WholesaleSettings,
} from "../../../services/wholesaleService";

export interface UpdateItem {
  sku: string;
  price: number;
  platform: "shopee" | "tiktok" | "lazada";
}

const props = defineProps<{
  show: boolean;
  items: UpdateItem[];
}>();

const emit = defineEmits<{
  close: [];
  completed: [result: any];
}>();

// State
const activeTab = ref<"wholesale" | "mpq" | "delete" | "settings">("wholesale");
const processing = ref(false);
const result = ref<any>(null);
const error = ref("");

// Settings
const settings = ref<WholesaleSettings>({
  tenant_id: "",
  platform: "shopee",
  admin_fee: 1500,
  min_order_1: 2,
  max_order_1: 3,
  max_order_tier_3: 1000,
});

// Filter items by platform
const shopeeItems = computed(() =>
  props.items.filter((i) => i.platform === "shopee"),
);
const tiktokItems = computed(() =>
  props.items.filter((i) => i.platform === "tiktok"),
);
const lazadaItems = computed(() =>
  props.items.filter((i) => i.platform === "lazada"),
);

// Load settings on mount
onMounted(loadSettings);

// Reset when modal opens
watch(
  () => props.show,
  async (newVal) => {
    if (newVal) {
      result.value = null;
      error.value = "";
      activeTab.value = "wholesale";
      await loadSettings();
    }
  },
);

async function loadSettings() {
  try {
    const loaded = await wholesaleService.getSettings();
    if (loaded) settings.value = loaded;
  } catch (err) {
    console.error("Failed to load settings:", err);
  }
}

async function handleSettingsSave(newSettings: WholesaleSettings) {
  settings.value = newSettings;
}

async function handleWholesaleUpdate(updateResult: any) {
  result.value = updateResult;
  emit("completed", updateResult);
}

async function handleMpqUpdate(updateResult: any) {
  result.value = updateResult;
  emit("completed", updateResult);
}

async function handleDeleteUpdate(updateResult: any) {
  result.value = updateResult;
  emit("completed", updateResult);
}

function closeIfNotProcessing() {
  if (!processing.value) {
    emit("close");
  }
}
</script>

<style src="./WholesaleMpqModal.styles.css" scoped></style>
