import { ProductData, CloneTargetsResult, ConflictResult } from "@/types/clone";

export type PriceAdjustmentType =
  | "none"
  | "fixed"
  | "percentage_inc"
  | "percentage_dec"
  | "amount_inc"
  | "amount_dec";

export interface CloneProductModalProps {
  open: boolean;
  onClose: () => void;
  initialSku?: string;
  initialPlatform?: string;
}

export interface CloneSourceStepProps {
  sourcePlatform: string;
  setSourcePlatform: (value: string) => void;
  sku: string;
  setSku: (value: string) => void;
  onSearch: () => void;
  isLoadingProduct: boolean;
  productError: Error | null;
  productData: ProductData | undefined;
}

export interface CloneTargetStepProps {
  targetPlatform: string;
  setTargetPlatform: (value: string) => void;
  sourcePlatform: string;
  isLoadingTargets: boolean;
  targetsError: Error | null;
  targetsData: CloneTargetsResult | undefined;
}

export interface CloneConfigurationStepProps {
  previewError: Error | null;
  productData: ProductData | undefined;
  previewData: ConflictResult | undefined;
  targetPlatform: string;
  isLoadingPreview: boolean;
  setNewPrice: (value: number | undefined) => void;
  saveAsDraft: boolean;
  setSaveAsDraft: (value: boolean) => void;
}

export interface CloneResultStepProps {
  isCloning: boolean;
  isCloneSuccess: boolean;
  isCloneError: boolean;
  cloneError: Error | null;
  targetPlatform: string;
  sourcePlatform: string;
  sku: string;
  newPrice: number | undefined;
  productData: ProductData | undefined;
  saveAsDraft: boolean;
  onClose: () => void;
  onReset: () => void;
  onRetry: () => void;
  onBack: () => void;
}
