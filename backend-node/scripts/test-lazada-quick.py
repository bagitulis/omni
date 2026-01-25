"""Quick Lazada Clone Test with Category Recommendation"""
import requests
import json

BASE_URL = "http://localhost:3000/api"
TENANT_ID = "yumna_bertigamart"

# Pre-generated token (bypasses login)
TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOjEsInVzZXJuYW1lIjoieXVtbmEiLCJyb2xlIjoib3duZXIiLCJ0ZW5hbnRJZCI6Inl1bW5hX2JlcnRpZ2FtYXJ0IiwiaWF0IjoxNzY4ODMyOTk0LCJleHAiOjE3Njg5MTkzOTR9.L2OhpuFo9ssOj4vulpN01zCKPJV89tRUwBBsYPey44Q"

def main():
    session = requests.Session()
    token = TOKEN
    print(f"[1] Using pre-generated token: {token[:30]}...")
    
    # Get CSRF
    session.get(f"{BASE_URL}/csrf-token", headers={"Authorization": f"Bearer {token}"})
    csrf = session.cookies.get("csrf_token", "")
    
    headers = {
        "Authorization": f"Bearer {token}",
        "x-tenant-id": TENANT_ID,
        "x-csrf-token": csrf,
        "Content-Type": "application/json"
    }
    
    # Test Lazada category recommendation
    print("\n[2] LAZADA CATEGORY RECOMMENDATION")
    title = "Vixal Pembersih Porselen Botol 470ml"
    resp = session.get(f"{BASE_URL}/lazada/categories/recommend?title={title}", headers=headers, timeout=30)
    print(f"  HTTP {resp.status_code}")
    print(f"  Response: {resp.text[:500]}")
    
    category_id = None
    if resp.status_code == 200:
        result = resp.json()
        if result.get("category"):
            category_id = result["category"].get("id") or result["category"].get("category_id")
            cat_name = result["category"].get("name", "N/A")
            print(f"  Recommended: {cat_name} (ID: {category_id})")
    
    if not category_id:
        # Fallback - search for cleaning category
        print("\n[2b] LAZADA CATEGORY SEARCH (fallback)")
        resp = session.get(f"{BASE_URL}/lazada/categories/search?keyword=pembersih kamar mandi", headers=headers, timeout=30)
        print(f"  HTTP {resp.status_code}")
        if resp.status_code == 200:
            result = resp.json()
            categories = result.get("categories", [])
            print(f"  Found {len(categories)} categories")
            for cat in categories[:5]:
                cat_id = cat.get("category_id") or cat.get("id")
                print(f"    - {cat.get('name', 'N/A')} (ID: {cat_id}, isLeaf: {cat.get('leaf', cat.get('isLeaf', 'N/A'))})")
                if cat.get("leaf") or cat.get("isLeaf"):
                    category_id = cat_id
                    break
    
    if not category_id:
        print("\n  ERROR: Could not find a valid category!")
        return
    
    # Test image migrate
    print(f"\n[3] LAZADA IMAGE MIGRATE")
    image_url = "https://cf.shopee.co.id/file/id-11134207-7qul4-lk2a2fniq09k3"
    resp = session.post(f"{BASE_URL}/lazada/images/upload", json={"image_url": image_url}, headers=headers, timeout=30)
    print(f"  HTTP {resp.status_code}")
    
    lazada_image = image_url
    if resp.status_code == 200:
        result = resp.json()
        if result.get("success") and result.get("imageUrl"):
            lazada_image = result["imageUrl"]
            print(f"  Migrated: {lazada_image[:60]}...")
        else:
            print(f"  Using original (migration failed): {result}")
    else:
        print(f"  Using original: {resp.text[:200]}")
    
    # Create product
    print(f"\n[4] LAZADA CREATE PRODUCT")
    payload = {
        "title": "Vixal Pembersih Porselen Botol 470ml",
        "name": "Vixal Pembersih Porselen Botol 470ml",
        "description": "<p>Vixal Pembersih Porslen Biru Botol 470ml. Vixal Pembersih Kamar Mandi Ekstra Kuat, Dengan Formula Yang Ampuh menghilangkan noda membandel.</p>",
        "category_id": category_id,
        "images": [lazada_image],
        "stock": 10,
        "price": 16300,
        "sku": "TEST-VIXAL-LAZADA-001",
        "package_weight": 500,
    }
    
    print(f"  Title: {payload['title']}")
    print(f"  Category ID: {payload['category_id']}")
    print(f"  SKU: {payload['sku']}")
    print(f"  Price: Rp {payload['price']:,}")
    
    resp = session.post(f"{BASE_URL}/lazada/products/create", json=payload, headers=headers, timeout=30)
    print(f"\n  HTTP {resp.status_code}")
    print(f"  Response: {resp.text[:800]}")
    
    if resp.status_code in [200, 201]:
        result = resp.json()
        if result.get("success"):
            item_id = result.get("data", {}).get("item_id") or result.get("data", {}).get("itemId")
            print(f"\n  ✅ SUCCESS! Item ID: {item_id}")
        else:
            print(f"\n  ❌ FAILED: {result.get('error') or result.get('message')}")
    else:
        print(f"\n  ❌ HTTP ERROR")

if __name__ == "__main__":
    main()
