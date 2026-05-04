#!/usr/bin/env python3
"""
OMNI Finance Report Generator - Shopee Dana Cair
=================================================
Generates detailed XLSX report of settled/disbursed funds from Shopee.

Usage:
    python main.py                          # Default: April 2026
    python main.py --month 4 --year 2026
    python main.py --output custom.xlsx

Output (4 sheets):
    1. Ringkasan       - Executive summary: totals, fee breakdown
    2. Daftar Pesanan  - Per order: escrow, komisi, service fee, ongkir
    3. Detail Produk   - Per item: product ID, SKU, qty, harga, diskon
    4. Rekap Potongan  - Fee summary with percentages
"""

import argparse
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from config import DEFAULT_MONTH, DEFAULT_YEAR
from api_client import OmniAPIClient
from shopee_finance import ShopeeFinanceFetcher
from xlsx_exporter import XLSXExporter


def main():
    parser = argparse.ArgumentParser(description="Generate Shopee Dana Cair Report")
    parser.add_argument("--month", type=int, default=DEFAULT_MONTH, help="Report month (1-12)")
    parser.add_argument("--year", type=int, default=DEFAULT_YEAR, help="Report year")
    parser.add_argument("--output", type=str, default=None, help="Custom output filename")
    parser.add_argument("--base-url", type=str, default=None, help="Override base URL")
    parser.add_argument("--username", type=str, default=None, help="Override username")
    parser.add_argument("--password", type=str, default=None, help="Override password")
    args = parser.parse_args()

    if not 1 <= args.month <= 12:
        print("[ERROR] Month must be between 1 and 12")
        sys.exit(1)

    print("=" * 60)
    print("  OMNI FINANCE REPORT GENERATOR")
    print("  Shopee - Dana Cair (Settled Funds)")
    print("=" * 60)

    # Init client
    kwargs = {}
    if args.base_url:
        kwargs["base_url"] = args.base_url
    if args.username:
        kwargs["username"] = args.username
    if args.password:
        kwargs["password"] = args.password

    client = OmniAPIClient(**kwargs)
    if not client.login():
        print("\n[FATAL] Cannot authenticate. Aborting.")
        sys.exit(1)

    # Fetch data
    fetcher = ShopeeFinanceFetcher(client)
    report_data = fetcher.get_full_settlement_report(args.month, args.year)

    if report_data["summary"]["order_count"] == 0:
        print("\n[WARN] No completed orders found for this period.")
        print("       Try running 'Sync Escrow' from the web UI first.")
        sys.exit(1)

    # Generate XLSX
    exporter = XLSXExporter()
    filepath = exporter.generate_full_report(report_data, args.month, args.year, args.output)

    s = report_data["summary"]
    print()
    print("=" * 60)
    print(f"  REPORT GENERATED SUCCESSFULLY")
    print(f"  File: {filepath}")
    print("=" * 60)
    print()
    print("Sheets:")
    print(f"  1. Ringkasan         - {s['order_count']} pesanan, Rp {s['total_escrow']:,.0f} dana cair")
    print(f"  2. Daftar Transaksi  - Per order: subtotal produk, biaya admin/layanan/proses, dana cair")
    print(f"  3. Analisis Margin   - Per SKU per tier qty: modal vs cair, % biaya, status")
    print(f"  4. Data Modal        - Referensi harga modal dari Sheet ALL PRODUCT")


if __name__ == "__main__":
    main()
