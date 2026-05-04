#!/usr/bin/env python3
"""
OMNI Finance Report Generator - Shopee Dana Cair
=================================================
Generates detailed XLSX report of settled/disbursed funds from Shopee.

Usage:
    python main.py                    # Default: April 2026
    python main.py --month 4 --year 2026
    python main.py --month 3 --year 2026 --output custom_report.xlsx

Output:
    Multi-sheet XLSX with:
    - Sheet 1: Ringkasan (Executive Summary)
    - Sheet 2: Transaksi Dana Cair (All settled transactions)
    - Sheet 3: Detail Escrow per Order (Fee breakdown per order)
    - Sheet 4: Detail Produk per Item (Product ID, SKU, Qty, Prices)
    - Sheet 5: Rekap Potongan (Fee/Deduction summary)
"""

import argparse
import sys
import os

# Add script directory to path for imports
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from config import DEFAULT_MONTH, DEFAULT_YEAR
from api_client import OmniAPIClient
from shopee_finance import ShopeeFinanceFetcher
from xlsx_exporter import XLSXExporter


def main():
    parser = argparse.ArgumentParser(
        description="Generate Shopee Dana Cair (Settled Funds) Report"
    )
    parser.add_argument("--month", type=int, default=DEFAULT_MONTH, help="Report month (1-12)")
    parser.add_argument("--year", type=int, default=DEFAULT_YEAR, help="Report year")
    parser.add_argument("--output", type=str, default=None, help="Custom output filename")
    parser.add_argument("--base-url", type=str, default=None, help="Override base URL")
    parser.add_argument("--username", type=str, default=None, help="Override username")
    parser.add_argument("--password", type=str, default=None, help="Override password")

    args = parser.parse_args()

    # Validate
    if not 1 <= args.month <= 12:
        print("[ERROR] Month must be between 1 and 12")
        sys.exit(1)
    if args.year < 2020:
        print("[ERROR] Year must be 2020 or later")
        sys.exit(1)

    print("=" * 70)
    print("  OMNI FINANCE REPORT GENERATOR")
    print("  Shopee - Dana Cair (Settled Funds)")
    print("=" * 70)
    print()

    # Initialize API client
    kwargs = {}
    if args.base_url:
        kwargs["base_url"] = args.base_url
    if args.username:
        kwargs["username"] = args.username
    if args.password:
        kwargs["password"] = args.password

    client = OmniAPIClient(**kwargs)

    # Login
    if not client.login():
        print("\n[FATAL] Cannot authenticate. Aborting.")
        sys.exit(1)

    # Fetch data
    fetcher = ShopeeFinanceFetcher(client)
    report_data = fetcher.get_full_settlement_report(args.month, args.year)

    # Check if we got data
    if report_data["summary"]["order_count"] == 0:
        print("\n[WARN] No orders found for this period.")
        print("       Generating report with available data anyway...")

    # Generate XLSX
    exporter = XLSXExporter()
    filepath = exporter.generate_full_report(report_data, args.month, args.year, args.output)

    print()
    print("=" * 70)
    print(f"  REPORT GENERATED SUCCESSFULLY")
    print(f"  File: {filepath}")
    print("=" * 70)
    print()
    print("Sheets included:")
    print("  1. Ringkasan       - Summary pesanan, escrow, items terjual")
    print("  2. Daftar Pesanan  - Semua order + escrow amount + produk + SKU")
    print("  3. Detail Escrow   - Breakdown biaya per order")
    print("  4. Detail Produk   - Product ID, SKU, qty, harga per item")
    print("  4. Detail Produk      - Product ID, SKU, qty, harga per item")
    print("  5. Rekap Potongan     - Total semua jenis potongan")


if __name__ == "__main__":
    main()
