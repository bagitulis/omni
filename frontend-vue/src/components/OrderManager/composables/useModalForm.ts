/**
 * useModalForm Composable
 * Shared form handling logic for order modals (cancel, ship)
 */

import { ref, watch } from "vue";

export interface ModalFormState {
  [key: string]: any;
}

export function useModalForm(initialState: ModalFormState) {
  const form = ref<ModalFormState>({ ...initialState });
  const loading = ref(false);
  const error = ref("");

  const resetForm = () => {
    form.value = { ...initialState };
    error.value = "";
  };

  const setLoading = (value: boolean) => {
    loading.value = value;
  };

  const setError = (message: string) => {
    error.value = message;
  };

  const clearError = () => {
    error.value = "";
  };

  return {
    form,
    loading,
    error,
    resetForm,
    setLoading,
    setError,
    clearError,
  };
}

/**
 * Platform formatting utility
 */
export const platformFormatter = {
  format: (platform?: string): string => {
    const map: Record<string, string> = {
      shopee: "Shopee",
      lazada: "Lazada",
      tiktok: "TikTok",
    };
    return map[platform?.toLowerCase() || ""] || platform || "";
  },
};

/**
 * Cancellation reasons by platform
 */
export interface CancellationReason {
  value: string;
  label: string;
}

export const getCancellationReasons = (
  platform?: string,
): CancellationReason[] => {
  const p = platform?.toLowerCase();
  const commonReasons: CancellationReason[] = [
    { value: "OUT_OF_STOCK", label: "Out of stock" },
    { value: "BUYER_REQUEST", label: "Buyer requested cancellation" },
    { value: "WRONG_PRICE", label: "Wrong price/listing error" },
    { value: "DUPLICATE_ORDER", label: "Duplicate order" },
    { value: "OTHER", label: "Other reason" },
  ];

  if (p === "shopee") {
    return [
      { value: "CUSTOMER_REQUEST", label: "Customer requested cancellation" },
      { value: "OUT_OF_STOCK", label: "Out of stock" },
      { value: "UNDELIVERABLE_AREA", label: "Undeliverable area" },
      { value: "COD_NOT_SUPPORTED", label: "COD not supported" },
      ...commonReasons.filter(
        (r) => !["OUT_OF_STOCK", "BUYER_REQUEST"].includes(r.value),
      ),
    ];
  }

  if (p === "lazada") {
    return [
      { value: "customer_request", label: "Customer requested cancellation" },
      { value: "out_of_stock", label: "Out of stock" },
      { value: "sourcing_failed", label: "Sourcing failed" },
      ...commonReasons.filter(
        (r) => !["OUT_OF_STOCK", "BUYER_REQUEST"].includes(r.value),
      ),
    ];
  }

  return commonReasons;
};

/**
 * Shipping providers list
 */
export interface ShippingProvider {
  value: string;
  label: string;
}

export const getShippingProviders = (): ShippingProvider[] => [
  { value: "jne", label: "JNE" },
  { value: "jnt", label: "J&T Express" },
  { value: "sicepat", label: "SiCepat" },
  { value: "anteraja", label: "AnterAja" },
  { value: "ninja", label: "Ninja Van" },
  { value: "shopee_express", label: "Shopee Express" },
  { value: "lazada_logistics", label: "Lazada Logistics" },
];
