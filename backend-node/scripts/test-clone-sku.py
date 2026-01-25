"""
Test Clone SKU - Check SKU exists across platforms
"""
import sqlite3

DB_PATH = "config/databases/yumna_bertigamart.db"

def test_clone_sku(sku):
    print("=" * 60)
    print(f"Testing Clone for SKU: {sku}")
    print("=" * 60)
    
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    sources = []
    targets = []
    
    # Check Shopee
    print("\nSHOPEE:")
    try:
        cursor.execute("SELECT id, productId, sellerSku, variantName, price, quantity FROM ShopeeSku WHERE sellerSku = ?", (sku,))
        shopee = cursor.fetchone()
        if shopee:
            print(f"  FOUND!")
            print(f"  - ID: {shopee[0]}, ProductID: {shopee[1]}")
            print(f"  - SellerSku: {shopee[2]}")
            print(f"  - Variant: {shopee[3]}")
            print(f"  - Price: {shopee[4]}, Qty: {shopee[5]}")
            sources.append("Shopee")
        else:
            print("  NOT FOUND")
            targets.append("Shopee")
    except Exception as e:
        print(f"  Error: {e}")
        targets.append("Shopee")
    
    # Check Lazada
    print("\nLAZADA:")
    try:
        cursor.execute("SELECT id, productId, sellerSku, skuId, price, quantity FROM LazadaSku WHERE sellerSku = ? OR skuId = ?", (sku, sku))
        lazada = cursor.fetchone()
        if lazada:
            print(f"  FOUND!")
            print(f"  - ID: {lazada[0]}, ProductID: {lazada[1]}")
            print(f"  - SellerSku: {lazada[2]}, SkuId: {lazada[3]}")
            print(f"  - Price: {lazada[4]}, Qty: {lazada[5]}")
            sources.append("Lazada")
        else:
            print("  NOT FOUND")
            targets.append("Lazada")
    except Exception as e:
        print(f"  Error: {e}")
        targets.append("Lazada")
    
    # Check TikTok
    print("\nTIKTOK:")
    try:
        cursor.execute("SELECT id, productId, sellerSku, skuId, price, stock FROM TiktokSku WHERE sellerSku = ? OR skuId = ?", (sku, sku))
        tiktok = cursor.fetchone()
        if tiktok:
            print(f"  FOUND!")
            print(f"  - ID: {tiktok[0]}, ProductID: {tiktok[1]}")
            print(f"  - SellerSku: {tiktok[2]}, SkuId: {tiktok[3]}")
            print(f"  - Price: {tiktok[4]}, Stock: {tiktok[5]}")
            sources.append("TikTok")
        else:
            print("  NOT FOUND")
            targets.append("TikTok")
    except Exception as e:
        print(f"  Error: {e}")
        targets.append("TikTok")
    
    conn.close()
    
    # Summary
    print("\n" + "=" * 60)
    print("CLONE STATUS:")
    print("=" * 60)
    print(f"Sources: {', '.join(sources) if sources else 'None'}")
    print(f"Targets: {', '.join(targets) if targets else 'All platforms have it'}")
    
    if sources and targets:
        print("\nClone Options:")
        for src in sources:
            for tgt in targets:
                print(f"  {src} -> {tgt}")
    
    return sources, targets

if __name__ == "__main__":
    import sys
    sku = sys.argv[1] if len(sys.argv) > 1 else "UNVIX4610"
    test_clone_sku(sku)
