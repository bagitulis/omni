/**
 * Image Optimizer Utility
 * Provides optimized image handling with WEBP fallback and responsive images
 */

interface ResponsiveImageSet {
  src: string;
  srcSet: string;
  sizes: string;
  alt: string;
  title?: string;
  loading?: "lazy" | "eager";
  decoding?: "auto" | "sync" | "async";
  fetchPriority?: "high" | "low" | "auto";
}

interface ImageBreakpoint {
  width: number;
  scale?: number;
}

/**
 * Generate WEBP versions of image with fallback
 * Usage: For production, pre-convert images to WEBP format
 */
export function getImageWithWebP(imagePath: string): {
  webp: string;
  fallback: string;
} {
  const pathWithoutExt = imagePath.replace(/\.[^/.]+$/, "");
  return {
    webp: `${pathWithoutExt}.webp`,
    fallback: imagePath,
  };
}

/**
 * Generate responsive image srcset for different screen sizes
 * Automatically creates 1x and 2x versions
 */
export function generateResponsiveImageSrcSet(
  imagePath: string,
  breakpoints: ImageBreakpoint[] = [
    { width: 320, scale: 1 },
    { width: 320, scale: 2 },
    { width: 640, scale: 1 },
    { width: 640, scale: 2 },
    { width: 1024, scale: 1 },
    { width: 1024, scale: 2 },
  ]
): string {
  const pathWithoutExt = imagePath.replace(/\.[^/.]+$/, "");
  const ext = imagePath.match(/\.[^/.]+$/)?.[0] || ".png";

  return breakpoints
    .map((bp) => {
      const filename = `${pathWithoutExt}-${bp.width}w${ext}`;
      const descriptor = bp.scale ? `${bp.width}w` : "1x";
      return `${filename} ${descriptor}`;
    })
    .join(", ");
}

/**
 * Generate sizes attribute for responsive images
 * Defines image size at different viewport widths
 */
export function generateSizesAttribute(
  defaultSize: string = "100vw",
  breakpoints?: Array<{ maxWidth: string; size: string }>
): string {
  if (!breakpoints || breakpoints.length === 0) {
    return defaultSize;
  }

  return breakpoints
    .map((bp) => `(max-width: ${bp.maxWidth}) ${bp.size}`)
    .concat(defaultSize)
    .join(", ");
}

/**
 * Create optimized responsive image configuration
 * Combines all optimization techniques
 */
export function createResponsiveImage(
  imagePath: string,
  alt: string,
  options?: {
    title?: string;
    loading?: "lazy" | "eager";
    decoding?: "auto" | "sync" | "async";
    fetchPriority?: "high" | "low" | "auto";
    breakpoints?: ImageBreakpoint[];
    sizes?: string;
    isLCP?: boolean; // Mark as Largest Contentful Paint element
  }
): ResponsiveImageSet {
  const { fallback } = getImageWithWebP(imagePath);
  const srcSet = generateResponsiveImageSrcSet(fallback, options?.breakpoints);

  // LCP images should load eagerly with high priority
  const isLCPImage = options?.isLCP === true;

  return {
    src: fallback, // Fallback untuk browser yang tidak support picture element
    srcSet: srcSet,
    sizes:
      options?.sizes ||
      generateSizesAttribute("100vw", [
        { maxWidth: "640px", size: "90vw" },
        { maxWidth: "1024px", size: "80vw" },
      ]),
    alt,
    title: options?.title || alt,
    loading: isLCPImage ? "eager" : (options?.loading || "lazy"),
    decoding: isLCPImage ? "sync" : (options?.decoding || "async"),
    fetchPriority: isLCPImage ? "high" : (options?.fetchPriority || "auto"),
  };
}

/**
 * Create picture element source set untuk WEBP dengan fallback
 * Lebih optimal daripada srcset biasa
 */
export function getPictureSourceSets(imagePath: string): {
  webpSrcSet: string;
  jpgSrcSet: string;
} {
  const pathWithoutExt = imagePath.replace(/\.[^/.]+$/, "");

  const webpSrcSet = [
    `${pathWithoutExt}-320w.webp 320w`,
    `${pathWithoutExt}-640w.webp 640w`,
    `${pathWithoutExt}-1024w.webp 1024w`,
  ].join(", ");

  const jpgSrcSet = [
    `${pathWithoutExt}-320w.jpg 320w`,
    `${pathWithoutExt}-640w.jpg 640w`,
    `${pathWithoutExt}-1024w.jpg 1024w`,
  ].join(", ");

  return { webpSrcSet, jpgSrcSet };
}

/**
 * Generate lazy loading image component markup
 * Untuk use di template dengan v-html (not recommended, use proper component instead)
 */
export function generateImageHTML(
  imagePath: string,
  alt: string,
  classes?: string
): string {
  const { src, srcSet, sizes } = createResponsiveImage(imagePath, alt);

  return `
    <picture>
      <source srcset="${srcSet}" type="image/webp" sizes="${sizes}">
      <img
        src="${src}"
        alt="${alt}"
        loading="lazy"
        decoding="async"
        class="${classes || ""}"
      />
    </picture>
  `;
}
