/**
 * Test Clone SKU - Check SKU exists across platforms
 * Usage: npx ts-node scripts/test-clone-sku.ts UNVIX4610
 */

import { PrismaClient } from '@prisma/client';

async function testCloneSku(sku: string) {
  const prisma = new PrismaClient();
  
  console.log('='.repeat(60));
  console.log('🔍 Testing Clone for SKU:', sku);
  console.log('='.repeat(60));
  
  try {
    // Check Shopee
    console.log('\n📦 SHOPEE:');
    const shopeeSku = await prisma.shopeeSku.findFirst({
      where: { sellerSku: sku },
      include: { product: true }
    });
    
    if (shopeeSku) {
      console.log('  ✅ Found in Shopee');
      console.log('  - Product:', (shopeeSku.product as any)?.name || (shopeeSku.product as any)?.itemName || 'N/A');
      console.log('  - Price:', shopeeSku.price);
      console.log('  - Stock:', shopeeSku.stock);
      if ((shopeeSku.product as any)?.images) {
        try {
          const images = JSON.parse((shopeeSku.product as any).images);
          console.log('  - Images:', images.length, 'image(s)');
        } catch(e) {}
      }
    } else {
      console.log('  ❌ Not found in Shopee');
    }
    
    // Check Lazada
    console.log('\n📦 LAZADA:');
    const lazadaSku = await prisma.lazadaSku.findFirst({
      where: { OR: [{ skuId: sku }, { sellerSku: sku }] },
      include: { product: true }
    });
    
    if (lazadaSku) {
      console.log('  ✅ Found in Lazada');
      console.log('  - Product:', (lazadaSku.product as any)?.name || 'N/A');
      console.log('  - Price:', lazadaSku.price);
      console.log('  - Stock:', lazadaSku.stock || (lazadaSku as any).quantity);
      if ((lazadaSku.product as any)?.images) {
        try {
          const images = JSON.parse((lazadaSku.product as any).images);
          console.log('  - Images:', images.length, 'image(s)');
        } catch(e) {}
      }
    } else {
      console.log('  ❌ Not found in Lazada');
    }
    
    // Check TikTok
    console.log('\n📦 TIKTOK:');
    const tiktokSku = await prisma.tiktokSku.findFirst({
      where: { OR: [{ skuId: sku }, { sellerSku: sku }] },
      include: { product: true }
    });
    
    if (tiktokSku) {
      console.log('  ✅ Found in TikTok');
      console.log('  - Product:', (tiktokSku.product as any)?.title || (tiktokSku.product as any)?.name || 'N/A');
      console.log('  - Price:', tiktokSku.price);
      console.log('  - Stock:', tiktokSku.stock || (tiktokSku as any).availableStock);
      if ((tiktokSku.product as any)?.images) {
        try {
          const images = JSON.parse((tiktokSku.product as any).images);
          console.log('  - Images:', images.length, 'image(s)');
        } catch(e) {}
      }
    } else {
      console.log('  ❌ Not found in TikTok');
    }
    
    // Summary
    console.log('\n' + '='.repeat(60));
    console.log('📋 CLONE STATUS SUMMARY:');
    console.log('='.repeat(60));
    
    const sources: string[] = [];
    const targets: string[] = [];
    
    if (shopeeSku) sources.push('Shopee'); else targets.push('Shopee');
    if (lazadaSku) sources.push('Lazada'); else targets.push('Lazada');
    if (tiktokSku) sources.push('TikTok'); else targets.push('TikTok');
    
    console.log('✅ Sources (produk ada):', sources.length > 0 ? sources.join(', ') : 'None');
    console.log('🎯 Targets (bisa di-clone ke):', targets.length > 0 ? targets.join(', ') : 'None - already on all platforms');
    
    if (sources.length > 0 && targets.length > 0) {
      console.log('\n💡 Clone Scenario:');
      for (const source of sources) {
        for (const target of targets) {
          console.log(`   ${source} → ${target}`);
        }
      }
    } else if (sources.length === 0) {
      console.log('\n⚠️ SKU tidak ditemukan di platform manapun!');
    } else {
      console.log('\n✅ Produk sudah ada di semua platform!');
    }
    
    await prisma.$disconnect();
    
    return { sources, targets, shopeeSku, lazadaSku, tiktokSku };
    
  } catch(e: any) {
    console.error('\n❌ Error:', e.message);
    await prisma.$disconnect();
    throw e;
  }
}

// Run if called directly
const sku = process.argv[2] || 'UNVIX4610';
testCloneSku(sku)
  .then(() => process.exit(0))
  .catch(() => process.exit(1));
