import requests
import json
import sys

BASE_URL = "http://localhost/api"
USERNAME = "yumna"
PASSWORD = "password123"

def login():
    resp = None
    try:
        resp = requests.post(f"{BASE_URL}/auth/login", json={"username": USERNAME, "password": PASSWORD})
        resp.raise_for_status()
        data = resp.json()
        
        # Check for token at top level (Go backend format)
        if data.get("success") and "token" in data:
            return data["token"]
        elif "access_token" in data:
             return data["access_token"]
        # Legacy/Node format check (just in case)
        elif data.get("success") and "data" in data and "token" in data["data"]:
            return data["data"]["token"]
        else:
            print(f"Login failed: {data}")
            sys.exit(1)
    except Exception as e:
        print(f"Login error: {e}")
        if resp:
            print(f"Response text: {resp.text}")
        sys.exit(1)

def create_product(token, title, status):
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "title": title,
        "description": f"Description for {title}",
        "status": status,
        "images": [],
        "skus": [
            {
                "seller_sku": f"SKU-{title.replace(' ', '-')}",
                "variant_name": "Default",
                "variant_data": {},
                "price": 100.0,
                "stock": 10
            }
        ]
    }
    
    resp = requests.post(f"{BASE_URL}/master-products", json=payload, headers=headers)
    if resp.status_code == 201 or resp.status_code == 200:
        print(f"Created product: {title} ({status})")
    else:
        print(f"Failed to create {title}: {resp.status_code} - {resp.text}")

def main():
    print("Logging in...")
    token = login()
    print("Logged in. Seeding products...")
    
    products = [
        ("Product Active 1", "active"),
        ("Product Draft 1", "draft"),
        ("Product Archived 1", "archived"),
        ("Product Active 2", "active")
    ]
    
    for title, status in products:
        create_product(token, title, status)
        
if __name__ == "__main__":
    main()
