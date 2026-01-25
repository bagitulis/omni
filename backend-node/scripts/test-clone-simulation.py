"""
Test Product Cloning Simulation (Fixed)
Fetch product from source platform and prepare for Add Product on target platforms
Uses correct database schema: ShopeeProduct.image (single URL) not images
"""
import sqlite3
import json
import sys

DB_PATH = "config/databases/yumna_bertigamart.db"

def fetch_shopee_product_for_clone(sku):
    """Fetch complete product data from Shopee database for cloning"""
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    cursor = conn.cursor()
    
    # Get SKU with product data - using correct column names from schema
    cursor.execute("""
        SELECT 
            s.id as sku_id,
            s.productId,
            s.itemId,
            s.modelId,
            s.sellerSku,
            s.variantName,
            s.variantData,
            s.price,
            s.quantity,
            p.id as product_id,
            p.itemId as product_itemId,
            p.name,
            p.description,
            p.status,
            p.price as product_price,
            p.quantity as product_quantity,
            p.image
        FROM ShopeeSku s
        LEFT JOIN ShopeeProduct p ON s.productId = p.id
        WHERE s.sellerSku = ?
    """, (sku,))
    
    row = cursor.fetchone()
    if not row:
        conn.close()
        return None
    
    # Get all SKUs for this product (variants)
    product_id = row['productId']
    cursor.execute("""
        SELECT sellerSku, variantName, variantData, price, quantity, modelId
        FROM ShopeeSku
        WHERE productId = ?
    """, (product_id,))
    all_skus = cursor.fetchall()
    
    conn.close()
    
    # Build clone data structure
    images = []
    if row['image']:
        # image is stored as single URL or JSON array
        try:
            parsed = json.loads(row['image'])
            if isinstance(parsed, list):
                images = parsed
            else:
                images = [row['image']]
        except:
            images = [row['image']]
    
    clone_data = {
        "source_platform": "Shopee",
        "source_product_id": product_id,
        "source_item_id": str(row['product_itemId']),
        "title": row['name'] or "",
        "description": row['description'] or "",
        "images": images,
        "skus": [],
        "status": row['status'],
    }
    
    # Process SKUs/variants
    for s in all_skus:
        clone_data["skus"].append({
            "seller_sku": s['sellerSku'],
            "variant_name": s['variantName'],
            "model_id": str(s['modelId']) if s['modelId'] else None,
            "price": s['price'],
            "stock": s['quantity']
        })
    
    return clone_data

def transform_for_lazada(clone_data):
    """Transform clone data to Lazada Add Product format"""
    return {
        "platform": "Lazada",
        "api_endpoint": "POST /product/create",
        "payload": {
            "PrimaryCategory": "<user_selects>",
            "Attributes": {
                "name": clone_data["title"],
                "description": clone_data["description"],
                "brand_id": "<optional>",
            },
            "Images": {
                "Image": clone_data["images"][:8]
            },
            "Skus": {
                "Sku": [
                    {
                        "SellerSku": sku["seller_sku"],
                        "price": sku["price"],
                        "quantity": sku["stock"],
                    }
                    for sku in clone_data["skus"]
                ]
            }
        }
    }

def transform_for_tiktok(clone_data):
    """Transform clone data to TikTok Shop Add Product format"""
    return {
        "platform": "TikTok Shop",
        "api_endpoint": "POST /product/202309/products",
        "payload": {
            "title": clone_data["title"],
            "description": clone_data["description"],
            "category_id": "<user_selects>",
            "main_images": [{"uri": img} for img in clone_data["images"][:9]],
            "skus": [
                {
                    "seller_sku": sku["seller_sku"],
                    "price": {"amount": str(int(sku["price"] * 100)), "currency": "IDR"},
                    "inventory": [{"quantity": sku["stock"]}]
                }
                for sku in clone_data["skus"]
            ]
        }
    }

def test_clone_simulation(sku):
    print("=" * 70)
    print(f"PRODUCT CLONING SIMULATION - SKU: {sku}")
    print("=" * 70)
    
    # Step 1: Fetch from source (Shopee)
    print("\n[STEP 1] Fetching product from Shopee database...")
    clone_data = fetch_shopee_product_for_clone(sku)
    
    if not clone_data:
        print(f"  ERROR: SKU {sku} not found in Shopee!")
        return None
    
    print(f"  Source: {clone_data['source_platform']}")
    print(f"  Product ID: {clone_data['source_product_id']}")
    print(f"  Item ID: {clone_data['source_item_id']}")
    print(f"  Title: {clone_data['title']}")
    desc_preview = clone_data['description'][:100] + "..." if len(clone_data['description']) > 100 else clone_data['description']
    print(f"  Description: {desc_preview}")
    print(f"  Status: {clone_data['status']}")
    print(f"  Images: {len(clone_data['images'])} image(s)")
    for i, img in enumerate(clone_data['images'][:3]):
        print(f"    [{i+1}] {img[:60]}...")
    print(f"  Variants/SKUs: {len(clone_data['skus'])}")
    for i, sku_data in enumerate(clone_data['skus']):
        print(f"    [{i+1}] {sku_data['seller_sku']} | {sku_data['variant_name']} | Rp {sku_data['price']:,.0f} | Stock: {sku_data['stock']}")
    
    # Step 2: Transform for Lazada
    print("\n" + "-" * 70)
    print("[STEP 2] Transform for LAZADA")
    print("-" * 70)
    lazada = transform_for_lazada(clone_data)
    print(f"  API: {lazada['api_endpoint']}")
    print(f"  Name: {lazada['payload']['Attributes']['name']}")
    print(f"  Images: {len(lazada['payload']['Images']['Image'])} (max 8)")
    print(f"  SKUs: {len(lazada['payload']['Skus']['Sku'])}")
    print("  Category: User selection required")
    
    # Step 3: Transform for TikTok
    print("\n" + "-" * 70)
    print("[STEP 3] Transform for TIKTOK SHOP")
    print("-" * 70)
    tiktok = transform_for_tiktok(clone_data)
    print(f"  API: {tiktok['api_endpoint']}")
    print(f"  Title: {tiktok['payload']['title']}")
    print(f"  Images: {len(tiktok['payload']['main_images'])} (max 9)")
    print(f"  SKUs: {len(tiktok['payload']['skus'])}")
    print("  Category: User selection required")
    
    # Summary
    print("\n" + "=" * 70)
    print("CLONE SUMMARY")
    print("=" * 70)
    print(f"Product: {clone_data['title']}")
    print(f"From: Shopee (ItemID: {clone_data['source_item_id']})")
    print(f"To: Lazada, TikTok Shop")
    print(f"\nData Ready:")
    print(f"  [{'X' if clone_data['title'] else ' '}] Title")
    print(f"  [{'X' if clone_data['description'] else ' '}] Description ({len(clone_data['description'])} chars)")
    print(f"  [{'X' if clone_data['images'] else ' '}] Images ({len(clone_data['images'])})")
    print(f"  [{'X' if clone_data['skus'] else ' '}] SKUs/Variants ({len(clone_data['skus'])})")
    print(f"\nManual Steps:")
    print("  1. Select category on target platform")
    print("  2. Fill required category attributes")
    print("  3. Upload images to target CDN")
    print("  4. Submit Add Product")
    
    return clone_data

if __name__ == "__main__":
    sku = sys.argv[1] if len(sys.argv) > 1 else "UNVIX4610"
    test_clone_simulation(sku)
