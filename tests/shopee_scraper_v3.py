import asyncio
import json
import re
from datetime import datetime
from bs4 import BeautifulSoup

try:
    from playwright.async_api import async_playwright
except ImportError:
    print("[-] Playwright tidak terinstall")
    exit(1)

COOKIES = [
    {"name": "SPC_EC", "value": ".MmoxNVFwNW9OUzhGNXFxTybZ+VPuC0H9TWDyfekITY8hwIoXPNT3alj+EgAtZrU++x/dVCDHk8d+jq7a0T48cMF+GJdJswnd7/oKVsbFYpMfjwSBc658kVN+1bOuQhYlCecn9M1kLlpVIEYvZlTsNAoo5hD4898zSZxOiWMpg7P2qKemHjC9e/vcQyxuV+lTK/xE6Nr0+WFRI2opIsqq8qcmrFqrFFgYobWQxZxKATvgrXQalxGbNqDn16Lvyr9s3+8Ma6b6PXMm6yx/1Evofw==", "domain": ".shopee.co.id", "path": "/"},
    {"name": "SPC_U", "value": "60492404", "domain": ".shopee.co.id", "path": "/"},
    {"name": "language", "value": "id", "domain": ".shopee.co.id", "path": "/"},
]

async def scrape_with_javascript_eval(url):
    """Scrape dengan evaluate JavaScript langsung di page"""
    
    async with async_playwright() as p:
        print("[*] Launching Chromium...")
        browser = await p.chromium.launch(headless=True)
        
        context = await browser.new_context()
        await context.add_cookies(COOKIES)
        
        page = await context.new_page()
        
        try:
            print(f"[*] Loading page...")
            await page.goto(url, wait_until="load", timeout=30000)
            
            print("[*] Waiting for JavaScript to render...")
            await page.wait_for_timeout(5000)
            
            # Evaluate JavaScript untuk extract data
            print("[*] Extracting data with JavaScript...")
            
            # Pattern 1: Cari window.__INITIAL_STATE__
            data = await page.evaluate("""() => {
                try {
                    if (window.__INITIAL_STATE__) {
                        return window.__INITIAL_STATE__;
                    }
                } catch (e) {
                    console.log('No __INITIAL_STATE__');
                }
                return null;
            }""")
            
            if data:
                print("[+] Found __INITIAL_STATE__!")
                result = parse_shopee_data(data)
                if result:
                    await browser.close()
                    return result
            
            # Pattern 2: Cari di page source
            print("[*] Extracting from page HTML...")
            html = await page.content()
            soup = BeautifulSoup(html, 'html.parser')
            
            # Cari script yang contain JSON
            scripts = soup.find_all('script')
            for script in scripts:
                if script.string:
                    script_content = script.string
                    
                    # Cari price di format tertentu
                    price_match = re.search(r'"price"\s*:\s*(\d+)', script_content)
                    name_match = re.search(r'"name"\s*:\s*"([^"]+)"', script_content)
                    
                    if price_match or name_match:
                        print("[+] Found data in script tag!")
                        
                        # Extract JSON object
                        try:
                            json_str = re.search(r'({[^{}]*"price"[^{}]*})', script_content)
                            if json_str:
                                item_data = json.loads(json_str.group(1))
                                return {
                                    "timestamp": datetime.now().isoformat(),
                                    "status": "success",
                                    "source": "JavaScript evaluation",
                                    "product": {
                                        "nama": item_data.get('name', 'N/A'),
                                        "harga": item_data.get('price', 'N/A'),
                                        "rating": item_data.get('rating', 'N/A'),
                                        "terjual": item_data.get('sold', 'N/A'),
                                    }
                                }
                        except Exception as e:
                            print(f"[-] JSON parse error: {e}")
            
            # Pattern 3: Fallback ke DOM querySelector
            print("[*] Using DOM query selectors...")
            
            dom_data = await page.evaluate("""() => {
                return {
                    name: document.querySelector('h1')?.textContent || null,
                    price: document.querySelector('[data-price]')?.textContent ||
                            document.querySelector('.product-price')?.textContent || null,
                    rating: document.querySelector('[data-rating]')?.textContent ||
                             document.querySelector('.rating')?.textContent || null,
                    sold: document.querySelector('[data-sold]')?.textContent ||
                          document.querySelector('.sold')?.textContent || null,
                };
            }""")
            
            if dom_data and any(v for v in dom_data.values()):
                print("[+] Found data via DOM!")
                return {
                    "timestamp": datetime.now().isoformat(),
                    "status": "success",
                    "source": "DOM selectors",
                    "product": {
                        "nama": dom_data.get('name', 'N/A'),
                        "harga": dom_data.get('price', 'N/A'),
                        "rating": dom_data.get('rating', 'N/A'),
                        "terjual": dom_data.get('sold', 'N/A'),
                    }
                }
            
            # Fallback: return HTML untuk manual inspection
            print("[*] Saving HTML untuk inspection...")
            with open('tests/shopee_page_snapshot.html', 'w', encoding='utf-8') as f:
                f.write(html)
            
            return {
                "timestamp": datetime.now().isoformat(),
                "status": "partial",
                "source": "HTML snapshot saved",
                "message": "No structured data found. Check shopee_page_snapshot.html",
                "product": {}
            }
            
        except Exception as e:
            print(f"[-] Error: {e}")
            return {
                "timestamp": datetime.now().isoformat(),
                "status": "error",
                "error": str(e)
            }
        
        finally:
            await browser.close()

def parse_shopee_data(data):
    """Parse data dari __INITIAL_STATE__"""
    
    try:
        # Struktur yang mungkin
        if isinstance(data, dict):
            # Cari item/product data
            for key in ['item', 'product', 'itemData']:
                if key in data:
                    item = data[key]
                    return {
                        "timestamp": datetime.now().isoformat(),
                        "status": "success",
                        "source": "__INITIAL_STATE__",
                        "product": {
                            "nama": item.get('name', 'N/A'),
                            "harga": item.get('price', 'N/A'),
                            "rating": item.get('rating_star', item.get('rating', 'N/A')),
                            "terjual": item.get('sold', 'N/A'),
                            "stok": item.get('stock', 'N/A'),
                            "raw_data": item
                        }
                    }
        
        return None
    
    except Exception as e:
        print(f"[-] Parse error: {e}")
        return None

async def main():
    url = "https://shopee.co.id/Nusseyba-Tahira-Top-All-Size-Bahan-Rayyon-Kerah-Lengkung-Elegan-i.170820234.49403264687?extraParams=%7B%22display_model_id%22%3A445313505035%2C%22model_selection_logic%22%3A3%7D&sp_atk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba&xptdk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba"
    
    print("=" * 70)
    print("SHOPEE SCRAPER v3 - dengan JavaScript Evaluation")
    print("=" * 70 + "\n")
    
    result = await scrape_with_javascript_eval(url)
    
    print("\n" + "=" * 70)
    print("HASIL:")
    print("=" * 70)
    print(json.dumps(result, indent=2, ensure_ascii=False))
    
    # Save
    with open('tests/shopee_product_v3.json', 'w', encoding='utf-8') as f:
        json.dump(result, f, indent=2, ensure_ascii=False)
    
    print("\n[+] Data tersimpan di: tests/shopee_product_v3.json")
    
    if result['status'] == 'partial':
        print("[!] HTML page snapshot tersimpan di: tests/shopee_page_snapshot.html")
        print("[!] Buka file tersebut untuk inspect struktur halaman")

if __name__ == "__main__":
    asyncio.run(main())
