#!/usr/bin/env python3
"""
Shopee Product Scraper - Simple Version
Menggunakan requests + parsing HTML dengan fallback ke regex
"""

import requests
import json
import re
from datetime import datetime
from bs4 import BeautifulSoup

# User Agent & Cookies
HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
    "Accept-Language": "id-ID,id;q=0.9,en;q=0.8",
    "Accept-Encoding": "gzip, deflate, br",
    "DNT": "1",
    "Connection": "keep-alive",
    "Upgrade-Insecure-Requests": "1",
    "Referer": "https://shopee.co.id/",
}

COOKIES = {
    "SPC_EC": ".MmoxNVFwNW9OUzhGNXFxTybZ+VPuC0H9TWDyfekITY8hwIoXPNT3alj+EgAtZrU++x/dVCDHk8d+jq7a0T48cMF+GJdJswnd7/oKVsbFYpMfjwSBc658kVN+1bOuQhYlCecn9M1kLlpVIEYvZlTsNAoo5hD4898zSZxOiWMpg7P2qKemHjC9e/vcQyxuV+lTK/xE6Nr0+WFRI2opIsqq8qcmrFqrFFgYobWQxZxKATvgrXQalxGbNqDn16Lvyr9s3+8Ma6b6PXMm6yx/1Evofw==",
    "SPC_U": "60492404",
    "language": "id",
}

def extract_from_json_in_html(html_content):
    """
    Extract data dari JSON yang embedded di HTML
    Shopee menyimpan data di <script> tag dengan pattern __data__
    """
    
    # Pattern 1: window.__INITIAL_STATE__
    match = re.search(r'window\.__INITIAL_STATE__\s*=\s*({.*?});', html_content, re.DOTALL)
    if match:
        try:
            data = json.loads(match.group(1))
            return data
        except:
            pass
    
    # Pattern 2: Data di script tag
    script_pattern = re.findall(r'<script[^>]*>(.*?)</script>', html_content, re.DOTALL)
    for script in script_pattern:
        if 'itemData' in script or 'product' in script:
            try:
                # Cari JSON object
                json_match = re.search(r'({.*"name".*"price".*})', script)
                if json_match:
                    return json.loads(json_match.group(1))
            except:
                pass
    
    return None

def extract_with_regex(html_content):
    """Extract data menggunakan regex sebagai fallback"""
    
    data = {}
    
    # Extract nama produk
    nama_match = re.search(r'<h1[^>]*>([^<]+)</h1>', html_content)
    if nama_match:
        data['nama'] = nama_match.group(1).strip()
    
    # Extract harga (format: Rp x.xxx.xxx)
    harga_match = re.search(r'Rp[\s]*([\d.]+(?:,\d+)?)', html_content)
    if harga_match:
        data['harga'] = f"Rp {harga_match.group(1)}"
    
    # Extract rating (format: 4.5 atau 4,5)
    rating_match = re.search(r'(\d+[.,]\d+)\s*(?:bintang|★)', html_content, re.IGNORECASE)
    if rating_match:
        data['rating'] = rating_match.group(1)
    
    # Extract terjual (format: "123 Terjual" atau "terjual")
    terjual_match = re.search(r'([\d.]+\s*(?:ribu|juta)?)\s*terjual', html_content, re.IGNORECASE)
    if terjual_match:
        data['terjual'] = terjual_match.group(1)
    
    # Extract stok
    stok_match = re.search(r'stok\s*[:\s]*(\d+)', html_content, re.IGNORECASE)
    if stok_match:
        data['stok'] = stok_match.group(1)
    
    return data

def scrape_shopee(url):
    """Main scrape function"""
    
    print(f"[*] Fetching URL...")
    
    try:
        response = requests.get(
            url,
            headers=HEADERS,
            cookies=COOKIES,
            timeout=15,
            allow_redirects=True
        )
        
        print(f"[+] Status Code: {response.status_code}")
        
        if response.status_code != 200:
            print(f"[-] Request gagal: {response.status_code}")
            return {"status": "error", "message": f"HTTP {response.status_code}"}
        
        html = response.text
        
        result = {
            "timestamp": datetime.now().isoformat(),
            "status": "success",
            "source": "HTML parsing",
            "product": {}
        }
        
        # Coba extract dari JSON dulu
        print("[*] Parsing JSON dari HTML...")
        json_data = extract_from_json_in_html(html)
        
        if json_data:
            print("[+] JSON found!")
            result['source'] = "JSON extracted"
            # Parse JSON
            # Struktur bisa berbeda, coba multiple path
            if 'item' in json_data:
                item = json_data['item']
                result['product'] = {
                    "nama": item.get('name', 'N/A'),
                    "harga": item.get('price', 'N/A'),
                    "rating": item.get('rating_star', 'N/A'),
                    "terjual": item.get('sold', 'N/A'),
                    "stok": item.get('stock', 'N/A'),
                }
        
        # Fallback ke regex parsing
        if not result['product']:
            print("[*] Fallback ke regex parsing...")
            regex_data = extract_with_regex(html)
            result['product'] = regex_data
            result['source'] = "Regex fallback"
        
        # HTML parsing sebagai last resort
        if not result['product']:
            print("[*] Parsing HTML dengan BeautifulSoup...")
            soup = BeautifulSoup(html, 'html.parser')
            
            # Cari title
            title = soup.find('title')
            if title:
                result['product']['nama'] = title.string.split('|')[0].strip()
            
            # Cari og:price meta tag
            price_meta = soup.find('meta', property='product:price:amount')
            if price_meta:
                result['product']['harga'] = f"Rp {price_meta.get('content', 'N/A')}"
            
            # Cari og:image
            img_meta = soup.find('meta', property='og:image')
            if img_meta:
                result['product']['gambar'] = img_meta.get('content', 'N/A')
            
            result['source'] = "BeautifulSoup HTML parsing"
        
        return result
        
    except requests.exceptions.Timeout:
        return {"status": "error", "message": "Request timeout"}
    except requests.exceptions.ConnectionError:
        return {"status": "error", "message": "Connection error"}
    except Exception as e:
        return {"status": "error", "message": str(e)}

def display_result(result):
    """Display hasil scraping"""
    
    print("\n" + "=" * 70)
    print("HASIL SCRAPING SHOPEE")
    print("=" * 70)
    
    if result['status'] == 'error':
        print(f"[-] Error: {result['message']}")
        return
    
    print(f"Status: {result['status']}")
    print(f"Source: {result['source']}")
    print(f"Timestamp: {result['timestamp']}\n")
    
    product = result.get('product', {})
    
    if product:
        for key, value in product.items():
            if isinstance(value, str) and len(value) > 100:
                print(f"{key:15}: {value[:100]}...")
            else:
                print(f"{key:15}: {value}")
    else:
        print("[-] Tidak ada data produk yang ditemukan")
    
    print("\n" + "=" * 70)

def main():
    url = "https://shopee.co.id/Nusseyba-Tahira-Top-All-Size-Bahan-Rayyon-Kerah-Lengkung-Elegan-i.170820234.49403264687?extraParams=%7B%22display_model_id%22%3A445313505035%2C%22model_selection_logic%22%3A3%7D&sp_atk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba&xptdk=d27ca7a1-6c67-4dd6-971e-e6bed60bb2ba"
    
    print("=" * 70)
    print("SHOPEE PRODUCT SCRAPER v2 (Simple Requests + Regex)")
    print("=" * 70 + "\n")
    
    result = scrape_shopee(url)
    display_result(result)
    
    # Save hasil
    output_file = "tests/shopee_product_data_simple.json"
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(result, f, indent=2, ensure_ascii=False)
    
    print(f"\n[+] Data tersimpan di: {output_file}")

if __name__ == "__main__":
    main()
