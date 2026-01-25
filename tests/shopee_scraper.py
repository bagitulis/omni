import requests
from bs4 import BeautifulSoup
import json
import re
from datetime import datetime

# Cookie dari browser
COOKIES = {
    "SPC_EC": ".MmoxNVFwNW9OUzhGNXFxTybZ+VPuC0H9TWDyfekITY8hwIoXPNT3alj+EgAtZrU++x/dVCDHk8d+jq7a0T48cMF+GJdJswnd7/oKVsbFYpMfjwSBc658kVN+1bOuQhYlCecn9M1kLlpVIEYvZlTsNAoo5hD4898zSZxOiWMpg7P2qKemHjC9e/vcQyxuV+lTK/xE6Nr0+WFRI2opIsqq8qcmrFqrFFgYobWQxZxKATvgrXQalxGbNqDn16Lvyr9s3+8Ma6b6PXMm6yx/1Evofw==",
    "SPC_U": "60492404",
    "SPC_SI": "9FBTaQAAAAA3ZnFBMHRTVwmPUAAAAAAAQ3hjN1IwM00=",
    "csrftoken": "rkWW3OI7S4hD1lcj0mAXYWBfwoqc4RGu",
    "_med": "refer",
    "language": "id",
    "SPC_CLIENTID": "NTVCNnloQURNaGZLbncyxkvhuejnfsvd",
    "SPC_F": "55B6yhADMhfKZbCFiZq2h5gBC2rHsAnU",
    "SPC_R_T_ID": "6Vln/Jb4FdvliwLHr6DM+GLG1wVNA7tH/OetxYtiVPFluKdj0tJRB2BrAHXQ4XZ6xHIdXwigEJKp+qJI2m7Lligjh7PCtUaUcCs5YI9xp9rWdQef1Td+rGoXaKO5cQLPa7Ve/hkKykT2A/BkdDWELmtuO3C1JV2CUJV3eVVbgnw=",
    "SPC_R_T_IV": "ZjlERVFXaW81SEZBWll1Zw==",
    "SPC_ST": ".cXhvOFZWNEU4bEI3TjZQVpqQxi7L+fmFV285GO/eTnTeiIII1eVkA/Gl4oc3r8AeQafNNnOxxpDTA5qVOOZ2P7+y6bBAb9CkRWZxmTiIByONEmc8ygs84fHSLOmwU1RAdXVSD8IFaZWZVGHCdewWAmm5XDBr/X89tCvXPexMXcUCnpgHTAZCAon+RmZbqPXBkbd5kny5TfBNM7kiRNFokM5YwlMMrkb8iRXfRf3jVJTtDGzsaVmzhovoCqecfpoDP5CWKRYbp4AFkaf7kwWfvA==",
    "SPC_T_ID": "6Vln/Jb4FdvliwLHr6DM+GLG1wVNA7tH/OetxYtiVPFluKdj0tJRB2BrAHXQ4XZ6xHIdXwigEJKp+qJI2m7Lligjh7PCtUaUcCs5YI9xp9rWdQef1Td+rGoXaKO5cQLPa7Ve/hkKykT2A/BkdDWELmtuO3C1JV2CUJV3eVVbgnw=",
    "SPC_T_IV": "ZjlERVFXaW81SEZBWll1Zw==",
}

def scrape_shopee_product(url):
    """
    Scrape produk Shopee dan extract: nama, harga, penjualan, rating, stok, dll
    """
    
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Language": "id-ID,id;q=0.9",
        "Referer": "https://shopee.co.id/",
    }
    
    try:
        print(f"[*] Fetching: {url}")
        response = requests.get(url, headers=headers, cookies=COOKIES, timeout=10)
        response.raise_for_status()
        
        print(f"[+] Status: {response.status_code}")
        
        soup = BeautifulSoup(response.text, 'html.parser')
        
        # Extract data dari JSON di script tag (Shopee render dengan JS)
        scripts = soup.find_all('script')
        product_data = None
        
        for script in scripts:
            if script.string and 'window.__INITIAL_STATE__' in script.string:
                try:
                    json_str = script.string.replace('window.__INITIAL_STATE__=', '')
                    product_data = json.loads(json_str)
                    break
                except:
                    pass
        
        if not product_data:
            print("[-] Tidak bisa extract data JSON. Coba parse HTML biasa...")
            return parse_html_fallback(soup)
        
        # Extract dari JSON
        return extract_from_json(product_data)
        
    except Exception as e:
        print(f"[-] Error: {e}")
        return None

def extract_from_json(data):
    """Extract product data dari JSON"""
    result = {
        "timestamp": datetime.now().isoformat(),
        "status": "success",
        "data": {}
    }
    
    try:
        # Shopee menyimpan data di itemData
        if 'item' in data:
            item = data['item']
            
            result['data'] = {
                "nama": item.get('name', 'N/A'),
                "harga": item.get('price', 'N/A'),
                "harga_normal": item.get('price_before_discount', 'N/A'),
                "rating": item.get('rating', {}).get('rating_star', 'N/A'),
                "total_rating": item.get('rating', {}).get('count', 'N/A'),
                "terjual": item.get('sold', 'N/A'),
                "stok": item.get('stock', 'N/A'),
                "penjual": item.get('shop', {}).get('name', 'N/A'),
                "kategori": item.get('category', 'N/A'),
                "deskripsi": item.get('description', 'N/A')[:200] if item.get('description') else 'N/A',
                "sku_list": item.get('variations', []),
                "gambar": item.get('images', []),
                "raw": item
            }
        
    except Exception as e:
        print(f"[-] Error parse JSON: {e}")
        result['status'] = "error"
        result['error'] = str(e)
    
    return result

def parse_html_fallback(soup):
    """Fallback: parse HTML statis"""
    result = {
        "timestamp": datetime.now().isoformat(),
        "status": "fallback",
        "data": {}
    }
    
    try:
        # Coba cari meta tags
        og_title = soup.find('meta', property='og:title')
        og_price = soup.find('meta', property='product:price:amount')
        og_image = soup.find('meta', property='og:image')
        
        result['data'] = {
            "nama": og_title['content'] if og_title else 'N/A',
            "harga": og_price['content'] if og_price else 'N/A',
            "gambar": og_image['content'] if og_image else 'N/A',
            "catatan": "Data terbatas dari meta tags (produk mungkin butuh JS render)"
        }
        
    except Exception as e:
        print(f"[-] Error parse HTML: {e}")
        result['status'] = "error"
        result['error'] = str(e)
    
    return result

def main():
    url = "https://shopee.co.id/Nusseyba-Tahira-Top-All-Size-Bahan-Rayyon-Kerah-Lengkung-Elegan-i.170820234.49403264687?extraParams=%7B%22display_model_id%22%3A445313505035%2C%22model_selection_logic%22%3A3%7D&sp_atk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba&xptdk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba"
    
    print("=" * 60)
    print("SHOPEE PRODUCT SCRAPER")
    print("=" * 60)
    
    result = scrape_shopee_product(url)
    
    if result:
        print("\n[+] HASIL SCRAPE:")
        print("=" * 60)
        
        data = result.get('data', {})
        if data:
            print(f"Status: {result.get('status')}")
            print(f"Timestamp: {result.get('timestamp')}\n")
            
            for key, value in data.items():
                if key not in ['raw', 'sku_list']:
                    if isinstance(value, str) and len(str(value)) > 100:
                        print(f"{key}: {str(value)[:100]}...")
                    else:
                        print(f"{key}: {value}")
        else:
            print(result)
        
        # Save ke file JSON
        output_file = "tests/shopee_product_data.json"
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(result, f, indent=2, ensure_ascii=False)
        print(f"\n[+] Data tersimpan di: {output_file}")
    
    return result

if __name__ == "__main__":
    main()
