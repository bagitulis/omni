// composables/useVirtualScroll.ts
// Virtual scrolling composable for infinite scroll with cursor pagination

import { ref, computed, Ref } from "vue";
import { useInfiniteQuery } from "@tanstack/vue-query";

export interface CursorPaginationResult<T> {
  data: T[];
  next_cursor?: string;
  has_more: boolean;
  total: number;
}

export interface UseVirtualScrollOptions<T> {
  queryKey: string[];
  queryFn: (cursor?: string) => Promise<CursorPaginationResult<T>>;
  limit?: number;
  enabled?: Ref<boolean>;
}

export function useVirtualScroll<T>({
  queryKey,
  queryFn,
  limit = 100,
  enabled = ref(true),
}: UseVirtualScrollOptions<T>) {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
    isError,
    error,
  } = useInfiniteQuery({
    queryKey,
    queryFn: ({ pageParam }) => queryFn(pageParam),
    getNextPageParam: (lastPage) => {
      return lastPage.has_more ? lastPage.next_cursor : undefined;
    },
    initialPageParam: undefined as string | undefined,
    enabled,
  });

  // Flatten all pages into single array
  const allItems = computed(() => {
    if (!data.value) return [];
    return data.value.pages.flatMap((page) => page.data);
  });

  const total = computed(() => {
    if (!data.value || data.value.pages.length === 0) return 0;
    return data.value.pages[0].total;
  });

  // Load more when scrolled near bottom
  const onLoadMore = () => {
    if (hasNextPage.value && !isFetchingNextPage.value) {
      fetchNextPage();
    }
  };

  return {
    items: allItems,
    total,
    isLoading,
    isError,
    error,
    isFetchingMore: isFetchingNextPage,
    hasMore: hasNextPage,
    loadMore: onLoadMore,
  };
}
