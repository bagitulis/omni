<template>
  <div class="skeleton-wrapper" :class="wrapperClass">
    <!-- Text Type -->
    <template v-if="type === 'text'">
      <div
        v-for="i in lines"
        :key="i"
        class="skeleton skeleton-text"
        :class="getLineClass(i)"
        :style="getLineStyle(i)"
      />
    </template>

    <!-- Avatar Type -->
    <template v-else-if="type === 'avatar'">
      <div
        class="skeleton skeleton-avatar"
        :class="size"
        :style="customStyle"
      />
    </template>

    <!-- Image Type -->
    <template v-else-if="type === 'image'">
      <div
        class="skeleton skeleton-image"
        :class="aspect"
        :style="customStyle"
      />
    </template>

    <!-- Button Type -->
    <template v-else-if="type === 'button'">
      <div
        class="skeleton skeleton-button"
        :class="{ full: fullWidth }"
        :style="customStyle"
      />
    </template>

    <!-- Card Type -->
    <template v-else-if="type === 'card'">
      <div class="skeleton skeleton-card" :style="customStyle">
        <div class="skeleton-header">
          <div class="skeleton skeleton-avatar" />
          <div class="skeleton-lines" style="flex: 1">
            <div class="skeleton skeleton-text medium" />
            <div class="skeleton skeleton-text short" />
          </div>
        </div>
        <div class="skeleton-lines">
          <div class="skeleton skeleton-text long" />
          <div class="skeleton skeleton-text" />
          <div class="skeleton skeleton-text medium" />
        </div>
      </div>
    </template>

    <!-- Table Type -->
    <template v-else-if="type === 'table'">
      <div class="skeleton-table">
        <div v-for="row in rows" :key="row" class="skeleton-table-row">
          <div
            v-for="col in columns"
            :key="col"
            class="skeleton skeleton-table-cell"
          />
        </div>
      </div>
    </template>

    <!-- Custom/Default Type -->
    <template v-else>
      <div class="skeleton" :style="customStyle" />
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, computed, PropType } from "vue";

type SkeletonType =
  | "text"
  | "avatar"
  | "image"
  | "button"
  | "card"
  | "table"
  | "custom";
type SkeletonSize = "sm" | "md" | "lg" | "xl";
type AspectRatio = "landscape" | "square" | "portrait";

export default defineComponent({
  name: "AppSkeleton",
  props: {
    type: {
      type: String as PropType<SkeletonType>,
      default: "text",
    },
    lines: { type: Number, default: 3 },
    rows: { type: Number, default: 5 },
    columns: { type: Number, default: 4 },
    size: { type: String as PropType<SkeletonSize>, default: "md" },
    aspect: { type: String as PropType<AspectRatio>, default: "landscape" },
    width: { type: String, default: "" },
    height: { type: String, default: "" },
    fullWidth: { type: Boolean, default: false },
    animated: { type: Boolean, default: true },
  },
  setup(props) {
    const wrapperClass = computed(() => ({
      "skeleton-animated": props.animated,
    }));

    const customStyle = computed(() => {
      const style: Record<string, string> = {};
      if (props.width) style.width = props.width;
      if (props.height) style.height = props.height;
      return style;
    });

    const getLineClass = (index: number) => {
      if (index === props.lines) return "short";
      if (index === 1) return "long";
      return index % 2 === 0 ? "medium" : "";
    };

    const getLineStyle = (index: number) => {
      if (props.width && index === 1) return { width: props.width };
      return {};
    };

    return { wrapperClass, customStyle, getLineClass, getLineStyle };
  },
});
</script>

<style src="./Skeleton.styles.css" scoped></style>
