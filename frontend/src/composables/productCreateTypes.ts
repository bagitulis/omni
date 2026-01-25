/**
 * Types for Product Creation
 */

export interface Category {
  id: string;
  name: string;
  parentId?: string;
  isLeaf: boolean;
}

export interface CategoryAttribute {
  id: string;
  name: string;
  type: string;
  required: boolean;
  description?: string;
  values?: { id: string; name: string }[];
}

export interface Brand {
  id: string;
  name: string;
}

export interface DeliveryOption {
  id: string;
  name: string;
  description?: string;
}

export interface VariantType {
  name: string;
  values: string[];
}

export interface ProductSku {
  sellerSku: string;
  price: number;
  stock: number;
  variantLabel?: string;
}

export interface PackageDimensions {
  length: number;
  width: number;
  height: number;
}

export interface CreateProductForm {
  title: string;
  description: string;
  categoryId: string;
  brandId: string;
  attributes: Record<string, string>;
  images: string[];
  videoUrl: string;
  hasVariants: boolean;
  variantTypes: VariantType[];
  skus: ProductSku[];
  packageWeight: number;
  packageDimensions: PackageDimensions;
  deliveryOptionIds: string[];
  saveMode: "AS_DRAFT" | "LISTING";
}

export interface CreateProductResult {
  success: boolean;
  productId?: string;
  message: string;
}

export type Platform = "tiktok" | "shopee" | "lazada";
