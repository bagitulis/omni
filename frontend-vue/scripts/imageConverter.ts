import sharp from "sharp";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";
 
 
declare const process: any;

/**
 * Image Auto-Converter to WEBP
 * TypeScript script untuk convert images ke WEBP dengan responsive sizes
 * Run: npx tsx scripts/imageConverter.ts
 */

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const ASSETS_DIR = path.join(__dirname, "../src/assets");
const IMAGES_DIR = path.join(ASSETS_DIR, "images");
const ICONS_DIR = path.join(ASSETS_DIR, "icons");

// Responsive breakpoints untuk convert
const BREAKPOINTS = [320, 640, 1024];

// Supported image formats
const SUPPORTED_FORMATS = [".jpg", ".jpeg", ".png"];

interface ConversionStats {
  total: number;
  converted: number;
  skipped: number;
  errors: number;
  totalSizeBefore: number;
  totalSizeAfter: number;
}

/**
 * Convert single image to WEBP
 */
async function convertImageToWebP(
  inputPath: string,
  outputPath: string,
  quality: number = 75
): Promise<boolean> {
  try {
    await sharp(inputPath).webp({ quality }).toFile(outputPath);
    return true;
  } catch (error) {
    console.error(`❌ Error converting ${inputPath}:`, error);
    return false;
  }
}

/**
 * Resize image untuk responsive breakpoints
 */
async function resizeImageForBreakpoint(
  inputPath: string,
  width: number,
  outputDir: string,
  outputName: string
): Promise<boolean> {
  try {
    const nameWithoutExt = path.parse(outputName).name;
    const outputPath = path.join(outputDir, `${nameWithoutExt}-${width}w.webp`);

    await sharp(inputPath)
      .resize(width, null, { withoutEnlargement: true })
      .webp({ quality: 75 })
      .toFile(outputPath);

    return true;
  } catch (error) {
    console.error(`❌ Error resizing ${outputName} to ${width}w:`, error);
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
 * Process all images in directory
 */
async function processDirectory(
  dirPath: string,
  options: {
    createBreakpoints: boolean;
    quality: number;
  } = { createBreakpoints: true, quality: 75 }
): Promise<ConversionStats> {
  const stats: ConversionStats = {
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
    const filePath = path.join(dirPath, file);
    const ext = path.extname(file).toLowerCase();

    // Skip if not supported format
    if (!SUPPORTED_FORMATS.includes(ext)) {
      continue;
    }

    stats.total++;
    const originalSize = getFileSizeKB(filePath);
    stats.totalSizeBefore += originalSize;

    // Check if WEBP already exists
    const webpPath = path.join(dirPath, `${path.parse(file).name}.webp`);
    if (fs.existsSync(webpPath)) {
      console.log(`⏭️  Skipped ${file} (already converted)`);
      stats.skipped++;
      const webpSize = getFileSizeKB(webpPath);
      stats.totalSizeAfter += webpSize;
      continue;
    }

    // Convert main image
    console.log(`🔄 Converting ${file}...`);
    const converted = await convertImageToWebP(
      filePath,
      webpPath,
      options.quality
    );

    if (converted) {
      stats.converted++;
      const webpSize = getFileSizeKB(webpPath);
      stats.totalSizeAfter += webpSize;

      const reduction = (
        ((originalSize - webpSize) / originalSize) *
        100
      ).toFixed(1);
      console.log(
        `✅ ${file}: ${originalSize.toFixed(1)}KB → ${webpSize.toFixed(1)}KB (${reduction}% smaller)`
      );

      // Create responsive breakpoints
      if (options.createBreakpoints) {
        for (const width of BREAKPOINTS) {
          const success = await resizeImageForBreakpoint(
            filePath,
            width,
            dirPath,
            file
          );
          if (success) {
            console.log(`   ✓ Created ${path.parse(file).name}-${width}w.webp`);
          }
        }
      }
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
  console.log("\n🚀 Starting image auto-conversion to WEBP...\n");

  // Ensure directories exist
  [IMAGES_DIR, ICONS_DIR].forEach((dir) => {
    if (!fs.existsSync(dir)) {
      console.log(`📁 Creating directory: ${dir}`);
      fs.mkdirSync(dir, { recursive: true });
    }
  });

  // Process images directory with breakpoints
  console.log("📸 Processing images with responsive breakpoints...");
  const imagesStats = await processDirectory(IMAGES_DIR, {
    createBreakpoints: true,
    quality: 75,
  });

  // Process icons directory (no breakpoints needed)
  console.log("\n🎨 Processing icons...");
  const iconsStats = await processDirectory(ICONS_DIR, {
    createBreakpoints: false,
    quality: 80, // Higher quality untuk icons
  });

  // Combine stats
  const totalStats: ConversionStats = {
    total: imagesStats.total + iconsStats.total,
    converted: imagesStats.converted + iconsStats.converted,
    skipped: imagesStats.skipped + iconsStats.skipped,
    errors: imagesStats.errors + iconsStats.errors,
    totalSizeBefore: imagesStats.totalSizeBefore + iconsStats.totalSizeBefore,
    totalSizeAfter: imagesStats.totalSizeAfter + iconsStats.totalSizeAfter,
  };

  // Print summary
  console.log("\n" + "=".repeat(50));
  console.log("📊 CONVERSION SUMMARY");
  console.log("=".repeat(50));
  console.log(`Total images processed: ${totalStats.total}`);
  console.log(`✅ Successfully converted: ${totalStats.converted}`);
  console.log(`⏭️  Already converted: ${totalStats.skipped}`);
  console.log(`❌ Errors: ${totalStats.errors}`);
  console.log("");
  console.log(`Total size before: ${totalStats.totalSizeBefore.toFixed(2)}KB`);
  console.log(`Total size after:  ${totalStats.totalSizeAfter.toFixed(2)}KB`);

  if (totalStats.totalSizeBefore > 0) {
    const reduction = (
      ((totalStats.totalSizeBefore - totalStats.totalSizeAfter) /
        totalStats.totalSizeBefore) *
      100
    ).toFixed(1);
    console.log(`💾 Total reduction: ${reduction}%`);
  }

  console.log("=".repeat(50) + "\n");

  if (totalStats.converted > 0) {
    console.log("✨ Conversion complete! WEBP files ready to use.\n");
  }

  process.exit(totalStats.errors > 0 ? 1 : 0);
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
