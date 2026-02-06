<template>
  <div class="tab-content">
    <!-- Platform Support Info -->
    <div class="info-banner">
      <Icon name="shopping-cart" size="sm" />
      <div class="info-text">
        <strong>MPQ (Min Purchase Quantity)</strong>
        <p class="info-note">
          Shopee: min_purchase_limit + delete wholesale | TikTok:
          minimum_order_quantity
        </p>
      </div>
    </div>

    <!-- Tier Selector -->
    <div class="tier-selector">
      <label>Pilih Tier:</label>
      <div class="tier-options">
        <button
          v-for="tier in tiers"
          :key="tier.value"
          :class="['tier-btn', { active: selectedTier === tier.value }]"
          @click="selectedTier = tier.value"
        >
          {{ tier.label }}
          <span class="tier-qty">Min: {{ tier.minQty }}</span>
        </button>
      </div>
    </div>

    <!-- Preview Section -->
    <div
      class="preview-section"
      v-if="shopeeItems.length + tiktokItems.length > 0"
    >
      <!-- Shopee Preview -->
      <div v-if="shopeeItems.length > 0" class="platform-preview">
        <h3 class="platform-title">
          <span class="platform-badge shopee">Shopee</span>
          {{ shopeeItems.length }} item
        </h3>
        <table class="preview-table" aria-label="Shopee MPQ preview">
          <thead>
            <tr>
              <th scope="col">SKU</th>
              <th scope="col">Harga Awal</th>
              <th scope="col">Harga Baru</th>
              <th scope="col">MPQ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in shopeePreview" :key="item.sku">
              <td class="sku-cell">{{ item.sku }}</td>
              <td>{{ formatPrice(item.price) }}</td>
              <td class="price-cell">{{ formatPrice(item.newPrice) }}</td>
              <td>{{ selectedMinQty }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="shopeeItems.length > 3" class="more-note">
          ... +{{ shopeeItems.length - 3 }} lainnya
        </p>
      </div>

      <!-- TikTok Preview -->
      <div v-if="tiktokItems.length > 0" class="platform-preview">
        <h3 class="platform-title">
          <span class="platform-badge tiktok">TikTok</span>
          {{ tiktokItems.length }} item
        </h3>
        <table class="preview-table" aria-label="TikTok MPQ preview">
          <thead>
            <tr>
              <th scope="col">SKU</th>
              <th scope="col">Harga Awal</th>
              <th scope="col">Harga Baru</th>
              <th scope="col">MPQ</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in tiktokPreview" :key="item.sku">
              <td class="sku-cell">{{ item.sku }}</td>
              <td>{{ formatPrice(item.price) }}</td>
              <td class="price-cell">{{ formatPrice(item.newPrice) }}</td>
              <td>{{ selectedMinQty }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="tiktokItems.length > 3" class="more-note">
          ... +{{ tiktokItems.length - 3 }} lainnya
        </p>
      </div>
    </div>

    <div v-else class="no-items">Tidak ada item Shopee/TikTok yang dipilih</div>

    <!-- Action Buttons -->
    <div
      class="action-buttons"
      v-if="shopeeItems.length + tiktokItems.length > 0"
    >
      <button
        class="btn-update"
        @click="handleUpdate"
        :disabled="processing || localProcessing"
      >
        <span v-if="localProcessing" class="btn-loading">
          <span class="spinner"></span>
          Updating...
        </span>
        <span v-else><Icon name="shopping-cart" size="sm" /> Update MPQ</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import Icon from "@/components/ui/Icon.vue";
import type { WholesaleSettings } from "../../../../services/wholesaleService";
import wholesaleService from "../../../../services/wholesaleService";

interface UpdateItem {
  sku: string;
  price: number;
  platform: string;
}

const props = defineProps<{
  shopeeItems: UpdateItem[];
  tiktokItems: UpdateItem[];
  settings: WholesaleSettings;
  processing: boolean;
}>();

const emit = defineEmits<{
  update: [result: any];
}>();

const selectedTier = ref<"normal" | "tier1" | "tier2" | "tier3">("tier1");
const localProcessing = ref(false);

const tiers = computed(() => [
  { value: "normal" as const, label: "Normal", minQty: 1 },
  {
    value: "tier1" as const,
    label: "Tier 1",
    minQty: props.settings.min_order_1,
  },
  {
    value: "tier2" as const,
    label: "Tier 2",
    minQty: props.settings.max_order_1 + 1,
  },
  {
    value: "tier3" as const,
    label: "Tier 3",
    minQty: props.settings.max_order_1 + 3,
  },
]);

const selectedMinQty = computed(() => {
  const tier = tiers.value.find((t) => t.value === selectedTier.value);
  return tier?.minQty || 1;
});

function calculateMpqPrice(originalPrice: number): number {
  if (selectedTier.value === "normal") return originalPrice;
  const tierCalc = wholesaleService.calculateTiersLocal(
    originalPrice,
    props.settings,
  );
  const tierIndex = { tier1: 0, tier2: 1, tier3: 2 }[selectedTier.value] ?? 0;
  return tierCalc[tierIndex]?.unit_price || originalPrice;
}

const shopeePreview = computed(() =>
  props.shopeeItems.slice(0, 3).map((item) => ({
    ...item,
    newPrice: calculateMpqPrice(item.price),
  })),
);

const tiktokPreview = computed(() =>
  props.tiktokItems.slice(0, 3).map((item) => ({
    ...item,
    newPrice: calculateMpqPrice(item.price),
  })),
);

function formatPrice(price: number): string {
  return new Intl.NumberFormat("id-ID").format(price);
}

async function handleUpdate() {
  if (localProcessing.value) return;
  localProcessing.value = true;
  const results: any[] = [];

  try {
    if (props.shopeeItems.length > 0) {
      // Map items with calculated newPrice based on selected tier
      const shopeeItemsWithPrice = props.shopeeItems.map((item) => ({
        sku: item.sku,
        price: calculateMpqPrice(item.price), // Use calculated tier price, not original
      }));

      const shopeeResult = await wholesaleService.batchShopeeMpq(
        shopeeItemsWithPrice,
        selectedMinQty.value,
      );
      // Extract failed SKUs from results array
      const failedSkus =
        shopeeResult.data?.results
          ?.filter((r: any) => !r.success)
          ?.map((r: any) => r.error || `Item ${r.item_id}`) || [];
      results.push({
        platform: "shopee",
        processed: shopeeResult.data?.processed || 0,
        failed: shopeeResult.data?.failed || 0,
        failedSkus,
      });
    }

    if (props.tiktokItems.length > 0) {
      // Map items with calculated newPrice based on selected tier
      const tiktokItemsWithPrice = props.tiktokItems.map((item) => ({
        sku: item.sku,
        price: calculateMpqPrice(item.price), // Use calculated tier price, not original
      }));

      const tiktokResult = await wholesaleService.batchTiktokMpq(
        tiktokItemsWithPrice,
        selectedMinQty.value,
      );
      // Extract failed from results array
      const failedSkus =
        tiktokResult.data?.results
          ?.filter((r: any) => !r.success)
          ?.map((r: any) => r.error || `Product ${r.product_id}`) || [];
      results.push({
        platform: "tiktok",
        processed: tiktokResult.data?.processed || 0,
        failed: tiktokResult.data?.failed || 0,
        failedSkus,
      });
    }

    const anyFailed = results.some((r) => r.failed > 0);
    emit("update", { success: !anyFailed, results });
  } catch (err: any) {
    emit("update", {
      success: false,
      message: err.message || "Gagal update MPQ",
    });
  } finally {
    localProcessing.value = false;
  }
}
</script>

<style src="./MpqTab.styles.css" scoped></style>
