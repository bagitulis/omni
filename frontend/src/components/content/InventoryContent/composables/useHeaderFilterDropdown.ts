import { ref } from "vue";

/**
 * Global state untuk manage dropdown yang terbuka
 * Hanya 1 dropdown yang boleh terbuka pada saat bersamaan
 */
const openDropdown = ref<string | null>(null);

export function useHeaderFilterDropdown() {
  const isDropdownOpen = (columnName: string): boolean => {
    return openDropdown.value === columnName;
  };

  const openDropdownFor = (columnName: string) => {
    openDropdown.value = columnName;
  };

  const closeDropdown = () => {
    openDropdown.value = null;
  };

  const toggleDropdown = (columnName: string) => {
    if (openDropdown.value === columnName) {
      closeDropdown();
    } else {
      openDropdownFor(columnName);
    }
  };

  return {
    isDropdownOpen,
    openDropdownFor,
    closeDropdown,
    toggleDropdown,
  };
}
