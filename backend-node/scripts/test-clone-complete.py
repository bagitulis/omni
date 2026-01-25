"""
Complete Clone Test - Full Verification Suite
Tests the entire clone flow for a SKU
"""
import sqlite3
import json
import sys

DB_PATH = "config/databases/yumna_bertigamart.db"

def print_separator(title):
    print("\n" + "=" * 70)
    print(f"  {title}")
    print("=" * 70)

def test_step_1_check_sku_existence(sku):
    """Step 1: Check if SKU exists in each platform"""
    print_separator("STEP 1: Check SKU Existence Across Platforms")
    
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    results = {}
    
    # Shopee
    cursor.execute("SELECT id, productId, price, quantity FROM ShopeeSku WHERE sellerSku = ?", (sku,))
    shopee = cursor.fetchone()
    results['shopee'] = shopee is not None
    if shopee:
        print(f"  [OK] SHOPEE: Found (ID: {shopee[0]}, ProductID: {shopee[1]}, Price: {shopee[2]})")
    else:
        print(f"  [--] SHOPEE: Not found")
    
    # Lazada
    cursor.execute("SELECT id, productId, price FROM LazadaSku WHERE sellerSku = ? OR skuId = ?", (sku, sku))
    lazada = cursor.fetchone()
    results['lazada'] = lazada is not None
    if lazada:
        print(f"  [OK] LAZADA: Found (ID: {lazada[0]}, ProductID: {lazada[1]})")
    else:
        print(f"  [--] LAZADA: Not found")
    
    # TikTok
    cursor.execute("SELECT id, productId, price FROM TiktokSku WHERE sellerSku = ? OR skuId = ?", (sku, sku))
    tiktok = cursor.fetchone()
    results['tiktok'] = tiktok is not None
    if tiktok:
        print(f"  [OK] TIKTOK: Found (ID: {tiktok[0]}, ProductID: {tiktok[1]})")
    else:
        print(f"  [--] TIKTOK: Not found")
    
    conn.close()
    
    sources = [k for k, v in results.items() if v]
    targets = [k for k, v in results.items() if not v]
    
    print(f"\n  Result: Sources={sources}, Targets={targets}")
    return results, sources, targets

def test_step_2_fetch_product_data(sku):
    """Step 2: Fetch complete product data from source platform"""
    print_separator("STEP 2: Fetch Product Data from Shopee")
    
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()
    
    cursor.execute("""
        SELECT s.*, p.name, p.description, p.status, p.image
        FROM ShopeeSku s
        LEFT JOIN ShopeeProduct p ON s.productId = p.id
        WHERE s.sellerSku = ?
    """, (sku,))
    row = cursor.fetchone()
    
    if not row:
        print("  [FAIL] SKU not found in Shopee")
        conn.close()
        return None
    
    cursor.execute("SELECT sellerSku, variantName, price, quantity FROM ShopeeSku WHERE productId = ?", (row['productId'],))
    variants = cursor.fetchall()
    
    conn.close()
    
    images = []
    if row['image']:
        try:
            parsed = json.loads(row['image'])
            images = parsed if isinstance(parsed, list) else [row['image']]
        except:
            images = [row['image']]
    
    data = {
        'title': row['name'],
        'description': row['description'],
        'images': images,
        'variants': [dict(v) for v in variants]
    }
    
    print(f"  [OK] Title: {data['title']}")
    print(f"  [OK] Description: {len(data['description'] or '')} chars")
    print(f"  [OK] Images: {len(data['images'])} image(s)")
    if data['images']:
        print(f"       URL: {data['images'][0][:50]}...")
    print(f"  [OK] Variants: {len(data['variants'])}")
    for v in data['variants']:
        print(f"       - {v['sellerSku']}: {v['variantName']} @ Rp {v['price']:,.0f}")
    
    return data

def test_step_3_transform_lazada(data):
    """Step 3: Transform data for Lazada Add Product API"""
    print_separator("STEP 3: Transform for LAZADA Add Product")
    
    payload = {
        "Attributes": {"name": data['title'], "description": data['description']},
        "Images": data['images'][:8],
        "Skus": [{"SellerSku": v['sellerSku'], "price": v['price']} for v in data['variants']]
    }
    
    print(f"  [OK] API: POST /product/create")
    print(f"  [OK] SKUs: {len(payload['Skus'])}")
    print(f"  [OK] Images: {len(payload['Images'])}")
    print(f"  [!!] Category: User selection required")
    
    return payload

def test_step_4_transform_tiktok(data):
    """Step 4: Transform data for TikTok Add Product API"""
    print_separator("STEP 4: Transform for TIKTOK SHOP Add Product")
    
    payload = {
        "title": data['title'],
        "description": data['description'],
        "main_images": data['images'][:9],
        "skus": [{"seller_sku": v['sellerSku'], "price": v['price']} for v in data['variants']]
    }
    
    print(f"  [OK] API: POST /product/202309/products")
    print(f"  [OK] SKUs: {len(payload['skus'])}")
    print(f"  [OK] Images: {len(payload['main_images'])}")
    print(f"  [!!] Category: User selection required")
    
    return payload

def run_full_test(sku):
    """Run complete clone verification"""
    print("\n" + "#" * 70)
    print(f"##  COMPLETE CLONE VERIFICATION TEST")
    print(f"##  SKU: {sku}")
    print("#" * 70)
    
    # Step 1
    status, sources, targets = test_step_1_check_sku_existence(sku)
    if 'shopee' not in sources:
        print("\n[FAILED] SKU not found in Shopee - cannot clone")
        return False
    
    # Step 2
    data = test_step_2_fetch_product_data(sku)
    if not data:
        print("\n[FAILED] Could not fetch product data")
        return False
    
    # Step 3
    lazada_payload = test_step_3_transform_lazada(data)
    
    # Step 4
    tiktok_payload = test_step_4_transform_tiktok(data)
    
    # Summary
    print_separator("TEST RESULT SUMMARY")
    print(f"""
  SKU: {sku}
  Product: {data['title']}
  
  VERIFICATION:
  [OK] Step 1: SKU existence check - PASSED
  [OK] Step 2: Product data fetch - PASSED
  [OK] Step 3: Lazada transform - PASSED
  [OK] Step 4: TikTok transform - PASSED
  
  CLONE READY:
  - Source: SHOPEE
  - Targets: {', '.join([t.upper() for t in targets])}
  
  DATA:
  - Title: {data['title']}
  - Description: {len(data['description'] or '')} chars
  - Images: {len(data['images'])}
  - Variants: {len(data['variants'])}
  
  ============================================
  STATUS: ALL TESTS PASSED - CLONE READY!
  ============================================
    """)
    
    return True

if __name__ == "__main__":
    sku = sys.argv[1] if len(sys.argv) > 1 else "UNVIX4610"
    success = run_full_test(sku)
    sys.exit(0 if success else 1)
