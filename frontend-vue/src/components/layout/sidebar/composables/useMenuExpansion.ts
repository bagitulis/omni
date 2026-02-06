import { useUIStore, type ExpandedSections } from "@/store/ui";
import { storeToRefs } from "pinia";

export type { ExpandedSections };

export function useMenuExpansion() {
  const uiStore = useUIStore();
  const { expandedSections } = storeToRefs(uiStore);

  const toggleExpand = (section: string) => {
    uiStore.toggleMenuSection(section);
  };

  return {
    expandedSections,
    toggleExpand,
  };
}
