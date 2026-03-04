/**
 * Page Size Persistence - Menyimpan dan memuat pageSize dari localStorage
 * File: usePageSizePersistence.ts
 * Mengikuti prinsip: Clean Code, DRY, SRP
 */

const PAGE_SIZE_STORAGE_KEY = "inventory_page_size";
const DEFAULT_PAGE_SIZE = 50;

/**
 * Load pageSize dari localStorage
 * @returns {number} - pageSize yang tersimpan atau default 50
 */
export function loadPageSize(): number {
  try {
    const stored = localStorage.getItem(PAGE_SIZE_STORAGE_KEY);
    if (stored) {
      const parsed = parseInt(stored, 10);
      // Validasi hanya nilai yang diperbolehkan
      if ([50, 100, 150].includes(parsed)) {
        return parsed;
      }
    }
  } catch (error) {
    console.error("Error loading page size from localStorage:", error);
  }
  return DEFAULT_PAGE_SIZE;
}

/**
 * Save pageSize ke localStorage
 * @param {number} pageSize - Jumlah baris per halaman
 */
export function savePageSize(pageSize: number): void {
  try {
    localStorage.setItem(PAGE_SIZE_STORAGE_KEY, pageSize.toString());
  } catch (error) {
    console.error("Error saving page size to localStorage:", error);
  }
}

/**
 * Reset pageSize ke default
 */
export function resetPageSize(): void {
  try {
    localStorage.removeItem(PAGE_SIZE_STORAGE_KEY);
  } catch (error) {
    console.error("Error resetting page size:", error);
  }
}
