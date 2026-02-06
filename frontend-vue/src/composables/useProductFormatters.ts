import { computed } from "vue";

export interface ProductBase {
  image_urls?: string;
  item_id?: number;
  item_name?: string;
  [key: string]: any;
}

export function useProductImages(productBase: any) {
  return computed<string[]>(() => {
    if (!productBase.value?.image_urls) return [];
    try {
      const urls = JSON.parse(productBase.value.image_urls);
      return Array.isArray(urls) ? urls : [];
    } catch {
      return [];
    }
  });
}

export function useProductVariations(variations: any) {
  return computed<Array<{ name: string; options: string[] }>>(() => {
    const grouped: Record<string, Set<string>> = {};
    variations.value.forEach((v: any) => {
      if (!grouped[v.variation_name]) {
        grouped[v.variation_name] = new Set();
      }
      if (v.option_name) {
        grouped[v.variation_name].add(v.option_name);
      }
    });

    return Object.entries(grouped).map(([name, options]) => ({
      name,
      options: Array.from(options),
    }));
  });
}

export function useProductFormatters() {
  const formatDate = (date: string | Date | null | undefined): string => {
    if (!date) return "-";
    return new Date(date).toLocaleDateString("id-ID");
  };

  const formatPrice = (price: number | null | undefined): string => {
    if (!price) return "0";
    return new Intl.NumberFormat("id-ID").format(price);
  };

  const parseAttributes = (
    attributeString: string | null | undefined
  ): Array<{ name: string; value: string }> => {
    if (!attributeString) return [];
    try {
      const attrs = JSON.parse(attributeString);
      return Array.isArray(attrs)
        ? attrs.map((a: any) => ({
            name: a.attribute_name || a.name || "",
            value: a.attribute_value || a.value || "",
          }))
        : [];
    } catch {
      return [];
    }
  };

  return { formatDate, formatPrice, parseAttributes };
}
