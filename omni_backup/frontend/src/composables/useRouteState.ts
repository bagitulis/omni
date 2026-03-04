import { ref, Ref, computed } from 'vue';

/**
 * Route State Management
 * Manages per-route loading, error, and idle states
 * Single Responsibility: Handle route-specific state only
 */

export type RouteStateType = 'idle' | 'loading' | 'error';

interface RouteStateData {
  state: RouteStateType;
  error: string | null;
  timestamp: number;
}

interface RouteStateMap {
  [routeName: string]: RouteStateData;
}

export function useRouteState() {
  const stateMap: Ref<RouteStateMap> = ref({});

  /**
   * Set route to loading state
   */
  const setLoading = (routeName: string): void => {
    if (!routeName) return;
    stateMap.value[routeName] = {
      state: 'loading',
      error: null,
      timestamp: Date.now(),
    };
  };

  /**
   * Set route to success/idle state
   */
  const setSuccess = (routeName: string): void => {
    if (!routeName) return;
    stateMap.value[routeName] = {
      state: 'idle',
      error: null,
      timestamp: Date.now(),
    };
  };

  /**
   * Set route to error state with error message
   */
  const setError = (routeName: string, errorMessage: string): void => {
    if (!routeName || !errorMessage) return;
    stateMap.value[routeName] = {
      state: 'error',
      error: errorMessage,
      timestamp: Date.now(),
    };
  };

  /**
   * Get current state of a route
   */
  const getState = (routeName: string): RouteStateType => {
    return stateMap.value[routeName]?.state ?? 'idle';
  };

  /**
   * Get error message of a route
   */
  const getError = (routeName: string): string | null => {
    return stateMap.value[routeName]?.error ?? null;
  };

  /**
   * Check if route is currently loading
   */
  const isLoading = (routeName: string): boolean => {
    return getState(routeName) === 'loading';
  };

  /**
   * Check if route has error
   */
  const hasError = (routeName: string): boolean => {
    return getState(routeName) === 'error';
  };

  /**
   * Check if route is idle
   */
  const isIdle = (routeName: string): boolean => {
    return getState(routeName) === 'idle';
  };

  /**
   * Clear state for a route
   */
  const clearState = (routeName: string): void => {
    if (stateMap.value[routeName]) {
      delete stateMap.value[routeName];
    }
  };

  /**
   * Clear all states
   */
  const clearAllStates = (): void => {
    stateMap.value = {};
  };

  /**
   * Get all route states (computed for reactivity)
   */
  const getAllStates = computed(() => {
    return Object.entries(stateMap.value).reduce(
      (acc, [routeName, data]) => {
        acc[routeName] = {
          state: data.state,
          error: data.error,
          timestamp: data.timestamp,
        };
        return acc;
      },
      {} as RouteStateMap
    );
  });

  return {
    setLoading,
    setSuccess,
    setError,
    getState,
    getError,
    isLoading,
    hasError,
    isIdle,
    clearState,
    clearAllStates,
    getAllStates,
  };
}
