import { ref, Ref } from 'vue';

/**
 * Navigation Queue Management
 * Prevents duplicate/concurrent navigation requests
 * Single Responsibility: Manage navigation request queue only
 */

interface QueuedNavigation {
  id: string;
  routeName: string;
  timestamp: number;
  priority: 'normal' | 'high';
}

export function useNavigationQueue() {
  const queue: Ref<Map<string, QueuedNavigation>> = ref(new Map());
  const processingRoute: Ref<string | null> = ref(null);

  /**
   * Generate unique ID for navigation
   */
  const generateId = (): string => {
    return `nav_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  };

  /**
   * Add navigation to queue
   * Returns ID for tracking
   */
  const addToQueue = (routeName: string, priority: 'normal' | 'high' = 'normal'): string => {
    if (!routeName) throw new Error('Route name is required');

    const id = generateId();
    
    // Remove existing entry for same route (keep only latest)
    const existingEntries = Array.from(queue.value.entries()).filter(
      ([_, nav]) => nav.routeName === routeName
    );
    
    existingEntries.forEach(([existingId]) => {
      queue.value.delete(existingId);
    });

    // Add new entry
    const navItem: QueuedNavigation = {
      id,
      routeName,
      timestamp: Date.now(),
      priority,
    };

    queue.value.set(id, navItem);
    return id;
  };

  /**
   * Get next navigation from queue
   */
  const getNextNavigation = (): QueuedNavigation | null => {
    if (processingRoute.value !== null) {
      return null; // Still processing, wait
    }

    if (queue.value.size === 0) {
      return null; // Queue empty
    }

    // Get first item (priority order: high first, then by timestamp)
    const items = Array.from(queue.value.values());
    items.sort((a, b) => {
      if (a.priority !== b.priority) {
        return a.priority === 'high' ? -1 : 1;
      }
      return a.timestamp - b.timestamp;
    });

    return items[0] || null;
  };

  /**
   * Start processing a navigation
   */
  const startProcessing = (navId: string): void => {
    const nav = queue.value.get(navId);
    if (!nav) {
      throw new Error(`Navigation ${navId} not found in queue`);
    }
    processingRoute.value = nav.routeName;
  };

  /**
   * Finish processing a navigation
   */
  const finishProcessing = (navId: string): void => {
    queue.value.delete(navId);
    processingRoute.value = null;
  };

  /**
   * Check if route is currently being processed
   */
  const isProcessing = (routeName: string): boolean => {
    return processingRoute.value === routeName;
  };

  /**
   * Check if route is in queue
   */
  const isInQueue = (routeName: string): boolean => {
    return Array.from(queue.value.values()).some(nav => nav.routeName === routeName);
  };

  /**
   * Get queue size
   */
  const getQueueSize = (): number => {
    return queue.value.size;
  };

  /**
   * Clear queue
   */
  const clearQueue = (): void => {
    queue.value.clear();
    processingRoute.value = null;
  };

  /**
   * Get all queued navigations
   */
  const getAllQueued = (): QueuedNavigation[] => {
    return Array.from(queue.value.values());
  };

  return {
    addToQueue,
    getNextNavigation,
    startProcessing,
    finishProcessing,
    isProcessing,
    isInQueue,
    getQueueSize,
    clearQueue,
    getAllQueued,
    processingRoute,
  };
}
