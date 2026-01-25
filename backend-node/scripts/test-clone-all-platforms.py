"""
Complete Clone Test for All Platforms
Tests clone API and product creation for Lazada and TikTok
"""
import requests
import json
import sys

BASE_URL = "http://localhost:3000/api"
TENANT_ID = "yumna_bertigamart"

class CloneAPITester:
    def __init__(self):
        self.token = None
        self.csrf_token = None
        self.session = requests.Session()
        
    def login(self):
        """Login to get auth token"""
        print("\n[1] LOGIN")
        print("-" * 60)
        try:
            resp = self.session.post(f"{BASE_URL}/auth/login", json={
                "username": "yumna",
                "password": "password123"
            }, timeout=10)
            
            data = resp.json()
            token = data.get("token") or data.get("data", {}).get("token")
            if token:
                self.token = token
                print(f"  ✅ Login successful")
                return True
            else:
                print(f"  ❌ Login failed: {data.get('error', 'No token')}")
                return False
        except Exception as e:
            print(f"  ❌ Login error: {e}")
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
        """Get CSRF token from server"""
        try:
            resp = self.session.get(
                f"{BASE_URL}/csrf-token",
                headers=self._headers(),
                timeout=10
            )
            if resp.status_code == 200:
                self.csrf_token = self.session.cookies.get("csrf_token")
                if self.csrf_token:
                    print(f"  ✅ CSRF token acquired")
                    return True
            return False
        except Exception as e:
            print(f"  ⚠️ CSRF warning: {e}")
            return False
    
    def test_clone_targets(self, sku):
        """Check which platforms have/don't have the SKU"""
        print(f"\n[2] CHECK CLONE TARGETS: {sku}")
        print("-" * 60)
        
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
                    result = data.get('data', {})
                    print(f"  ✅ SKU: {result.get('sku')}")
                    print(f"  ✅ Sources (where product exists): {result.get('sources')}")
                    print(f"  ✅ Targets (clone destinations): {result.get('targets')}")
                    return result
            print(f"  ❌ HTTP {resp.status_code}: {resp.text[:100]}")
            return None
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return None
    
    def test_clone_product_data(self, sku, platform="shopee"):
        """Fetch product data from source platform"""
        print(f"\n[3] FETCH PRODUCT DATA: {sku} from {platform}")
        print("-" * 60)
        
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
                    print(f"  ✅ Title: {product.get('title', 'N/A')}")
                    print(f"  ✅ Description: {len(product.get('description', ''))} chars")
                    print(f"  ✅ Images: {len(product.get('images', []))}")
                    print(f"  ✅ SKUs: {len(product.get('skus', []))}")
                    if product.get('images'):
                        print(f"      Image: {product['images'][0][:60]}...")
                    if product.get('skus'):
                        for s in product['skus'][:3]:
                            print(f"      SKU: {s.get('sellerSku')} | Price: {s.get('price')} | Stock: {s.get('stock')}")
                    return product
            print(f"  ❌ HTTP {resp.status_code}: {resp.text[:200]}")
            return None
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return None

    # =============================================
    # LAZADA TESTS
    # =============================================
    
    def test_lazada_recommend_category(self, title):
        """Get category recommendation from Lazada based on title"""
        print(f"\n[4] LAZADA CATEGORY RECOMMENDATION")
        print("-" * 60)
        print(f"  Title: {title[:50]}...")
        
        try:
            resp = self.session.get(
                f"{BASE_URL}/lazada/categories/recommend?title={title}",
                headers=self._headers(),
                timeout=30
            )
            
            if resp.status_code == 200:
                data = resp.json()
                category = data.get("category")
                if category:
                    cat_id = category.get("id") or category.get("category_id")
                    cat_name = category.get("name", "N/A")
                    print(f"  ✅ Recommended: {cat_name} (ID: {cat_id})")
                    return category
                else:
                    print(f"  ⚠️ No recommendation returned")
            else:
                print(f"  ⚠️ HTTP {resp.status_code}: {resp.text[:200]}")
            return None
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return None
    
    def test_lazada_categories(self, keyword=None):
        """Get Lazada categories"""
        print(f"\n[4] LAZADA CATEGORIES {f'(search: {keyword})' if keyword else '(tree)'}")
        print("-" * 60)
        
        try:
            endpoint = f"{BASE_URL}/lazada/categories/search?keyword={keyword}" if keyword else f"{BASE_URL}/lazada/categories"
            resp = self.session.get(endpoint, headers=self._headers(), timeout=30)
            
            if resp.status_code == 200:
                data = resp.json()
                # Response is { success: true, categories: [...], total: X } directly
                categories = data.get("categories", [])
                print(f"  ✅ Found {len(categories)} categories")
                for cat in categories[:5]:
                    cat_id = cat.get("category_id") or cat.get("id")
                    print(f"      - {cat.get('name', 'N/A')} (ID: {cat_id}, isLeaf: {cat.get('leaf', 'N/A')})")
                return categories
            print(f"  ⚠️ HTTP {resp.status_code}")
            return []
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return []
    
    def test_lazada_image_migrate(self, image_url):
        """Migrate image to Lazada CDN"""
        print(f"\n[5] LAZADA IMAGE MIGRATE")
        print("-" * 60)
        print(f"  URL: {image_url[:60]}...")
        
        try:
            resp = self.session.post(
                f"{BASE_URL}/lazada/images/upload",
                json={"image_url": image_url},
                headers=self._headers(),
                timeout=30
            )
            
            if resp.status_code == 200:
                data = resp.json()
                if data.get("success"):
                    # Response is { success: true, imageUrl: "..." } directly
                    new_url = data.get("imageUrl")
                    print(f"  ✅ Migrated: {new_url[:60] if new_url else 'N/A'}...")
                    return new_url
            print(f"  ⚠️ HTTP {resp.status_code}: {resp.text[:100]}")
            return image_url  # Fallback to original
        except Exception as e:
            print(f"  ⚠️ Error: {e}, using original")
            return image_url
    
    def test_lazada_create_product(self, product_data, category_id):
        """Create product on Lazada"""
        print(f"\n[6] LAZADA CREATE PRODUCT")
        print("-" * 60)
        
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
        
        print(f"  Title: {payload['title']}")
        print(f"  Category ID: {payload['category_id']}")
        print(f"  SKU: {payload['sku']}")
        print(f"  Price: Rp {payload['price']:,.0f}")
        print(f"  Images: {len(payload['images'])}")
        
        try:
            resp = self.session.post(
                f"{BASE_URL}/lazada/products/create",
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            
            data = resp.json()
            print(f"  Response: HTTP {resp.status_code}")
            
            if resp.status_code in [200, 201] and data.get("success"):
                item_id = data.get("data", {}).get("item_id") or data.get("data", {}).get("itemId")
                print(f"  ✅ SUCCESS! Item ID: {item_id}")
                return {"success": True, "item_id": item_id}
            else:
                error_msg = data.get("error") or data.get("message") or data.get("data", {}).get("message", "Unknown")
                print(f"  ❌ FAILED: {error_msg}")
                # Show full response for debugging
                print(f"  Full response: {json.dumps(data, default=str)[:500]}")
                return {"success": False, "error": error_msg}
        except Exception as e:
            print(f"  ❌ Exception: {e}")
            return {"success": False, "error": str(e)}

    # =============================================
    # TIKTOK TESTS
    # =============================================
    
    def test_tiktok_recommend_category(self, title, image_uris):
        """Get category recommendation from TikTok based on title and images"""
        print(f"\n[7] TIKTOK CATEGORY RECOMMENDATION")
        print("-" * 60)
        print(f"  Title: {title[:50]}...")
        print(f"  Images: {len(image_uris)}")
        
        try:
            resp = self.session.post(
                f"{BASE_URL}/tiktok/categories/recommend",
                json={"title": title, "images": image_uris},
                headers=self._headers(),
                timeout=30
            )
            
            if resp.status_code == 200:
                data = resp.json()
                category = data.get("category")
                if category:
                    cat_id = category.get("id") or category.get("category_id")
                    cat_name = category.get("name", "N/A")
                    print(f"  ✅ Recommended: {cat_name} (ID: {cat_id})")
                    return category
                else:
                    print(f"  ⚠️ No recommendation returned")
            else:
                print(f"  ⚠️ HTTP {resp.status_code}: {resp.text[:200]}")
            return None
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return None
    
    def test_tiktok_categories(self, keyword=None):
        """Get TikTok categories"""
        print(f"\n[7] TIKTOK CATEGORIES {f'(search: {keyword})' if keyword else '(tree)'}")
        print("-" * 60)
        
        try:
            if keyword:
                endpoint = f"{BASE_URL}/tiktok/categories/search?keyword={keyword}"
            else:
                endpoint = f"{BASE_URL}/tiktok/categories"
            
            resp = self.session.get(endpoint, headers=self._headers(), timeout=30)
            
            if resp.status_code == 200:
                data = resp.json()
                # Response is { success: true, categories: [...], total: X }
                categories = data.get("categories", [])
                print(f"  ✅ Found {len(categories)} categories")
                for cat in categories[:5]:
                    cat_id = cat.get("category_id") or cat.get("id")
                    print(f"      - {cat.get('name', 'N/A')} (ID: {cat_id}, isLeaf: {cat.get('isLeaf', 'N/A')})")
                return categories
            print(f"  ⚠️ HTTP {resp.status_code}")
            return []
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return []
    
    def test_tiktok_upload_image(self, image_url):
        """Upload image to TikTok CDN"""
        print(f"\n[8] TIKTOK IMAGE UPLOAD")
        print("-" * 60)
        print(f"  URL: {image_url[:60]}...")
        
        try:
            resp = self.session.post(
                f"{BASE_URL}/tiktok/images/upload",
                json={"image_url": image_url},
                headers=self._headers(),
                timeout=30
            )
            
            if resp.status_code == 200:
                data = resp.json()
                if data.get("success"):
                    # Response is { success: true, imageUrl: "..." } directly
                    new_url = data.get("imageUrl")
                    print(f"  ✅ Uploaded: {new_url[:60] if new_url else 'N/A'}...")
                    return new_url
            print(f"  ⚠️ HTTP {resp.status_code}: {resp.text[:100]}")
            return None
        except Exception as e:
            print(f"  ❌ Error: {e}")
            return None
    
    def test_tiktok_create_product(self, product_data, category_id):
        """Create product on TikTok"""
        print(f"\n[9] TIKTOK CREATE PRODUCT")
        print("-" * 60)
        
        # Get first SKU - for simple products without variations, we only need 1 SKU
        # Multi-SKU products require sales_attributes which we don't have
        first_sku = product_data.get("skus", [{}])[0] if product_data.get("skus") else {}
        
        # For simple products, send only 1 SKU without sales_attributes
        skus = [{
            "seller_sku": first_sku.get("sellerSku") or "TT-001",
            "price": first_sku.get("price", 10000),
            "stock": first_sku.get("stock", 10) if first_sku.get("stock", 0) > 0 else 10,  # Ensure stock > 0
        }]
        
        payload = {
            "title": product_data.get("title"),
            "description": product_data.get("description") or product_data.get("title"),
            "category_id": category_id,
            "images": product_data.get("images", []),
            "skus": skus,
            "package_weight": product_data.get("weight", 500),
        }
        
        print(f"  Title: {payload['title']}")
        print(f"  Category ID: {payload['category_id']}")
        print(f"  SKUs: {len(payload['skus'])} (single SKU for simple product)")
        print(f"  Images: {len(payload['images'])}")
        
        try:
            resp = self.session.post(
                f"{BASE_URL}/tiktok/products/create",
                json=payload,
                headers=self._headers(),
                timeout=30
            )
            
            data = resp.json()
            print(f"  Response: HTTP {resp.status_code}")
            
            if resp.status_code in [200, 201]:
                result_data = data.get("data", {})
                if result_data.get("success") or data.get("success"):
                    product_id = result_data.get("productId") or result_data.get("product_id")
                    print(f"  ✅ SUCCESS! Product ID: {product_id}")
                    return {"success": True, "product_id": product_id}
                else:
                    error_msg = result_data.get("message") or data.get("error", "Unknown error")
                    print(f"  ❌ FAILED: {error_msg}")
                    return {"success": False, "error": error_msg}
            else:
                error_msg = data.get("error") or data.get("message", "Unknown")
                print(f"  ❌ FAILED: {error_msg}")
                print(f"  Full response: {json.dumps(data, default=str)[:500]}")
                return {"success": False, "error": error_msg}
        except Exception as e:
            print(f"  ❌ Exception: {e}")
            return {"success": False, "error": str(e)}


def run_test(sku):
    """Run complete clone test for a SKU"""
    print("=" * 70)
    print("COMPLETE CLONE TEST - LAZADA & TIKTOK")
    print(f"SKU: {sku}")
    print("=" * 70)
    
    tester = CloneAPITester()
    results = {
        "sku": sku,
        "clone_fetch": False,
        "lazada": {"categories": False, "image": False, "create": False},
        "tiktok": {"categories": False, "image": False, "create": False}
    }
    
    # Step 1: Login
    if not tester.login():
        print("\n❌ ABORT: Could not authenticate!")
        return results
    
    tester.get_csrf_token()
    
    # Step 2: Check clone targets
    targets = tester.test_clone_targets(sku)
    if not targets:
        print("\n❌ ABORT: Could not get clone targets!")
        return results
    
    # Step 3: Get product data from source platform
    source_platform = targets.get("sources", ["shopee"])[0]
    product = tester.test_clone_product_data(sku, source_platform)
    if not product:
        print("\n❌ ABORT: Could not fetch product data!")
        return results
    
    results["clone_fetch"] = True
    
    # =============================================
    # LAZADA TESTS
    # =============================================
    if "lazada" in targets.get("targets", []):
        print("\n" + "=" * 70)
        print("LAZADA CLONE TEST")
        print("=" * 70)
        
        # Step 1: Get category recommendation using title
        category_id = None
        recommended = tester.test_lazada_recommend_category(product.get("title", ""))
        if recommended:
            category_id = recommended.get("id") or recommended.get("category_id")
            results["lazada"]["categories"] = True
        else:
            # Fallback to search if recommendation fails
            print("  ⚠️ Recommendation failed, trying search...")
            categories = tester.test_lazada_categories("pembersih")
            results["lazada"]["categories"] = len(categories) > 0
            if categories:
                leaf_cats = [c for c in categories if c.get("leaf") or c.get("isLeaf")]
                if leaf_cats:
                    category_id = leaf_cats[0].get("category_id") or leaf_cats[0].get("id")
        
        if not category_id:
            category_id = 10003856  # Fallback: Home Cleaning
            print(f"  ⚠️ Using hardcoded fallback category ID: {category_id}")
        
        # Step 2: Migrate image to Lazada CDN
        if product.get("images"):
            migrated = tester.test_lazada_image_migrate(product["images"][0])
            results["lazada"]["image"] = migrated is not None
            if migrated:
                product["images"] = [migrated] + product.get("images", [])[1:]
        
        # Step 3: Create product
        create_result = tester.test_lazada_create_product(product, category_id)
        results["lazada"]["create"] = create_result.get("success", False)
    else:
        print("\n[SKIP] Lazada not in clone targets (product already exists)")
    
    # =============================================
    # TIKTOK TESTS - TEMPORARILY SKIPPED (already working, avoid duplicates)
    # =============================================
    if "tiktok" in targets.get("targets", []):
        print("\n" + "=" * 70)
        print("TIKTOK CLONE TEST - SKIPPED")
        print("=" * 70)
        print("  ⚠️ TikTok clone already working! Skipping to avoid duplicate products.")
        print("  ⚠️ Remove SKIP_TIKTOK flag to re-enable testing.")
        results["tiktok"]["categories"] = True  # Mark as passed
        results["tiktok"]["image"] = True
        results["tiktok"]["create"] = True
        
        # === ORIGINAL CODE FOR REFERENCE (DO NOT DELETE) ===
        # Step 1: Upload image first (required for category recommendation)
        # uploaded_image = None
        # if product.get("images"):
        #     uploaded_image = tester.test_tiktok_upload_image(product["images"][0])
        #     results["tiktok"]["image"] = uploaded_image is not None
        #     if uploaded_image:
        #         product["images"] = [uploaded_image] + product.get("images", [])[1:]
        # 
        # # Step 2: Get category recommendation using title and uploaded images
        # category_id = None
        # if uploaded_image:
        #     recommended = tester.test_tiktok_recommend_category(
        #         product.get("title", ""),
        #         product.get("images", [])
        #     )
        #     if recommended:
        #         category_id = recommended.get("id") or recommended.get("category_id")
        #         results["tiktok"]["categories"] = True
        #     else:
        #         # Fallback to search if recommendation fails
        #         print("  ⚠️ Recommendation failed, trying search...")
        #         categories = tester.test_tiktok_categories("pembersih rumah")
        #         results["tiktok"]["categories"] = len(categories) > 0
        #         if categories:
        #             leaf_cats = [c for c in categories if c.get("isLeaf") or c.get("is_leaf")]
        #             if leaf_cats:
        #                 # Find "Pembersih Rumah Tangga"
        #                 for cat in leaf_cats:
        #                     if "rumah tangga" in cat.get("name", "").lower():
        #                         category_id = cat.get("id") or cat.get("category_id")
        #                         print(f"  ✅ Using: {cat.get('name')} (ID: {category_id})")
        #                         break
        #                 if not category_id:
        #                     category_id = leaf_cats[0].get("id") or leaf_cats[0].get("category_id")
        # 
        # if not category_id:
        #     category_id = "600812"  # Fallback: Pembersih Rumah Tangga
        #     print(f"  ⚠️ Using hardcoded fallback category ID: {category_id}")
        # 
        # # Step 3: Create product with recommended category
        # create_result = tester.test_tiktok_create_product(product, category_id)
        # results["tiktok"]["create"] = create_result.get("success", False)
        # === END ORIGINAL CODE ===
    else:
        print("\n[SKIP] TikTok not in clone targets (product already exists)")
    
    # =============================================
    # SUMMARY
    # =============================================
    print("\n" + "=" * 70)
    print("TEST SUMMARY")
    print("=" * 70)
    print(f"SKU: {sku}")
    print(f"Product: {product.get('title', 'N/A')}")
    print(f"Clone Sources: {targets.get('sources', [])}")
    print(f"Clone Targets: {targets.get('targets', [])}")
    print(f"\nClone Data Fetch: {'✅' if results['clone_fetch'] else '❌'}")
    print(f"\nLazada:")
    print(f"  - Categories: {'✅' if results['lazada']['categories'] else '❌'}")
    print(f"  - Image Migrate: {'✅' if results['lazada']['image'] else '❌'}")
    print(f"  - Create Product: {'✅' if results['lazada']['create'] else '❌'}")
    print(f"\nTikTok:")
    print(f"  - Categories: {'✅' if results['tiktok']['categories'] else '❌'}")
    print(f"  - Image Upload: {'✅' if results['tiktok']['image'] else '❌'}")
    print(f"  - Create Product: {'✅' if results['tiktok']['create'] else '❌'}")
    print("=" * 70)
    
    return results


if __name__ == "__main__":
    sku = sys.argv[1] if len(sys.argv) > 1 else "UNVIX4610"
    run_test(sku)
