import sharp from "sharp";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
 
 
declare const process: any;

/**
 * SVG to WEBP Converter
 * Convert SVG icons ke WEBP format untuk optimal performa
 * Run: npx tsx scripts/svgToWebp.ts
 */

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ICONS_DIR = path.join(__dirname, "../src/assets/icons");

interface SVGConversionStats {
  total: number;
  converted: number;
  skipped: number;
  errors: number;
  totalSizeBefore: number;
  totalSizeAfter: number;
}

/**
 * Convert SVG to PNG then WEBP
 */
async function convertSvgToWebP(
  svgPath: string,
  outputPath: string,
  quality: number = 85
): Promise<boolean> {
  try {
    // Read SVG file
    const svgBuffer = fs.readFileSync(svgPath);

    // Convert SVG to PNG to WEBP
    await sharp(svgBuffer, { density: 150 }) // High DPI untuk crisp icons
      .png()
      .webp({ quality })
      .toFile(outputPath);

    return true;
  } catch (error) {
    console.error(`❌ Error converting ${svgPath}:`, error);
    return false;
  }
}

/**
 * Get file size in KB
 */
function getFileSizeKB(filePath: string): number {
  try {
    const stats = fs.statSync(filePath);
    return stats.size / 1024;
  } catch {
    return 0;
  }
}

/**
 * Process SVG files in directory
 */
async function processSvgDirectory(
  dirPath: string
): Promise<SVGConversionStats> {
  const stats: SVGConversionStats = {
    total: 0,
    converted: 0,
    skipped: 0,
    errors: 0,
    totalSizeBefore: 0,
    totalSizeAfter: 0,
  };

  if (!fs.existsSync(dirPath)) {
    console.log(`⚠️  Directory not found: ${dirPath}`);
    return stats;
  }

  const files = fs.readdirSync(dirPath);

  for (const file of files) {
    if (!file.endsWith(".svg")) continue;

    const filePath = path.join(dirPath, file);
    stats.total++;

    const originalSize = getFileSizeKB(filePath);
    stats.totalSizeBefore += originalSize;

    const webpPath = filePath.replace(/\.svg$/, ".webp");

    // Check if WEBP already exists
    if (fs.existsSync(webpPath)) {
      console.log(`⏭️  Skipped ${file} (already converted)`);
      stats.skipped++;
      const webpSize = getFileSizeKB(webpPath);
      stats.totalSizeAfter += webpSize;
      continue;
    }

    console.log(`🔄 Converting ${file}...`);
    const converted = await convertSvgToWebP(filePath, webpPath);

    if (converted) {
      stats.converted++;
      const webpSize = getFileSizeKB(webpPath);
      stats.totalSizeAfter += webpSize;

      const reduction = (
        ((originalSize - webpSize) / originalSize) *
        100
      ).toFixed(1);

      console.log(
        `✅ ${file} (SVG ${originalSize.toFixed(1)}KB) → ${webpSize.toFixed(1)}KB WEBP (${reduction}% reduction)`
      );
    } else {
      stats.errors++;
    }
  }

  return stats;
}

/**
 * Main execution
 */
async function main() {
  console.log("\n🎨 Starting SVG to WEBP conversion...\n");

  const stats = await processSvgDirectory(ICONS_DIR);

  console.log("\n" + "=".repeat(50));
  console.log("📊 SVG CONVERSION SUMMARY");
  console.log("=".repeat(50));
  console.log(`Total SVG files found: ${stats.total}`);
  console.log(`✅ Successfully converted: ${stats.converted}`);
  console.log(`⏭️  Already converted: ${stats.skipped}`);
  console.log(`❌ Errors: ${stats.errors}`);
  console.log("");
  console.log(`Original SVG size: ${stats.totalSizeBefore.toFixed(2)}KB`);
  console.log(`Converted WEBP size: ${stats.totalSizeAfter.toFixed(2)}KB`);

  if (stats.totalSizeBefore > 0) {
    const reduction = (
      ((stats.totalSizeBefore - stats.totalSizeAfter) / stats.totalSizeBefore) *
      100
    ).toFixed(1);
    console.log(`💾 Total reduction: ${reduction}%`);
  }

  console.log("=".repeat(50) + "\n");

  if (stats.converted > 0) {
    console.log("✨ SVG conversion complete! WEBP icons ready to use.\n");
  }

  process.exit(stats.errors > 0 ? 1 : 0);
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
