/**
 * SVG Icon Optimizer
 * SVG icons are already optimized, but can be compressed further
 *
 * Untuk optimize SVG icons:
 * 1. Use SVGO (SVG Optimizer)
 * 2. Inline SVG untuk critical icons
 * 3. Use CSS untuk styling
 */

/**
 * Create optimized SVG icon path
 * SVG icons lebih baik daripada PNG/JPG untuk icons karena:
 * - Scalable (tidak pixelated)
 * - Smaller file size
 * - Can be styled dengan CSS
 */
export function getIconPath(
  iconName: string,
  variant: "solid" | "outline" = "solid"
): string {
  return `/src/assets/icons/${variant}/${iconName}.svg`;
}

/**
 * Alternative: Generate data URL SVG inline (untuk critical icons)
 * Mengurangi HTTP requests
 */
export const ICON_DATA_URLS = {
  shopee: `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%23EE0011' d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z'/%3E%3C/svg%3E`,

  lazada: `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%2326A1D5' d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z'/%3E%3C/svg%3E`,

  tiktok: `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%23000000' d='M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z'/%3E%3C/svg%3E`,

  settings: `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24'%3E%3Cpath fill='%23666' d='M19.14 12.94c.04-.3.06-.61.06-.94 0-.32-.02-.64-.07-.94l2.03-1.58c.18-.14.23-.41.12-.62l-1.92-3.32c-.12-.22-.37-.29-.59-.22l-2.39.96c-.5-.38-1.03-.7-1.62-.94l-.36-2.54c-.04-.24-.24-.41-.48-.41h-3.84c-.24 0-.43.17-.47.41l-.36 2.54c-.59.24-1.13.57-1.62.94l-2.39-.96c-.22-.08-.47 0-.59.22L2.74 8.87c-.12.21-.08.48.1.62l2.03 1.58c-.05.3-.07.62-.07.94s.02.64.07.94l-2.03 1.58c-.18.14-.23.41-.12.62l1.92 3.32c.12.22.37.29.59.22l2.39-.96c.5.38 1.03.7 1.62.94l.36 2.54c.05.24.24.41.48.41h3.84c.24 0 .44-.17.47-.41l.36-2.54c.59-.24 1.13-.56 1.62-.94l2.39.96c.22.08.47 0 .59-.22l1.92-3.32c.12-.22.07-.48-.1-.62l-2.03-1.58zM12 15.6c-1.98 0-3.6-1.62-3.6-3.6s1.62-3.6 3.6-3.6 3.6 1.62 3.6 3.6-1.62 3.6-3.6 3.6z'/%3E%3C/svg%3E`,
};

/**
 * Get SVG icon dengan fallback to data URL
 */
export function getSvgIcon(
  iconName: string,
  options?: {
    inline?: boolean; // Use data URL for critical icons
    classes?: string;
    size?: number; // width dan height
    fill?: string; // override color
  }
): {
  src: string;
  type: "file" | "inline";
  isSvg: true;
} {
  const useInline =
    options?.inline && ICON_DATA_URLS[iconName as keyof typeof ICON_DATA_URLS];

  return {
    src: useInline
      ? ICON_DATA_URLS[iconName as keyof typeof ICON_DATA_URLS]
      : getIconPath(iconName),
    type: useInline ? "inline" : "file",
    isSvg: true,
  };
}

/**
 * Generate optimized SVG tag
 */
export function generateSvgTag(
  iconName: string,
  options?: {
    alt?: string;
    classes?: string;
    size?: number;
    title?: string;
  }
): string {
  const size = options?.size || 24;
  const classes = options?.classes || "";
  const icon = getSvgIcon(iconName, { inline: true });

  return `
    <img
      src="${icon.src}"
      alt="${options?.alt || iconName}"
      title="${options?.title || iconName}"
      class="svg-icon ${classes}"
      width="${size}"
      height="${size}"
      loading="lazy"
      decoding="async"
    />
  `;
}

/**
 * Recommended icon sizes untuk different use cases
 */
export const ICON_SIZES = {
  xs: 16, // badges, inline labels
  sm: 20, // buttons, small UI
  md: 24, // standard icons
  lg: 32, // large buttons
  xl: 48, // hero sections
  xxl: 64, // banners
} as const;

/**
 * SVG Optimization Tips:
 *
 * 1. Use SVGO untuk minify SVG files
 *    npm install -g svgo
 *    svgo icons/*.svg --folder=icons-optimized
 *
 * 2. Inline critical SVG icons (brand logos, navigation)
 *    Reduce HTTP requests
 *
 * 3. Use CSS untuk style SVG
 *    .icon { fill: currentColor; }
 *
 * 4. Use viewBox untuk responsive scaling
 *    <svg viewBox="0 0 24 24"> ... </svg>
 *
 * 5. Preload critical SVG icons
 *    <link rel="preload" href="/icons/shopee.svg" as="image">
 */
