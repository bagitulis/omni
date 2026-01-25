"""
Test Lazada Category and Image Migration APIs
"""
import requests
import json

BASE_URL = "http://localhost:3000/api"
TENANT_ID = "yumna_bertigamart"

def get_token():
    resp = requests.post(f"{BASE_URL}/auth/login", json={
        "username": "yumna",
        "password": "password123"
    })
    data = resp.json()
    return data.get("token") or data.get("data", {}).get("token")

def headers(token, csrf=None):
    h = {
        "Authorization": f"Bearer {token}",
        "x-tenant-id": TENANT_ID,
        "Content-Type": "application/json"
    }
    if csrf:
        h["x-csrf-token"] = csrf
    return h

def test_category_tree(token):
    """Test getting Lazada category tree"""
    print("\n[1] GET LAZADA CATEGORY TREE")
    print("-" * 50)
    
    resp = requests.get(
        f"{BASE_URL}/lazada/categories",
        headers=headers(token),
        timeout=30
    )
    
    if resp.status_code == 200:
        data = resp.json()
        categories = data.get("data", {}).get("categories", [])
        print(f"  [OK] Found {len(categories)} root categories")
        
        # Find cleaning related categories
        cleaning_cats = []
        for cat in categories[:50]:  # Check first 50
            name = cat.get("name", "").lower()
            if any(x in name for x in ["clean", "household", "home", "bersih"]):
                cleaning_cats.append(cat)
        
        if cleaning_cats:
            print(f"  [OK] Found {len(cleaning_cats)} cleaning-related categories:")
            for cat in cleaning_cats[:5]:
                print(f"      - {cat.get('name')} (ID: {cat.get('category_id', cat.get('id'))})")
            return cleaning_cats[0].get("category_id") or cleaning_cats[0].get("id")
        
        # Fallback: show first 5 categories
        print(f"  [INFO] First 5 categories:")
        for cat in categories[:5]:
            print(f"      - {cat.get('name')} (ID: {cat.get('category_id', cat.get('id'))})")
        
        return categories[0].get("category_id") if categories else None
    else:
        print(f"  [FAIL] HTTP {resp.status_code}: {resp.text[:200]}")
        return None

def test_image_migrate(token, image_url):
    """Test image migration to Lazada CDN"""
    print(f"\n[2] TEST IMAGE MIGRATE")
    print("-" * 50)
    print(f"  URL: {image_url[:60]}...")
    
    resp = requests.post(
        f"{BASE_URL}/lazada/images/upload",
        json={"image_url": image_url},
        headers=headers(token),
        timeout=30
    )
    
    if resp.status_code == 200:
        data = resp.json()
        if data.get("success"):
            new_url = data.get("data", {}).get("imageUrl")
            print(f"  [OK] Migrated: {new_url[:60] if new_url else 'N/A'}...")
            return new_url
        else:
            print(f"  [FAIL] {data.get('error', 'Unknown')}")
    else:
        print(f"  [FAIL] HTTP {resp.status_code}: {resp.text[:200]}")
    
    return None

def main():
    print("=" * 60)
    print("LAZADA API DEBUG TEST")
    print("=" * 60)
    
    token = get_token()
    if not token:
        print("[ERROR] Failed to login")
        return
    
    print(f"  [OK] Token acquired")
    
    # Test 1: Get categories
    category_id = test_category_tree(token)
    print(f"\n  Suggested Category ID: {category_id}")
    
    # Test 2: Image migration
    shopee_image = "https://cf.shopee.co.id/file/id-11134207-7qul4-lk2a2fniq09k3"
    migrated = test_image_migrate(token, shopee_image)
    
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"Category ID: {category_id}")
    print(f"Image migrated: {'Yes' if migrated else 'No (use original)'}")

if __name__ == "__main__":
    main()
