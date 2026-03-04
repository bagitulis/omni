<template>
  <div class="route-mapping-page">
    <RouteMappingViewer />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import RouteMappingViewer from "../components/RouteMapper/RouteMappingViewer.vue";
import { useAppStore } from "@/store/app";

const appStore = useAppStore();

onMounted(() => {
  if (appStore.connectionStatus !== "connected") {
    appStore.initializeApp();
  }
});
</script>

<style scoped>
.route-mapping-page {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  background: #f5f5f5;
}
</style>
