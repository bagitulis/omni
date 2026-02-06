import { ref } from 'vue';

/**
 * SINGLETON: Column-based edit lock manager
 * Shared state across all components with localStorage persistence
 * 
 * DEFAULT: All columns except those in default list are editable
 * User can customize via UI
 */
let instance: ReturnType<typeof createEditableLock> | null = null;

const STORAGE_KEY = 'inventory_locked_columns';

// Default locked columns - CAN be customized by user
function getDefaultLockedColumns(): Set<string> {
  return new Set([
    'SKU',
    'CEK',
    'TOTAL',
    'Final Masuk',
    'EXPIRED',
    'EX',
    'KARTON',
    'LUSIN',
    'PCS',
    'Masuk',
  ]);
}

function loadLockedColumnsFromStorage(): Set<string> {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      if (Array.isArray(parsed)) {
        console.log('📂 Loaded locked columns from storage:', parsed);
        return new Set(parsed);
      }
    }
  } catch (error) {
    console.error('❌ Error loading locked columns from storage:', error);
  }
  return getDefaultLockedColumns();
}

function saveLockedColumnsToStorage(locked: Set<string>) {
  try {
    const array = Array.from(locked);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(array));
    console.log('💾 Saved locked columns to storage:', array);
  } catch (error) {
    console.error('❌ Error saving locked columns to storage:', error);
  }
}

function createEditableLock() {
  // Load from localStorage, fallback to defaults
  const lockedColumns = ref<Set<string>>(loadLockedColumnsFromStorage());

  const isColumnEditable = (columnName: string): boolean => {
    return !lockedColumns.value.has(columnName);
  };

  const isColumnLocked = (columnName: string): boolean => {
    return lockedColumns.value.has(columnName);
  };

  const setLocked = (columnNames: string[]) => {
    lockedColumns.value = new Set(columnNames);
    saveLockedColumnsToStorage(lockedColumns.value);
    console.log('🔒 Lock configuration set:', columnNames);
  };

  const addLockedColumn = (columnName: string) => {
    lockedColumns.value.add(columnName);
    saveLockedColumnsToStorage(lockedColumns.value);
    console.log(`✅ Column "${columnName}" locked`);
  };

  const removeLockedColumn = (columnName: string) => {
    lockedColumns.value.delete(columnName);
    saveLockedColumnsToStorage(lockedColumns.value);
    console.log(`🔓 Column "${columnName}" unlocked`);
  };

  const toggleColumnLock = (columnName: string) => {
    if (lockedColumns.value.has(columnName)) {
      lockedColumns.value.delete(columnName);
      console.log(`🔓 Toggled unlock: "${columnName}"`);
    } else {
      lockedColumns.value.add(columnName);
      console.log(`🔒 Toggled lock: "${columnName}"`);
    }
    saveLockedColumnsToStorage(lockedColumns.value);
  };

  const getLockedColumnsList = (): string[] => {
    return Array.from(lockedColumns.value);
  };

  const getEditableColumnsList = (allColumns: string[]): string[] => {
    return allColumns.filter(col => !lockedColumns.value.has(col));
  };

  return {
    lockedColumns,
    isColumnEditable,
    isColumnLocked,
    setLocked,
    addLockedColumn,
    removeLockedColumn,
    toggleColumnLock,
    getLockedColumnsList,
    getEditableColumnsList,
  };
}

export function useEditableLock() {
  if (!instance) {
    instance = createEditableLock();
  }
  return instance;
}

/**
 * Reset singleton instance (for testing or component cleanup)
 */
export function resetEditableLock() {
  instance = null;
}
