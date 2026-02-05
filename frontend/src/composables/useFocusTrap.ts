import { ref, onMounted, onUnmounted, watch, type Ref } from "vue";

/**
 * Focus trap composable for modal accessibility
 * Traps focus within a container element when active
 */
export function useFocusTrap(
  containerRef: Ref<HTMLElement | null>,
  isActive: Ref<boolean>,
) {
  const previousActiveElement = ref<HTMLElement | null>(null);

  const getFocusableElements = (container: HTMLElement): HTMLElement[] => {
    const selectors = [
      "button:not([disabled])",
      "input:not([disabled])",
      "select:not([disabled])",
      "textarea:not([disabled])",
      "a[href]",
      '[tabindex]:not([tabindex="-1"])',
    ].join(", ");

    return Array.from(container.querySelectorAll<HTMLElement>(selectors));
  };

  const handleKeydown = (e: KeyboardEvent) => {
    if (e.key !== "Tab" || !containerRef.value) return;

    const focusableElements = getFocusableElements(containerRef.value);
    if (focusableElements.length === 0) return;

    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    if (e.shiftKey) {
      // Shift + Tab: going backwards
      if (document.activeElement === firstElement) {
        e.preventDefault();
        lastElement.focus();
      }
    } else {
      // Tab: going forward
      if (document.activeElement === lastElement) {
        e.preventDefault();
        firstElement.focus();
      }
    }
  };

  const activate = () => {
    if (!containerRef.value) return;

    // Store current active element to restore later
    previousActiveElement.value = document.activeElement as HTMLElement;

    // Focus first focusable element in container
    const focusableElements = getFocusableElements(containerRef.value);
    if (focusableElements.length > 0) {
      // Try to focus close button first, otherwise first focusable
      const closeButton = containerRef.value.querySelector<HTMLElement>(
        '[aria-label="Close"], [aria-label="Close modal"]',
      );
      if (closeButton) {
        closeButton.focus();
      } else {
        focusableElements[0].focus();
      }
    }

    document.addEventListener("keydown", handleKeydown);
  };

  const deactivate = () => {
    document.removeEventListener("keydown", handleKeydown);

    // Restore focus to previously focused element
    if (previousActiveElement.value) {
      previousActiveElement.value.focus();
      previousActiveElement.value = null;
    }
  };

  watch(isActive, (active) => {
    if (active) {
      // Use nextTick equivalent to ensure DOM is ready
      setTimeout(activate, 0);
    } else {
      deactivate();
    }
  });

  onMounted(() => {
    if (isActive.value) {
      setTimeout(activate, 0);
    }
  });

  onUnmounted(() => {
    deactivate();
  });

  return {
    activate,
    deactivate,
  };
}
