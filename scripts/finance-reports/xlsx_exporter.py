"""
XLSX Report Exporter.
Generates professional multi-sheet Excel reports for company finance reporting.
"""

import os
import sys
from datetime import datetime
from typing import Optional

try:
    from openpyxl import Workbook
    from openpyxl.styles import Font, PatternFill, Alignment, Border, Side, numbers
    from openpyxl.utils import get_column_letter
except ImportError:
    print("Installing openpyxl...")
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "openpyxl", "-q"])
    from openpyxl import Workbook
    from openpyxl.styles import Font, PatternFill, Alignment, Border, Side, numbers
    from openpyxl.utils import get_column_letter

from config import (
    HEADER_FILL_COLOR,
    HEADER_FONT_COLOR,
    CURRENCY_FORMAT,
    FEE_CATEGORIES,
    OUTPUT_DIR,
)


class XLSXExporter:
    """Generates professional XLSX finance reports."""

    def __init__(self):
        self.wb = Workbook()
        # Remove default sheet
        self.wb.remove(self.wb.active)

        # Styles
        self.header_font = Font(name="Calibri", size=11, bold=True, color=HEADER_FONT_COLOR)
        self.header_fill = PatternFill(start_color=HEADER_FILL_COLOR, end_color=HEADER_FILL_COLOR, fill_type="solid")
        self.header_alignment = Alignment(horizontal="center", vertical="center", wrap_text=True)
        self.currency_font = Font(name="Calibri", size=10)
        self.title_font = Font(name="Calibri", size=14, bold=True)
        self.subtitle_font = Font(name="Calibri", size=11, italic=True)
        self.thin_border = Border(
            left=Side(style="thin"),
            right=Side(style="thin"),
            top=Side(style="thin"),
            bottom=Side(style="thin"),
        )

    def _style_header_row(self, ws, row: int, col_count: int):
        """Apply header styling to a row."""
        for col in range(1, col_count + 1):
            cell = ws.cell(row=row, column=col)
            cell.font = self.header_font
            cell.fill = self.header_fill
            cell.alignment = self.header_alignment
            cell.border = self.thin_border

    def _auto_width(self, ws):
        """Auto-adjust column widths based on content."""
        for col in ws.columns:
            max_length = 0
            col_letter = get_column_letter(col[0].column)
            for cell in col:
                if cell.value:
                    cell_len = len(str(cell.value))
                    max_length = max(max_length, min(cell_len, 50))
            ws.column_dimensions[col_letter].width = max(max_length + 3, 12)

    def _apply_currency_format(self, ws, col: int, start_row: int, end_row: int):
        """Apply currency format to a column range."""
        for row in range(start_row, end_row + 1):
            cell = ws.cell(row=row, column=col)
            cell.number_format = CURRENCY_FORMAT

    def create_summary_sheet(self, summary: dict, month: int, year: int):
        """Sheet 1: Executive Summary."""
        ws = self.wb.create_sheet("Ringkasan")

        # Title
        ws.cell(row=1, column=1, value="LAPORAN KEUANGAN - SHOPEE")
        ws.cell(row=1, column=1).font = self.title_font
        ws.merge_cells("A1:D1")

        ws.cell(row=2, column=1, value=f"Periode: {_month_name(month)} {year}")
        ws.cell(row=2, column=1).font = self.subtitle_font

        ws.cell(row=3, column=1, value=f"Generated: {datetime.now().strftime('%d %B %Y %H:%M WIB')}")
        ws.cell(row=3, column=1).font = self.subtitle_font

        # Summary table
        row = 5
        headers = ["Kategori", "Jumlah", "Total (Rp)", "Keterangan"]
        for col, h in enumerate(headers, 1):
            ws.cell(row=row, column=col, value=h)
        self._style_header_row(ws, row, len(headers))

        rows_data = [
            ("Total Pesanan", summary.get("order_count", 0), summary.get("total_order_amount", 0), "Semua pesanan Shopee di periode ini"),
            ("Escrow Amount (Dana Cair)", summary.get("escrow_count", 0), summary.get("total_escrow_amount", 0), "Jumlah bersih yang diterima seller"),
            ("Total Item Terjual", summary.get("total_items_sold", 0), "-", "Total qty semua produk"),
            ("Wallet Transactions", summary.get("wallet_tx_count", 0), "-", "Transaksi wallet (jika tersedia)"),
        ]

        for i, (cat, count, total, note) in enumerate(rows_data, 1):
            r = row + i
            ws.cell(row=r, column=1, value=cat)
            ws.cell(row=r, column=2, value=count)
            ws.cell(row=r, column=3, value=total)
            if isinstance(total, (int, float)):
                ws.cell(row=r, column=3).number_format = CURRENCY_FORMAT
            ws.cell(row=r, column=4, value=note)

            if "Escrow" in cat:
                for col in range(1, 5):
                    ws.cell(row=r, column=col).font = Font(bold=True, size=11)

        self._auto_width(ws)

    def create_transactions_sheet(self, orders: list[dict], escrow_details: list[dict]):
        """Sheet 2: Orders with Escrow Overview."""
        ws = self.wb.create_sheet("Daftar Pesanan")

        headers = [
            "No", "Order SN", "Status", "Buyer",
            "Total Amount (Rp)", "Escrow Amount (Rp)",
            "Produk", "SKU", "Qty", "Kurir", "Payment"
        ]

        for col, h in enumerate(headers, 1):
            ws.cell(row=1, column=col, value=h)
        self._style_header_row(ws, 1, len(headers))

        for i, (order, escrow) in enumerate(zip(orders, escrow_details), 1):
            row = i + 1
            ws.cell(row=row, column=1, value=i)
            ws.cell(row=row, column=2, value=order.get("order_sn", ""))
            ws.cell(row=row, column=3, value=order.get("order_status", ""))
            ws.cell(row=row, column=4, value=escrow.get("buyer_username", order.get("buyer_username", "")))
            ws.cell(row=row, column=5, value=order.get("total_amount", 0))
            ws.cell(row=row, column=5).number_format = CURRENCY_FORMAT
            ws.cell(row=row, column=6, value=escrow.get("total_amount", 0))
            ws.cell(row=row, column=6).number_format = CURRENCY_FORMAT
            # First item info
            items = escrow.get("items", [])
            if items:
                ws.cell(row=row, column=7, value=items[0].get("item_name", ""))
                ws.cell(row=row, column=8, value=items[0].get("sku", ""))
                ws.cell(row=row, column=9, value=sum(it.get("quantity", 0) for it in items))
            ws.cell(row=row, column=10, value=order.get("shipping_carrier", ""))
            ws.cell(row=row, column=11, value=order.get("payment_method", ""))

            for col in range(1, len(headers) + 1):
                ws.cell(row=row, column=col).border = self.thin_border

        # Total row
        if orders:
            total_row = len(orders) + 2
            ws.cell(row=total_row, column=4, value="TOTAL:")
            ws.cell(row=total_row, column=4).font = Font(bold=True)
            ws.cell(row=total_row, column=5, value=sum(o.get("total_amount", 0) for o in orders))
            ws.cell(row=total_row, column=5).number_format = CURRENCY_FORMAT
            ws.cell(row=total_row, column=5).font = Font(bold=True)
            ws.cell(row=total_row, column=6, value=sum(e.get("total_amount", 0) for e in escrow_details))
            ws.cell(row=total_row, column=6).number_format = CURRENCY_FORMAT
            ws.cell(row=total_row, column=6).font = Font(bold=True)

        self._auto_width(ws)

    def create_escrow_detail_sheet(self, escrow_details: list[dict]):
        """Sheet 3: Escrow Detail per Order with Item Breakdown."""
        ws = self.wb.create_sheet("Detail Escrow per Order")

        headers = [
            "No", "Order SN", "Buyer", "Status",
            "Escrow Amount (Rp)", "Komisi (Rp)", "Biaya Layanan (Rp)",
            "Biaya Transaksi (Rp)", "Biaya Proses (Rp)",
            "Ongkir Aktual (Rp)", "Subsidi Ongkir (Rp)",
            "Total Pembeli (Rp)", "Metode Bayar"
        ]

        for col, h in enumerate(headers, 1):
            ws.cell(row=1, column=col, value=h)
        self._style_header_row(ws, 1, len(headers))

        for i, detail in enumerate(escrow_details, 1):
            row = i + 1
            income = detail.get("order_income", {})

            ws.cell(row=row, column=1, value=i)
            ws.cell(row=row, column=2, value=detail.get("order_sn", ""))
            ws.cell(row=row, column=3, value=detail.get("buyer_username", ""))
            ws.cell(row=row, column=4, value=detail.get("order_status", ""))
            ws.cell(row=row, column=5, value=income.get("escrow_amount", 0))
            ws.cell(row=row, column=6, value=income.get("commission_fee", 0))
            ws.cell(row=row, column=7, value=income.get("service_fee", 0))
            ws.cell(row=row, column=8, value=income.get("seller_transaction_fee", 0))
            ws.cell(row=row, column=9, value=income.get("seller_order_processing_fee", 0))
            ws.cell(row=row, column=10, value=income.get("actual_shipping_fee", 0))
            ws.cell(row=row, column=11, value=income.get("shopee_shipping_rebate", 0))
            ws.cell(row=row, column=12, value=income.get("buyer_total_amount", 0))
            ws.cell(row=row, column=13, value=income.get("buyer_payment_method", ""))

            # Currency format for money columns
            for col in range(5, 13):
                ws.cell(row=row, column=col).number_format = CURRENCY_FORMAT

            # Border
            for col in range(1, len(headers) + 1):
                ws.cell(row=row, column=col).border = self.thin_border

        # Totals
        if escrow_details:
            total_row = len(escrow_details) + 2
            ws.cell(row=total_row, column=4, value="TOTAL:")
            ws.cell(row=total_row, column=4).font = Font(bold=True)
            for col_idx, key in [(5, "escrow_amount"), (6, "commission_fee"), (7, "service_fee"),
                                  (8, "seller_transaction_fee"), (9, "seller_order_processing_fee"),
                                  (10, "actual_shipping_fee"), (11, "shopee_shipping_rebate"),
                                  (12, "buyer_total_amount")]:
                total = sum(d.get("order_income", {}).get(key, 0) for d in escrow_details)
                ws.cell(row=total_row, column=col_idx, value=total)
                ws.cell(row=total_row, column=col_idx).number_format = CURRENCY_FORMAT
                ws.cell(row=total_row, column=col_idx).font = Font(bold=True)

        self._auto_width(ws)

    def create_item_detail_sheet(self, escrow_details: list[dict]):
        """Sheet 4: Item-level detail (Product ID, SKU, Qty, Prices)."""
        ws = self.wb.create_sheet("Detail Produk per Item")

        headers = [
            "No", "Order SN", "Item ID", "Model ID", "Nama Produk",
            "Varian/Model", "SKU (Item)", "SKU (Model)",
            "Qty", "Harga Asli (Rp)", "Harga Jual (Rp)", "Harga Diskon (Rp)",
            "Diskon Seller (Rp)", "Diskon Shopee (Rp)",
            "Diskon Coin (Rp)", "Voucher Seller (Rp)", "Voucher Shopee (Rp)",
            "Komisi AMS (Rp)", "Biaya Proses (Rp)"
        ]

        for col, h in enumerate(headers, 1):
            ws.cell(row=1, column=col, value=h)
        self._style_header_row(ws, 1, len(headers))

        row_num = 2
        for detail in escrow_details:
            order_sn = detail.get("order_sn", "")
            items = detail.get("items", [])

            for item in items:
                ws.cell(row=row_num, column=1, value=row_num - 1)
                ws.cell(row=row_num, column=2, value=order_sn)
                ws.cell(row=row_num, column=3, value=item.get("item_id", ""))
                ws.cell(row=row_num, column=4, value=item.get("model_id", ""))
                ws.cell(row=row_num, column=5, value=item.get("item_name", ""))
                ws.cell(row=row_num, column=6, value=item.get("model_name", ""))
                ws.cell(row=row_num, column=7, value=item.get("item_sku", ""))
                ws.cell(row=row_num, column=8, value=item.get("model_sku", item.get("sku", "")))
                ws.cell(row=row_num, column=9, value=item.get("quantity", item.get("quantity_purchased", 0)))
                ws.cell(row=row_num, column=10, value=item.get("original_price", 0))
                ws.cell(row=row_num, column=11, value=item.get("selling_price", 0))
                ws.cell(row=row_num, column=12, value=item.get("discounted_price", 0))
                ws.cell(row=row_num, column=13, value=item.get("seller_discount", 0))
                ws.cell(row=row_num, column=14, value=item.get("shopee_discount", 0))
                ws.cell(row=row_num, column=15, value=item.get("discount_from_coin", 0))
                ws.cell(row=row_num, column=16, value=item.get("discount_from_voucher_seller", 0))
                ws.cell(row=row_num, column=17, value=item.get("discount_from_voucher_shopee", 0))
                ws.cell(row=row_num, column=18, value=item.get("ams_commission_fee", 0))
                ws.cell(row=row_num, column=19, value=item.get("seller_order_processing_fee", 0))

                # Currency format
                for col in range(10, 20):
                    ws.cell(row=row_num, column=col).number_format = CURRENCY_FORMAT

                # Border
                for col in range(1, len(headers) + 1):
                    ws.cell(row=row_num, column=col).border = self.thin_border

                row_num += 1

        # Summary at bottom
        if row_num > 2:
            ws.cell(row=row_num + 1, column=1, value=f"Total Items: {row_num - 2}")
            ws.cell(row=row_num + 1, column=1).font = Font(bold=True)
            total_qty = 0
            for detail in escrow_details:
                for item in detail.get("items", []):
                    total_qty += item.get("quantity", item.get("quantity_purchased", 0))
            ws.cell(row=row_num + 1, column=9, value=total_qty)
            ws.cell(row=row_num + 1, column=9).font = Font(bold=True)

        self._auto_width(ws)

    def create_fee_breakdown_sheet(self, escrow_details: list[dict]):
        """Sheet 5: Fee/Deduction Breakdown Summary."""
        ws = self.wb.create_sheet("Rekap Potongan")

        # Title
        ws.cell(row=1, column=1, value="REKAP POTONGAN & BIAYA")
        ws.cell(row=1, column=1).font = self.title_font
        ws.merge_cells("A1:C1")

        # Fee summary table
        headers = ["Jenis Potongan", "Total (Rp)", "Jumlah Order Terkena"]
        row = 3
        for col, h in enumerate(headers, 1):
            ws.cell(row=row, column=col, value=h)
        self._style_header_row(ws, row, len(headers))

        # Calculate fee totals
        fee_totals = {}
        fee_counts = {}
        for key, label in FEE_CATEGORIES.items():
            total = 0
            count = 0
            for detail in escrow_details:
                income = detail.get("order_income", {})
                val = income.get(key, 0)
                if val != 0:
                    total += val
                    count += 1
            fee_totals[key] = total
            fee_counts[key] = count

        # Write fee rows
        for i, (key, label) in enumerate(FEE_CATEGORIES.items(), 1):
            r = row + i
            ws.cell(row=r, column=1, value=label)
            ws.cell(row=r, column=2, value=fee_totals[key])
            ws.cell(row=r, column=2).number_format = CURRENCY_FORMAT
            ws.cell(row=r, column=3, value=fee_counts[key])

            for col in range(1, 4):
                ws.cell(row=r, column=col).border = self.thin_border

        # Grand total
        grand_total_row = row + len(FEE_CATEGORIES) + 1
        ws.cell(row=grand_total_row, column=1, value="TOTAL POTONGAN")
        ws.cell(row=grand_total_row, column=1).font = Font(bold=True)
        ws.cell(row=grand_total_row, column=2, value=sum(fee_totals.values()))
        ws.cell(row=grand_total_row, column=2).number_format = CURRENCY_FORMAT
        ws.cell(row=grand_total_row, column=2).font = Font(bold=True)

        self._auto_width(ws)

    def save(self, filename: str, month: int, year: int) -> str:
        """Save workbook to file. Returns full path."""
        os.makedirs(OUTPUT_DIR, exist_ok=True)

        if not filename:
            filename = f"Laporan_Dana_Cair_Shopee_{_month_name(month)}_{year}.xlsx"

        filepath = os.path.join(OUTPUT_DIR, filename)
        self.wb.save(filepath)
        print(f"\n[EXPORT] Report saved: {os.path.abspath(filepath)}")
        return os.path.abspath(filepath)

    def generate_full_report(self, report_data: dict, month: int, year: int, filename: Optional[str] = None) -> str:
        """Generate complete multi-sheet XLSX report."""
        summary = report_data["summary"]
        orders = report_data["orders"]
        escrow_details = report_data["escrow_details"]

        print("\n[EXPORT] Generating XLSX report...")

        # Sheet 1: Summary
        self.create_summary_sheet(summary, month, year)

        # Sheet 2: Orders with escrow overview
        self.create_transactions_sheet(orders, escrow_details)

        # Sheet 3: Escrow detail per order (fee breakdown)
        self.create_escrow_detail_sheet(escrow_details)

        # Sheet 4: Item-level detail
        self.create_item_detail_sheet(escrow_details)

        # Sheet 5: Fee breakdown summary
        self.create_fee_breakdown_sheet(escrow_details)

        # Save
        return self.save(filename, month, year)


def _month_name(month: int) -> str:
    """Get Indonesian month name."""
    months = {
        1: "Januari", 2: "Februari", 3: "Maret", 4: "April",
        5: "Mei", 6: "Juni", 7: "Juli", 8: "Agustus",
        9: "September", 10: "Oktober", 11: "November", 12: "Desember"
    }
    return months.get(month, str(month))
