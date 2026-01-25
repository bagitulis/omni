"""
Real API Test - Clone and Add Product
Tests actual backend API endpoints

// turbo-all
"""
import requests
import json
import sys

BASE_URL = "http://localhost:3000/api"
TENANT_ID = "yumna_bertigamart"

class APITester:
    def __init__(self):
        self.token = None
        self.csrf_token = None
        self.session = requests.Session()
        
    def login(self):
        """Login to get auth token"""
        print("\n[1] LOGIN TO GET AUTH TOKEN")
        print("-" * 50)
        
        try:
            resp = self.session.post(f"{BASE_URL}/auth/login", json={
                "username": "yumna",
                "password": "password123"
            }, timeout=10)
            
            data = resp.json()
            # Handle various response formats
            token = data.get("token") or data.get("data", {}).get("token")
            if token:
                self.token = token
                print(f"  [OK] Login successful")
                print(f"  [OK] Token: {self.token[:40]}...")
                return True
            elif data.get("message") == "Login successful":
                # Token might be in user object or headers
                self.token = resp.headers.get("authorization", "").replace("Bearer ", "")
                if not self.token and "user" in data:
                    # Try to get token from response
                    self.token = data.get("accessToken") or data.get("access_token", "")
                if self.token:
                    print(f"  [OK] Login successful")
                    return True
                else:
                    print(f"  [WARN] Login OK but no token in response")
                    print(f"  [DEBUG] Response: {json.dumps(data, default=str)[:300]}")
                    return False
            else:
                print(f"  [FAIL] Login failed: {data.get('error', data.get('message', 'Unknown'))}")
                return False
        except Exception as e:
            print(f"  [FAIL] Login error: {e}")
            return False
    
    def _headers(self):
        headers = {
            "Authorization": f"Bearer {self.token}",
            "x-tenant-id": TENANT_ID,
            "Content-Type": "application/json"
        }
        if self.csrf_token:
            headers["x-csrf-token"] = self.csrf_token
        return headers
    
    def get_csrf_token(self):
        """Get CSRF token from server - token is set in cookie"""
        try:
            resp = self.session.get(
                f"{BASE_URL}/csrf-token",
                headers=self._headers(),
                timeout=10
            )
            if resp.status_code == 200:
                # Token is set in cookie by server, session.cookies automatically stores it
                # We need to read from session cookies and also send as header
                self.csrf_token = self.session.cookies.get("csrf_token")
                if self.csrf_token:
                    print(f"  [OK] CSRF token acquired from cookie")
                    return True
                else:
                    print(f"  [WARN] CSRF token not in cookies")
            return False
        except Exception as e:
            print(f"  [FAIL] CSRF error: {e}")
            return False
    
    def test_health(self):
        """Test health endpoint"""
        print("\n[2] TEST HEALTH ENDPOINT")
        print("-" * 50)
        try:
            resp = self.session.get(f"{BASE_URL}/health", timeout=5)
            data = resp.json()
            print(f"  [OK] Status: {data.get('status', 'unknown')}")
            return True
        except Exception as e:
            print(f"  [FAIL] Health check failed: {e}")
            return False
    
    def test_clone_available_targets(self, sku):
        """Test clone available targets endpoint"""
        print(f"\n[3] TEST CLONE AVAILABLE TARGETS: {sku}")
        print("-" * 50)
        
        try:
            resp = self.session.get(
                f"{BASE_URL}/clone/available-targets",
                params={"sku": sku},
                headers=self._headers(),
                timeout=10
            )
            
            if resp.status_code == 200:
                data = resp.json()
                if data.get("success"):
                    print(f"  [OK] SKU: {data.get('data', {}).get('sku')}")
                    print(f"  [OK] Sources: {data.get('data', {}).get('sources')}")
                    print(f"  [OK] Targets: {data.get('data', {}).get('targets')}")
                    return data.get('data')
                else:
                    print(f"  [FAIL] API returned error: {data.get('error')}")
            else:
                print(f"  [FAIL] HTTP {resp.status_code}: {resp.text[:100]}")
            return None
        except Exception as e:
            print(f"  [FAIL] Error: {e}")
            return None
    
    def test_clone_product_data(self, sku, platform="shopee"):
        """Test clone product data endpoint"""
        print(f"\n[4] TEST CLONE PRODUCT DATA: {sku} from {platform}")
        print("-" * 50)
        
        try:
            resp = self.session.get(
                f"{BASE_URL}/clone/product-data",
                params={"sku": sku, "platform": platform},
                headers=self._headers(),
                timeout=10
            )
            
            if resp.status_code == 200:
                data = resp.json()
                if data.get("success"):
                    product = data.get("data", {}).get("product", {})
                    print(f"  [OK] Title: {product.get('title', 'N/A')}")
                    print(f"  [OK] Description: {len(product.get('description', ''))} chars")
                    print(f"  [OK] Images: {len(product.get('images', []))}")
                    print(f"  [OK] SKUs: {len(product.get('skus', []))}")
                    return product
                else:
                    print(f"  [FAIL] API returned error: {data.get('error')}")
            else:
                print(f"  [FAIL] HTTP {resp.status_code}: {resp.text[:200]}")
            return None
        except Exception as e:
            print(f"  [FAIL] Error: {e}")
            return None
    
    def test_lazada_categories_search(self, keyword):
        """Test Lazada category search"""
        print(f"\n[5] TEST LAZADA CATEGORY SEARCH: {keyword}")
        print("-" * 50)
        
        try:
            resp = self.session.get(
                f"{BASE_URL}/lazada/categories/search",
                params={"keyword": keyword},
                headers=self._headers(),
                timeout=15
            )
            
            if resp.status_code == 200:
                data = resp.json()
                if data.get("success"):
                    categories = data.get("data", {}).get("categories", [])
                    print(f"  [OK] Found {len(categories)} categories")
                    if categories:
                        for cat in categories[:3]:
                            print(f"      - {cat.get('name', 'N/A')} (ID: {cat.get('category_id', cat.get('id', 'N/A'))})")
                        return categories[0]  # Return first match
                else:
                    print(f"  [WARN] {data.get('error', 'No categories found')}")
            else:
                print(f"  [WARN] HTTP {resp.status_code}")
            return None
        except Exception as e:
            print(f"  [FAIL] Error: {e}")
            return None
    
    def test_lazada_create_product(self, product_data, category_id):
        """Test Lazada create product"""
        print(f"\n[6] TEST LAZADA CREATE PRODUCT")
        print("-" * 50)
        
        # Get first SKU data
        first_sku = product_data.get("skus", [{}])[0] if product_data.get("skus") else {}
        
        payload = {
            "title": product_data.get("title"),
            "name": product_data.get("title"),
            "description": product_data.get("description") or f"<p>{product_data.get('title')}</p>",
            "category_id": category_id,
            "images": product_data.get("images", []),
            "stock": first_sku.get("stock", 10),
            "price": first_sku.get("price", 10000),
            "sku": first_sku.get("sellerSku", "TEST-SKU"),
            "package_weight": product_data.get("weight", 500),
        }
        
        print(f"  Payload:")
        print(f"    - Title: {payload['title']}")
        print(f"    - Category ID: {payload['category_id']}")
        print(f"    - SKU: {payload['sku']}")
        print(f"    - Price: Rp {payload['price']:,.0f}")
        print(f"    - Stock: {payload['stock']}")
        print(f"    - Images: {len(payload['images'])}")
        
        try:
            resp = self.session.post(
                f"{BASE_URL}/lazada/products/create",
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            
            data = resp.json()
            print(f"  Response status: HTTP {resp.status_code}")
            
            if resp.status_code in [200, 201] and data.get("success"):
                print(f"  [OK] Product created on Lazada!")
                if data.get("data", {}).get("item_id"):
                    print(f"  [OK] Item ID: {data['data']['item_id']}")
                return {"success": True, "data": data}
            else:
                print(f"  [FAIL] Create failed")
                error_msg = data.get("error") or data.get("message") or data.get("data", {}).get("message", "Unknown")
                print(f"  [FAIL] Error: {error_msg}")
                return {"success": False, "error": error_msg, "data": data}
        except Exception as e:
            print(f"  [FAIL] Exception: {e}")
            return {"success": False, "error": str(e)}

def run_test(sku):
    print("=" * 60)
    print("REAL API CLONE TEST")
    print(f"SKU: {sku}")
    print("=" * 60)
    
    tester = APITester()
    
    # Step 1: Health check
    if not tester.test_health():
        print("\n[ABORT] Backend not running!")
        return False
    
    # Step 2: Login
    if not tester.login():
        print("\n[ABORT] Could not authenticate!")
        return False
    
    # Step 2.5: Get CSRF token for POST requests
    tester.get_csrf_token()
    
    # Step 3: Check available targets
    targets = tester.test_clone_available_targets(sku)
    if not targets:
        print("\n[ABORT] Could not get clone targets!")
        return False
    
    # Step 4: Get product data
    product = tester.test_clone_product_data(sku, "shopee")
    if not product:
        print("\n[ABORT] Could not fetch product data!")
        return False
    
    # Step 5: Search for Lazada category 
    category = tester.test_lazada_categories_search("pembersih")
    
    # Step 6: Create product on Lazada (if category found and target available)
    if "lazada" in [t.lower() for t in targets.get("targets", [])]:
        if category:
            category_id = category.get("category_id") or category.get("id")
            result = tester.test_lazada_create_product(product, category_id)
        else:
            # Use fallback category ID for testing
            print("\n[INFO] No category found, using fallback category ID")
            result = tester.test_lazada_create_product(product, 10003856)  # Home Cleaning category
    else:
        print("\n[INFO] Lazada not in targets - skipping")
        result = None
    
    # Summary
    print("\n" + "=" * 60)
    print("TEST RESULT SUMMARY")
    print("=" * 60)
    print(f"SKU: {sku}")
    print(f"Product: {product.get('title', 'N/A')}")
    print(f"Clone Sources: {targets.get('sources', [])}")
    print(f"Clone Targets: {targets.get('targets', [])}")
    print(f"Images: {len(product.get('images', []))}")
    print(f"Category Found: {'Yes' if category else 'No (used fallback)'}")
    if result:
        print(f"Lazada Create: {'SUCCESS' if result.get('success') else 'FAILED - ' + str(result.get('error', 'Unknown'))}")
    print("=" * 60)
    
    return True

if __name__ == "__main__":
    sku = sys.argv[1] if len(sys.argv) > 1 else "UNVIX4610"
    run_test(sku)
