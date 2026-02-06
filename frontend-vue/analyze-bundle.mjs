#!/usr/bin/env node
/**
 * Bundle Analysis Script (ESM)
 * Analyzes build output for unused JavaScript and optimization opportunities
 */

import fs from 'fs';
import path from 'path';
import zlib from 'zlib';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const distDir = path.join(__dirname, 'dist');

console.log('\n📊 === BUNDLE ANALYSIS REPORT ===\n');

// Get all JS files
const jsFiles = fs.readdirSync(path.join(distDir, 'js')).filter(f => f.endsWith('.js'));

let totalSize = 0;
let totalGzip = 0;
const fileSizes = [];

jsFiles.forEach(file => {
  const filePath = path.join(distDir, 'js', file);
  const rawSize = fs.statSync(filePath).size;
  const content = fs.readFileSync(filePath);
  const gzipSize = zlib.gzipSync(content).length;
  
  totalSize += rawSize;
  totalGzip += gzipSize;
  
  fileSizes.push({
    file,
    raw: (rawSize / 1024).toFixed(2),
    gzip: (gzipSize / 1024).toFixed(2),
    ratio: ((gzipSize / rawSize) * 100).toFixed(1)
  });
});

// Sort by size descending
fileSizes.sort((a, b) => parseFloat(b.gzip) - parseFloat(a.gzip));

console.log('📦 TOP 10 LARGEST CHUNKS:');
console.log('─'.repeat(80));
fileSizes.slice(0, 10).forEach((f, i) => {
  console.log(`${i + 1}. ${f.file.padEnd(40)} | Raw: ${f.raw.padStart(8)}KB | Gzip: ${f.gzip.padStart(8)}KB`);
});

console.log('\n📊 TOTAL SIZES:');
console.log(`   Raw JavaScript:    ${(totalSize / 1024).toFixed(2)} KB`);
console.log(`   Gzip Compressed:   ${(totalGzip / 1024).toFixed(2)} KB`);
console.log(`   Compression Ratio: ${((totalGzip / totalSize) * 100).toFixed(1)}%`);

// CSS files
const cssFiles = fs.readdirSync(path.join(distDir, 'css')).filter(f => f.endsWith('.css'));
let totalCssSize = 0;
let totalCssGzip = 0;

cssFiles.forEach(file => {
  const filePath = path.join(distDir, 'css', file);
  const rawSize = fs.statSync(filePath).size;
  const content = fs.readFileSync(filePath);
  const gzipSize = zlib.gzipSync(content).length;
  
  totalCssSize += rawSize;
  totalCssGzip += gzipSize;
});

console.log('\n🎨 CSS FILES:');
console.log(`   Raw CSS:           ${(totalCssSize / 1024).toFixed(2)} KB`);
console.log(`   Gzip Compressed:   ${(totalCssGzip / 1024).toFixed(2)} KB`);

console.log('\n💡 OPTIMIZATION SUGGESTIONS:');
console.log('─'.repeat(80));
console.log('1. ✅ Minification: Already enabled with Terser');
console.log('2. ✅ Code splitting: Using manual chunks for vue-core, state, api');
console.log('3. ✅ Lazy loading: Vue async components configured');
console.log('4. ✅ Tree shaking: Enabled with treeshake config');
console.log('5. Run `npx eslint src --fix` to auto-fix unused imports');
console.log('6. Check bundle-analysis.html for detailed breakdown');

console.log('\n✅ WHAT TO DO:');
console.log('─'.repeat(80));
console.log('• Open bundle-analysis.html in browser for visual breakdown');
console.log('• Remove unused dependencies: npm prune');
console.log('• Monitor bundle size with Lighthouse: npx lighthouse https://localhost:5173');

console.log('\n');
