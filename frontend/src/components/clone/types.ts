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
  productData: any; // Ideally this should be a specific type from @/types/clone or similar
}

export interface CloneTargetStepProps {
  targetPlatform: string;
  setTargetPlatform: (value: string) => void;
  sourcePlatform: string;
  isLoadingTargets: boolean;
  targetsError: Error | null;
  targetsData: any; // Ideally specific type
}

export interface CloneConfigurationStepProps {
  previewError: Error | null;
  productData: any;
  previewData: any;
  targetPlatform: string;
  isLoadingPreview: boolean;
  newPrice: number | undefined;
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
  productData: any;
  saveAsDraft: boolean;
  onClose: () => void;
  onReset: () => void;
  onRetry: () => void;
  onBack: () => void;
}
