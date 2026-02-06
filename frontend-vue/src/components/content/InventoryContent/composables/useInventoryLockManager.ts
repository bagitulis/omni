import { computed, watch } from 'vue';
import { useInventoryConfig } from './useInventoryConfig';

/**
 * Composable for the InventoryLockPanel component
 * Manages the UI state for the lock filter panel
 * Wraps useInventoryConfig to provide lock-specific interface
 */
export function useInventoryLockManager(
  schemaColumns: any[],
  columnVisibility: Record<string, boolean>
) {
  const config = useInventoryConfig();
  
  // Log initialization
  console.log('%c🔒 [useInventoryLockManager] INITIALIZED', 'color: #e74c3c; font-weight: bold', {
    schema_columns: schemaColumns?.map((c: any) => c.column_name || c) || [],
    visible_columns: Object.entries(columnVisibility).filter(([_, v]) => v).map(([k]) => k),
    initial_locked: Array.from(config.lockedColumns.value),
    timestamp: new Date().toLocaleTimeString(),
  });
  
  // Watch lock changes
  watch(config.lockedColumns, (newLocked) => {
    console.log('%c🔄 [useInventoryLockManager] LOCK STATE CHANGED', 'color: #f39c12; font-weight: bold', {
      locked_columns: Array.from(newLocked),
      locked_count: newLocked.size,
      timestamp: new Date().toLocaleTimeString(),
    });
  });

  // Get editable columns based on schema and visibility
  const editableColumns = computed(() => {
    if (!schemaColumns || schemaColumns.length === 0) {
      return [];
    }
    
    return schemaColumns
      .map((col: any) => col.column_name || col.name || col)
      .filter((col: string) => columnVisibility[col] !== false); // Include visible columns
  });

  // Get locked count
  const lockedCount = computed(() => {
    return config.lockedColumns.value.size;
  });

  // Get locked list
  const getLockedList = computed(() => {
    return Array.from(config.lockedColumns.value);
  });

  // Check if column is locked
  const isLocked = (column: string): boolean => {
    const locked = config.isColumnLocked(column);
    console.log(`%c🔍 [isLocked] "${column}" is ${locked ? '🔒 LOCKED' : '✏️ EDITABLE'}`, 
      locked ? 'color: #e74c3c' : 'color: #27ae60');
    return locked;
  };

  // Toggle lock for a single column
  const toggleLock = (column: string) => {
    const wasLocked = config.isColumnLocked(column);
    config.toggleColumnLock(column);
    const isNowLocked = config.isColumnLocked(column);
    console.log(
      `%c🔄 [toggleLock] "${column}" toggled: ${wasLocked ? 'LOCKED' : 'EDITABLE'} → ${isNowLocked ? 'LOCKED' : 'EDITABLE'}`,
      isNowLocked ? 'color: #e74c3c; font-weight: bold' : 'color: #27ae60; font-weight: bold'
    );
  };

  // Lock all editable columns
  const lockAll = () => {
    const allColumns = editableColumns.value;
    const currentLocked = config.getLockedColumns();
    const newLocked = new Set([...currentLocked, ...allColumns]);
    config.setLockedColumns(Array.from(newLocked));
    console.log(
      '%c🔐 [lockAll] LOCKING ALL COLUMNS',
      'color: #e74c3c; font-weight: bold; font-size: 12px',
      {
        columns_to_lock: allColumns,
        total_locked: newLocked.size,
        all_locked_list: Array.from(newLocked),
        timestamp: new Date().toLocaleTimeString(),
      }
    );
  };

  // Unlock all columns
  const unlockAll = () => {
    const beforeCount = config.lockedColumns.value.size;
    config.setLockedColumns([]);
    console.log(
      '%c🔓 [unlockAll] UNLOCKING ALL COLUMNS',
      'color: #27ae60; font-weight: bold; font-size: 12px',
      {
        unlocked_count: beforeCount,
        remaining_locked: 0,
        timestamp: new Date().toLocaleTimeString(),
      }
    );
  };

  return {
    editableColumns,
    lockedCount,
    getLockedList,
    isLocked,
    toggleLock,
    lockAll,
    unlockAll,
    lockedColumns: config.lockedColumns,
  };
}
