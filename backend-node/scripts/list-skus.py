"""
List available SKUs from all platforms
"""

import sqlite3
import os

def get_db_path():
    paths = [
        "config/databases/yumna_bertigamart.db",
        "../config/databases/yumna_bertigamart.db",
        "backend/config/databases/yumna_bertigamart.db"
    ]
    for path in paths:
        if os.path.exists(path):
            return path
    return None

def list_skus():
    db_path = get_db_path()
    
    if not db_path:
        print("ERROR: Database not found!")
        return
    
    print(f"Database: {db_path}")
    print("=" * 60)
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Count SKUs per platform
    print("\nSKU COUNT PER PLATFORM:")
    print("-" * 40)
    
    try:
        cursor.execute("SELECT COUNT(*) FROM ShopeeSku")
        shopee_count = cursor.fetchone()[0]
        print(f"Shopee SKUs: {shopee_count}")
    except Exception as e:
        print(f"Shopee: Error - {e}")
        shopee_count = 0
    
    try:
        cursor.execute("SELECT COUNT(*) FROM LazadaSku")
        lazada_count = cursor.fetchone()[0]
        print(f"Lazada SKUs: {lazada_count}")
    except Exception as e:
        print(f"Lazada: Error - {e}")
        lazada_count = 0
    
    try:
        cursor.execute("SELECT COUNT(*) FROM TiktokSku")
        tiktok_count = cursor.fetchone()[0]
        print(f"TikTok SKUs: {tiktok_count}")
    except Exception as e:
        print(f"TikTok: Error - {e}")
        tiktok_count = 0
    
    # Sample SKUs from each platform
    print("\n" + "=" * 60)
    print("SAMPLE SKUs (first 10 from each platform):")
    print("=" * 60)
    
    # Shopee samples
    print("\nSHOPEE SKUs:")
    try:
        cursor.execute("SELECT sellerSku FROM ShopeeSku WHERE sellerSku IS NOT NULL AND sellerSku != '' LIMIT 10")
        for row in cursor.fetchall():
            print(f"  - {row[0]}")
    except Exception as e:
        print(f"  Error: {e}")
    
    # Lazada samples
    print("\nLAZADA SKUs:")
    try:
        cursor.execute("SELECT COALESCE(sellerSku, skuId) as sku FROM LazadaSku WHERE (sellerSku IS NOT NULL AND sellerSku != '') OR (skuId IS NOT NULL AND skuId != '') LIMIT 10")
        for row in cursor.fetchall():
            print(f"  - {row[0]}")
    except Exception as e:
        print(f"  Error: {e}")
    
    # TikTok samples
    print("\nTIKTOK SKUs:")
    try:
        cursor.execute("SELECT COALESCE(sellerSku, skuId) as sku FROM TiktokSku WHERE (sellerSku IS NOT NULL AND sellerSku != '') OR (skuId IS NOT NULL AND skuId != '') LIMIT 10")
        for row in cursor.fetchall():
            print(f"  - {row[0]}")
    except Exception as e:
        print(f"  Error: {e}")
    
    # Find SKUs that exist in multiple platforms (potential clones)
    print("\n" + "=" * 60)
    print("MATCHING SKUs (exist in multiple platforms):")
    print("=" * 60)
    
    try:
        # Get all Shopee SKUs
        cursor.execute("SELECT DISTINCT sellerSku FROM ShopeeSku WHERE sellerSku IS NOT NULL AND sellerSku != ''")
        shopee_skus = set(row[0] for row in cursor.fetchall())
        
        # Get all Lazada SKUs
        cursor.execute("SELECT DISTINCT COALESCE(sellerSku, skuId) FROM LazadaSku WHERE (sellerSku IS NOT NULL AND sellerSku != '') OR (skuId IS NOT NULL AND skuId != '')")
        lazada_skus = set(row[0] for row in cursor.fetchall())
        
        # Get all TikTok SKUs
        cursor.execute("SELECT DISTINCT COALESCE(sellerSku, skuId) FROM TiktokSku WHERE (sellerSku IS NOT NULL AND sellerSku != '') OR (skuId IS NOT NULL AND skuId != '')")
        tiktok_skus = set(row[0] for row in cursor.fetchall())
        
        # Find matches
        all_three = shopee_skus & lazada_skus & tiktok_skus
        shopee_lazada = (shopee_skus & lazada_skus) - all_three
        shopee_tiktok = (shopee_skus & tiktok_skus) - all_three
        lazada_tiktok = (lazada_skus & tiktok_skus) - all_three
        
        print(f"\nSKUs in ALL 3 platforms: {len(all_three)}")
        if all_three:
            for sku in list(all_three)[:5]:
                print(f"  - {sku}")
            if len(all_three) > 5:
                print(f"  ... and {len(all_three) - 5} more")
        
        print(f"\nSKUs in Shopee + Lazada only: {len(shopee_lazada)}")
        if shopee_lazada:
            for sku in list(shopee_lazada)[:3]:
                print(f"  - {sku}")
        
        print(f"\nSKUs in Shopee + TikTok only: {len(shopee_tiktok)}")
        if shopee_tiktok:
            for sku in list(shopee_tiktok)[:3]:
                print(f"  - {sku}")
        
        print(f"\nSKUs in Lazada + TikTok only: {len(lazada_tiktok)}")
        if lazada_tiktok:
            for sku in list(lazada_tiktok)[:3]:
                print(f"  - {sku}")
        
        # SKUs only in one platform (candidates for cloning)
        only_shopee = shopee_skus - lazada_skus - tiktok_skus
        only_lazada = lazada_skus - shopee_skus - tiktok_skus
        only_tiktok = tiktok_skus - shopee_skus - lazada_skus
        
        print(f"\nSKUs ONLY in Shopee (can clone to Lazada/TikTok): {len(only_shopee)}")
        if only_shopee:
            for sku in list(only_shopee)[:3]:
                print(f"  - {sku}")
        
        print(f"\nSKUs ONLY in Lazada (can clone to Shopee/TikTok): {len(only_lazada)}")
        if only_lazada:
            for sku in list(only_lazada)[:3]:
                print(f"  - {sku}")
        
        print(f"\nSKUs ONLY in TikTok (can clone to Shopee/Lazada): {len(only_tiktok)}")
        if only_tiktok:
            for sku in list(only_tiktok)[:3]:
                print(f"  - {sku}")
        
    except Exception as e:
        print(f"Error analyzing SKUs: {e}")
    
    conn.close()

if __name__ == "__main__":
    list_skus()
