import asyncio
import json
from datetime import datetime
from bs4 import BeautifulSoup
import sys

try:
    from playwright.async_api import async_playwright
except ImportError:
    print("[-] Playwright tidak terinstall. Install dengan: pip install playwright")
    sys.exit(1)

COOKIES = [
    {"name": "SPC_EC", "value": ".MmoxNVFwNW9OUzhGNXFxTybZ+VPuC0H9TWDyfekITY8hwIoXPNT3alj+EgAtZrU++x/dVCDHk8d+jq7a0T48cMF+GJdJswnd7/oKVsbFYpMfjwSBc658kVN+1bOuQhYlCecn9M1kLlpVIEYvZlTsNAoo5hD4898zSZxOiWMpg7P2qKemHjC9e/vcQyxuV+lTK/xE6Nr0+WFRI2opIsqq8qcmrFqrFFgYobWQxZxKATvgrXQalxGbNqDn16Lvyr9s3+8Ma6b6PXMm6yx/1Evofw==", "domain": ".shopee.co.id", "path": "/"},
    {"name": "SPC_U", "value": "60492404", "domain": ".shopee.co.id", "path": "/"},
    {"name": "SPC_SI", "value": "9FBTaQAAAAA3ZnFBMHRTVwmPUAAAAAAAQ3hjN1IwM00=", "domain": ".shopee.co.id", "path": "/"},
    {"name": "csrftoken", "value": "rkWW3OI7S4hD1lcj0mAXYWBfwoqc4RGu", "domain": ".shopee.co.id", "path": "/"},
    {"name": "language", "value": "id", "domain": ".shopee.co.id", "path": "/"},
    {"name": "SPC_F", "value": "55B6yhADMhfKZbCFiZq2h5gBC2rHsAnU", "domain": ".shopee.co.id", "path": "/"},
]

async def scrape_with_browser(url):
    """Scrape dengan headless browser untuk JS rendering"""
    
    async with async_playwright() as p:
        print("[*] Launching browser...")
        browser = await p.chromium.launch(headless=True, args=["--no-sandbox"])
        context = await browser.new_context()
        
        # Set cookies
        print("[*] Setting cookies...")
        await context.add_cookies(COOKIES)
        
        page = await context.new_page()
        
        try:
            print(f"[*] Loading page: {url[:60]}...")
            await page.goto(url, wait_until="networkidle", timeout=30000)
            
            print("[*] Waiting for content to load...")
            await page.wait_for_timeout(3000)  # Wait 3 detik untuk JS render
            
            # Get page content
            content = await page.content()
            
            # Parse dengan BeautifulSoup
            soup = BeautifulSoup(content, 'html.parser')
            
            result = extract_product_data(soup)
            
            return result
            
        except Exception as e:
            print(f"[-] Error: {e}")
            return {"status": "error", "error": str(e)}
        
        finally:
            await browser.close()

def extract_product_data(soup):
    """Extract data dari halaman yang sudah di-render"""
    
    data = {
        "timestamp": datetime.now().isoformat(),
        "status": "success",
        "product": {}
    }
    
    try:
        # Cari nama produk
        nama_elem = soup.find('h1', class_=lambda x: x and 'product' in x.lower())
        if not nama_elem:
            nama_elem = soup.find('span', {'data-test': 'product-title'})
        
        nama = nama_elem.get_text(strip=True) if nama_elem else 'N/A'
        
        # Cari harga
        price_elem = soup.find('span', class_=lambda x: x and 'price' in x.lower())
        if not price_elem:
            price_elem = soup.find('span', {'data-test': 'product-price'})
        
        harga = price_elem.get_text(strip=True) if price_elem else 'N/A'
        
        # Cari rating
        rating_elem = soup.find('span', class_=lambda x: x and 'rating' in x.lower())
        rating = rating_elem.get_text(strip=True) if rating_elem else 'N/A'
        
        # Cari sold/terjual
        sold_elem = soup.find(text=lambda x: x and 'terjual' in x.lower())
        terjual = sold_elem if sold_elem else 'N/A'
        
        # Cari stok
        stock_elem = soup.find(text=lambda x: x and 'stok' in x.lower())
        stok = stock_elem if stock_elem else 'N/A'
        
        # Cari penjual
        shop_elem = soup.find('a', {'data-test': 'seller-name'})
        if not shop_elem:
            shop_elem = soup.find('a', class_=lambda x: x and 'shop' in x.lower())
        
        penjual = shop_elem.get_text(strip=True) if shop_elem else 'N/A'
        
        # Cari deskripsi (dari section pertama)
        desc_elem = soup.find('div', class_=lambda x: x and 'description' in x.lower())
        if not desc_elem:
            desc_elem = soup.find('div', {'data-test': 'product-description'})
        
        deskripsi = desc_elem.get_text(strip=True)[:200] if desc_elem else 'N/A'
        
        # Cari gambar
        img_elems = soup.find_all('img', class_=lambda x: x and ('product' in x.lower() or 'image' in x.lower()))
        gambar = [img.get('src', '') for img in img_elems[:5]]
        
        data['product'] = {
            "nama": nama,
            "harga": harga,
            "rating": rating,
            "terjual": str(terjual),
            "stok": str(stok),
            "penjual": penjual,
            "deskripsi": deskripsi,
            "gambar": gambar,
        }
        
        print("\n[+] DATA BERHASIL DIEKSTRAK:")
        print(f"Nama: {nama}")
        print(f"Harga: {harga}")
        print(f"Rating: {rating}")
        print(f"Terjual: {terjual}")
        print(f"Stok: {stok}")
        print(f"Penjual: {penjual}")
        print(f"Gambar: {len(gambar)} ditemukan")
        
    except Exception as e:
        print(f"[-] Error extract: {e}")
        data['status'] = "error"
        data['error'] = str(e)
    
    return data

async def main():
    url = "https://shopee.co.id/Nusseyba-Tahira-Top-All-Size-Bahan-Rayyon-Kerah-Lengkung-Elegan-i.170820234.49403264687?extraParams=%7B%22display_model_id%22%3A445313505035%2C%22model_selection_logic%22%3A3%7D&sp_atk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba&xptdk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba"
    
    print("=" * 70)
    print("SHOPEE PRODUCT SCRAPER (dengan Playwright)")
    print("=" * 70)
    
    result = await scrape_with_browser(url)
    
    # Save hasil
    output_file = "tests/shopee_product_data_playwright.json"
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(result, f, indent=2, ensure_ascii=False)
    
    print(f"\n[+] Data tersimpan di: {output_file}")
    print("\n" + "=" * 70)

if __name__ == "__main__":
    asyncio.run(main())
