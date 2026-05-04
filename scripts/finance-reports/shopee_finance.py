"""
Shopee Finance Data Fetcher.
Pulls order data and escrow details for a given month.

Strategy:
1. Sync orders from Shopee API (triggers backend to fetch fresh data)
2. Get order list from local DB
3. Get escrow details for each order (calls Shopee API via backend)
4. Combine into report data
"""

import time
from typing import Optional

from api_client import OmniAPIClient
from config import (
    WALLET_TRANSACTIONS_ENDPOINT,
    ESCROW_BATCH_ENDPOINT,
    WALLET_PAGE_SIZE,
    ESCROW_BATCH_SIZE,
    SETTLED_TRANSACTION_TYPES,
)


class ShopeeFinanceFetcher:
    """Fetches Shopee financial data via OMNI backend API."""

    def __init__(self, client: OmniAPIClient):
        self.client = client

    def sync_orders(self) -> dict:
        """Trigger order sync from Shopee API to local DB."""
        print("\n[SYNC] Syncing orders from Shopee...")
        resp = self.client.post("/api/orders/sync-all")

        if not resp.get("success") and resp.get("code") != "PARTIAL_SYNC_FAILURE":
            print(f"[SYNC] Warning: {resp.get('error', 'Unknown')}")
            return {}

        sync_data = resp.get("data", {})
        shopee_count = 0
        for category, cat_data in sync_data.items():
            if isinstance(cat_data, dict) and "data" in cat_data:
                shopee_info = cat_data["data"].get("shopee", {})
                shopee_count += shopee_info.get("count", 0)

        print(f"[SYNC] Synced {shopee_count} Shopee orders")
        return sync_data

    def get_shopee_orders(self, month: int, year: int) -> list[dict]:
        """Get all Shopee orders, filtered by month from order_sn prefix."""
        print(f"\n[ORDERS] Fetching Shopee orders for {month:02d}/{year}...")

        all_orders = []
        page = 1
        while True:
            resp = self.client.get("/api/shopee/orders", {"page": str(page), "page_size": "50"})
            if not resp.get("success"):
                print(f"[ORDERS] Error: {resp.get('error', 'Unknown')}")
                break
            orders = resp.get("data", [])
            if not orders:
                break
            all_orders.extend(orders)
            meta = resp.get("meta", {})
            total = meta.get("total", len(all_orders))
            if page * 50 >= total:
                break
            page += 1

        # Filter by order_sn date prefix: YYMMDD
        yy = str(year)[2:]
        month_prefixes = [f"{yy}{month:02d}{day:02d}" for day in range(1, 32)]
        filtered = [o for o in all_orders if any(o.get("order_sn", "").startswith(p) for p in month_prefixes)]

        print(f"[ORDERS] Total in DB: {len(all_orders)}, For {month:02d}/{year}: {len(filtered)}")
        return filtered

    def get_escrow_details_batch(self, order_sns: list[str]) -> list[dict]:
        """Get escrow details for orders in batches of 50."""
        if not order_sns:
            return []

        print(f"\n[ESCROW] Fetching escrow details for {len(order_sns)} orders...")
        all_details = []

        for i in range(0, len(order_sns), ESCROW_BATCH_SIZE):
            batch = order_sns[i:i + ESCROW_BATCH_SIZE]
            batch_num = (i // ESCROW_BATCH_SIZE) + 1
            total_batches = (len(order_sns) + ESCROW_BATCH_SIZE - 1) // ESCROW_BATCH_SIZE
            print(f"  [Batch {batch_num}/{total_batches}] Processing {len(batch)} orders...")

            try:
                resp = self.client.post(ESCROW_BATCH_ENDPOINT, {"order_sn_list": batch})
                if not resp.get("success"):
                    print(f"  [WARN] Batch {batch_num} failed: {resp.get('error', 'Unknown')}")
                    continue
                details = resp.get("data", {}).get("responses", [])
                all_details.extend(details)
            except Exception as e:
                print(f"  [WARN] Batch {batch_num} error: {e}")
                break

            if i + ESCROW_BATCH_SIZE < len(order_sns):
                time.sleep(1.0)

        print(f"[ESCROW] Got escrow details for {len(all_details)} orders")
        return all_details

    def get_wallet_transactions(self, month: int, year: int) -> list[dict]:
        """Try to get wallet transactions (supplementary)."""
        print(f"\n[WALLET] Checking wallet transactions for {month:02d}/{year}...")
        start_date = f"{year}-{month:02d}-01"
        end_date = f"{year}-{month + 1:02d}-01" if month < 12 else f"{year + 1}-01-01"

        resp = self.client.get(WALLET_TRANSACTIONS_ENDPOINT, {
            "start_date": start_date, "end_date": end_date, "page_size": "100",
        })
        if not resp.get("success"):
            return []
        transactions = resp.get("data", {}).get("transactions", [])
        print(f"[WALLET] Found {len(transactions)} wallet transactions")
        return transactions

    def get_full_settlement_report(self, month: int, year: int) -> dict:
        """
        Complete settlement report:
        1. Sync fresh orders from Shopee
        2. Get orders for the target month
        3. Get escrow details for each order
        4. Try wallet transactions as supplement
        5. Build summary
        """
        self.sync_orders()
        orders = self.get_shopee_orders(month, year)
        order_sns = [o.get("order_sn") for o in orders if o.get("order_sn")]
        escrow_details = self.get_escrow_details_batch(order_sns)
        wallet_tx = self.get_wallet_transactions(month, year)

        total_escrow = sum(d.get("total_amount", 0) for d in escrow_details)
        total_items = sum(
            sum(item.get("quantity", 0) for item in d.get("items", []))
            for d in escrow_details
        )
        total_order_amount = sum(o.get("total_amount", 0) for o in orders)

        summary = {
            "period": f"{month:02d}/{year}",
            "order_count": len(orders),
            "escrow_count": len(escrow_details),
            "total_order_amount": total_order_amount,
            "total_escrow_amount": total_escrow,
            "total_items_sold": total_items,
            "wallet_tx_count": len(wallet_tx),
            "platform": "Shopee",
        }

        print(f"\n{'='*60}")
        print(f"[SUMMARY] Period: {summary['period']}")
        print(f"  Total Orders:        {summary['order_count']}")
        print(f"  Escrow Data:         {summary['escrow_count']} orders")
        print(f"  Total Order Amount:  Rp {total_order_amount:,.0f}")
        print(f"  Total Escrow (Cair): Rp {total_escrow:,.0f}")
        print(f"  Total Items Sold:    {total_items} pcs")
        print(f"  Wallet Transactions: {summary['wallet_tx_count']}")
        print(f"{'='*60}")

        return {
            "orders": orders,
            "escrow_details": escrow_details,
            "wallet_transactions": wallet_tx,
            "summary": summary,
        }
