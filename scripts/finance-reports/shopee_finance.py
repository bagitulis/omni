"""
Shopee Finance Data Fetcher.
Pulls COMPLETED (dana cair) orders + margin analysis vs harga modal.

Flow:
1. Trigger escrow sync via /api/analytics/shopee/sync
2. Query escrow data from DB
3. Query margin analysis per SKU per qty tier
4. Read Google Sheets for harga modal per SKU
"""

import time
import subprocess
import sys
from pathlib import Path

from api_client import OmniAPIClient

DB_CONTAINER = "omni-postgres"
DB_USER = "omni"
DB_NAME = "omni_main"
# SA_FILE relative to project root (this script lives in scripts/finance-reports/)
_SCRIPT_DIR = Path(__file__).resolve().parent
_PROJECT_ROOT = _SCRIPT_DIR.parent.parent
SA_FILE = str(_PROJECT_ROOT / "backend" / "config" / "static" / "google" / "bertigahemat-f1bd6932b229.json")
SHEET_ALL_PRODUCT = "1H3TltJKcUfnrizmgRNXqjsHgeR44wNZYmfJ784SPXiI"


def run_db_query(query: str) -> str:
    cmd = ["docker", "exec", DB_CONTAINER, "psql", "-U", DB_USER, "-d", DB_NAME, "-t", "-A", "-F", "\t", "-c", query]
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
    if result.returncode != 0:
        raise RuntimeError(f"DB query failed: {result.stderr}")
    return result.stdout.strip()


def _extract_date(sn: str) -> str:
    return f"{sn[4:6]}/{sn[2:4]}/20{sn[0:2]}" if len(sn) >= 6 else ""


def _f(v): 
    try: return float(v) if v else 0.0
    except: return 0.0


def _i(v):
    try: return int(v) if v else 0
    except: return 0


def _parse_rp(v):
    if not v: return 0
    v = v.replace("Rp", "").replace(" ", "").replace(".", "").replace(",", "")
    try: return int(v)
    except: return 0

def _extract_isi(nama: str) -> int:
    """Extract isi per renceng/pack dari nama produk. Returns 1 jika satuan."""
    import re
    n = nama.lower()
    # 'isi 12', 'isi 10', 'Isi 6'
    m = re.search(r'isi\s*(\d+)', n)
    if m: return int(m.group(1))
    # '(6+1)', '12+1' — renceng promo (angka pertama = isi)
    m = re.search(r'(\d+)\+\d+', nama)
    if m:
        v = int(m.group(1))
        if v <= 20: return v
    return 1

class ShopeeFinanceFetcher:

    def __init__(self, client: OmniAPIClient):
        self.client = client
        self.schema = f"tenant_{client.tenant_id}"

    def trigger_escrow_sync(self, month, year):
        print(f"\n[SYNC] Checking escrow sync for {month:02d}/{year}...")
        s = self.client.get("/api/analytics/shopee/sync-status", {"month": str(month), "year": str(year)})
        if s.get("data", {}).get("synced"):
            print(f"[SYNC] Already synced: {s['data'].get('total_orders', 0)} orders.")
            return True
        print("[SYNC] Triggering sync...")
        self.client.post("/api/analytics/shopee/sync", {"month": month, "year": year, "force_resync": False})
        for i in range(60):
            time.sleep(3)
            s = self.client.get("/api/analytics/shopee/sync-status", {"month": str(month), "year": str(year)})
            if s.get("data", {}).get("synced"):
                print(f"[SYNC] Done: {s['data'].get('total_orders', 0)} orders")
                return True
            if i % 5 == 0: print(f"  Syncing... ({i*3}s)")
        return False

    def get_escrow_orders(self, month, year):
        print("\n[DB] Querying escrow orders...")
        raw = run_db_query(f"""
        SELECT o.order_sn, o.escrow_amount, o.commission_fee, o.service_fee,
               o.seller_processing_fee, o.actual_shipping_fee,
               o.shopee_shipping_rebate, o.buyer_total_amount, o.buyer_user_name
        FROM {self.schema}.shopee_escrow_orders o
        WHERE o.month={month} AND o.year={year} ORDER BY o.order_sn;""")
        if not raw: return []
        orders = []
        for line in raw.split("\n"):
            c = line.split("\t")
            if len(c) < 9: continue
            orders.append({"order_sn": c[0], "order_date": _extract_date(c[0]),
                "escrow_amount": _f(c[1]), "commission_fee": _f(c[2]), "service_fee": _f(c[3]),
                "seller_processing_fee": _f(c[4]), "actual_shipping_fee": _f(c[5]),
                "shopee_shipping_rebate": _f(c[6]), "buyer_total_amount": _f(c[7]),
                "buyer_username": c[8] if c[8] else ""})
        print(f"[DB] {len(orders)} orders")
        return orders

    def get_escrow_items(self, month, year):
        print("[DB] Querying escrow items...")
        raw = run_db_query(f"""
        SELECT i.order_sn, i.item_id, i.model_id, i.item_name, i.model_name,
               i.sku, i.model_sku, i.quantity, i.original_price,
               i.selling_price, i.discounted_price, i.seller_discount, i.shopee_discount
        FROM {self.schema}.shopee_escrow_items i
        JOIN {self.schema}.shopee_escrow_orders o ON o.id = i.escrow_order_id
        WHERE o.month={month} AND o.year={year} ORDER BY i.order_sn, i.item_id;""")
        if not raw: return []
        items = []
        for line in raw.split("\n"):
            c = line.split("\t")
            if len(c) < 13: continue
            items.append({"order_sn": c[0], "item_id": c[1], "model_id": c[2],
                "item_name": c[3], "model_name": c[4], "sku": c[5], "model_sku": c[6],
                "quantity": _i(c[7]), "original_price": _f(c[8]), "selling_price": _f(c[9]),
                "discounted_price": _f(c[10]), "seller_discount": _f(c[11]), "shopee_discount": _f(c[12])})
        print(f"[DB] {len(items)} items")
        return items

    def get_margin_analysis(self, month, year):
        """Per SKU per qty-tier. Uses discounted_price as subtotal (after seller wholesale discount)."""
        print("[DB] Querying margin analysis...")
        raw = run_db_query(f"""
        WITH single_item AS (
          SELECT COALESCE(NULLIF(i.model_sku,''), i.sku) as effective_sku,
                 i.item_name, i.model_name, i.quantity,
                 i.discounted_price, i.seller_discount,
                 o.escrow_amount, o.commission_fee, o.service_fee, o.seller_processing_fee
          FROM {self.schema}.shopee_escrow_items i
          JOIN {self.schema}.shopee_escrow_orders o ON o.id = i.escrow_order_id
          WHERE o.month={month} AND o.year={year}
          AND o.id IN (SELECT escrow_order_id FROM {self.schema}.shopee_escrow_items GROUP BY escrow_order_id HAVING COUNT(*)=1)
        ), tiered AS (
          SELECT *, CASE WHEN quantity=1 THEN '1 pcs' WHEN quantity BETWEEN 2 AND 3 THEN '2-3 pcs'
            WHEN quantity BETWEEN 4 AND 5 THEN '4-5 pcs' ELSE '6+ pcs' END as tier,
            CASE WHEN quantity=1 THEN 1 WHEN quantity BETWEEN 2 AND 3 THEN 2
            WHEN quantity BETWEEN 4 AND 5 THEN 3 ELSE 4 END as tier_sort
          FROM single_item
        )
        SELECT effective_sku, item_name, model_name, tier,
          COUNT(*), SUM(quantity),
          ROUND(AVG(discounted_price::numeric/quantity),0),
          ROUND(AVG(escrow_amount::numeric/quantity),0),
          ROUND(AVG(commission_fee::numeric/quantity),0),
          ROUND(AVG(service_fee::numeric/quantity),0),
          ROUND(AVG(seller_processing_fee::numeric/quantity),0),
          ROUND(AVG(seller_discount::numeric/quantity),0)
        FROM tiered GROUP BY effective_sku, item_name, model_name, tier, tier_sort
        ORDER BY effective_sku, tier_sort;""")
        if not raw: return []
        rows = []
        for line in raw.split("\n"):
            c = line.split("\t")
            if len(c) < 12: continue
            rows.append({"sku": c[0], "item_name": c[1], "model_name": c[2], "tier": c[3],
                "jml_order": _i(c[4]), "total_qty": _i(c[5]),
                "avg_subtotal": _f(c[6]), "avg_cair": _f(c[7]),
                "avg_komisi": _f(c[8]), "avg_svc": _f(c[9]), "avg_proc": _f(c[10]),
                "avg_seller_disc": _f(c[11])})
        print(f"[DB] {len(rows)} margin rows")
        return rows

    def get_product_costs(self, skus, shopee_names):
        """Read Google Sheets ALL PRODUCT. shopee_names = {sku: nama_di_shopee} for renceng parsing."""
        print("\n[SHEETS] Reading harga modal...")
        try:
            from google.oauth2 import service_account
            from googleapiclient.discovery import build
        except ImportError:
            subprocess.check_call([sys.executable, "-m", "pip", "install", "google-auth", "google-api-python-client", "-q"])
            from google.oauth2 import service_account
            from googleapiclient.discovery import build

        creds = service_account.Credentials.from_service_account_file(SA_FILE,
            scopes=["https://www.googleapis.com/auth/spreadsheets.readonly"])
        svc = build("sheets", "v4", credentials=creds)
        data = svc.spreadsheets().values().get(spreadsheetId=SHEET_ALL_PRODUCT, range="'ALL PRODUCT'!A1:O1000").execute()

        # Header: 0=Margin%, 1=Biaya%, 2=SKU, 3=Kode, 4=Brand, 5=Nama, 6=HargaMP,
        #         7=Ukuran, 8=PkgLusin, 9=PkgPcs, 10=BeliExPPN, 11=BeliInPPN, 12=Margin, 13=HargaJual, 14=MPKarton
        costs = {}
        for row in data.get("values", [])[1:]:
            if len(row) < 14: continue
            sku = row[2]
            if sku not in skus: continue
            nama = row[5] if len(row) > 5 else ""
            harga_mp = int(row[6]) if len(row) > 6 and row[6].isdigit() else 0
            pcs = int(row[9]) if len(row) > 9 and row[9].isdigit() else 1
            beli_karton = _parse_rp(row[11]) if len(row) > 11 else 0
            jual_karton = _parse_rp(row[13]) if len(row) > 13 else 0

            # Pakai nama Shopee (lebih deskriptif) untuk detect rencengan
            nama_shopee = shopee_names.get(sku, nama)
            isi_renceng = _extract_isi(nama_shopee)
            modal_per_pcs = round(jual_karton / pcs) if pcs else 0
            modal_per_unit = modal_per_pcs * isi_renceng  # kalikan isi jika rencengan

            costs[sku] = {
                "nama_sheet": nama,
                "pcs_per_karton": pcs,
                "isi_per_renceng": isi_renceng,
                "beli_per_karton": beli_karton,
                "beli_per_unit": round(beli_karton / pcs * isi_renceng) if pcs else 0,
                "jual_karton": jual_karton,
                "modal_per_unit": modal_per_unit,  # MODAL = (Beli+3%)/pcs x isi renceng
                "harga_mp": harga_mp,
            }

        nf = [s for s in skus if s and s not in costs]
        print(f"[SHEETS] {len(costs)} ditemukan, {len(nf)} tidak: {nf if nf else '-'}")
        return costs

    def get_full_settlement_report(self, month, year):
        self.trigger_escrow_sync(month, year)
        orders = self.get_escrow_orders(month, year)
        items = self.get_escrow_items(month, year)
        margin_rows = self.get_margin_analysis(month, year)
        all_skus = list(set(r["sku"] for r in margin_rows if r["sku"]))
        # Build shopee names map for renceng detection
        shopee_names = {}
        for r in margin_rows:
            if r["sku"] and r["sku"] not in shopee_names:
                shopee_names[r["sku"]] = r["item_name"]
        costs = self.get_product_costs(all_skus, shopee_names)

        te = sum(o["escrow_amount"] for o in orders)
        total_subtotal = sum(i["discounted_price"] for i in items)  # subtotal aktual (setelah diskon seller)

        summary = {"period": f"{month:02d}/{year}", "order_count": len(orders),
            "item_count": len(items), "total_qty_sold": sum(i["quantity"] for i in items),
            "total_subtotal": total_subtotal,
            "total_escrow": te,
            "total_commission": sum(o["commission_fee"] for o in orders),
            "total_service_fee": sum(o["service_fee"] for o in orders),
            "total_processing_fee": sum(o["seller_processing_fee"] for o in orders)}

        print(f"\n{'='*55}")
        print(f"  DANA CAIR SHOPEE - {month:02d}/{year}")
        print(f"  {len(orders)} pesanan | {summary['total_qty_sold']} unit | Rp {te:,.0f}")
        print(f"{'='*55}")

        return {"orders": orders, "items": items, "summary": summary,
                "margin_rows": margin_rows, "costs": costs}
