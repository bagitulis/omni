"""
XLSX Report Exporter - Shopee Dana Cair.
7 Sheets: Ringkasan, Daftar Pesanan, Detail Produk, Ringkasan Produk,
          Rekap Potongan, Analisis Margin, Data Produk (Referensi)
Sheets:
1. Ringkasan           - Executive summary
2. Daftar Pesanan      - Per order: escrow, fee breakdown
3. Detail Produk       - Per item: product ID, SKU, qty, harga
4. Ringkasan Produk    - Total qty terjual per produk/SKU
5. Rekap Potongan      - Fee summary with percentages
"""

import os
import sys
from collections import defaultdict
from datetime import datetime
from typing import Optional

try:
    from openpyxl import Workbook
    from openpyxl.styles import Font, PatternFill, Alignment, Border, Side
    from openpyxl.utils import get_column_letter
except ImportError:
    import subprocess
    subprocess.check_call([sys.executable, "-m", "pip", "install", "openpyxl", "-q"])
    from openpyxl import Workbook
    from openpyxl.styles import Font, PatternFill, Alignment, Border, Side
    from openpyxl.utils import get_column_letter

from config import HEADER_FILL_COLOR, HEADER_FONT_COLOR, CURRENCY_FORMAT, OUTPUT_DIR


class XLSXExporter:

    def __init__(self):
        self.wb = Workbook()
        self.wb.remove(self.wb.active)
        self.hdr_font = Font(name="Calibri", size=10, bold=True, color=HEADER_FONT_COLOR)
        self.hdr_fill = PatternFill(start_color=HEADER_FILL_COLOR, end_color=HEADER_FILL_COLOR, fill_type="solid")
        self.hdr_align = Alignment(horizontal="center", vertical="center", wrap_text=True)
        self.bold = Font(name="Calibri", size=10, bold=True)
        self.title = Font(name="Calibri", size=14, bold=True)
        self.sub = Font(name="Calibri", size=10, italic=True, color="555555")
        self.border = Border(left=Side("thin"), right=Side("thin"), top=Side("thin"), bottom=Side("thin"))

    def _hdr(self, ws, row, count):
        for c in range(1, count + 1):
            cell = ws.cell(row=row, column=c)
            cell.font, cell.fill, cell.alignment, cell.border = self.hdr_font, self.hdr_fill, self.hdr_align, self.border

    def _auto(self, ws):
        for col in ws.columns:
            mx = max((min(len(str(c.value or "")), 40) for c in col), default=8)
            ws.column_dimensions[get_column_letter(col[0].column)].width = mx + 3

    def _row(self, ws, r, data, money_cols=None):
        for c, v in enumerate(data, 1):
            cell = ws.cell(row=r, column=c, value=v)
            cell.border = self.border
            if money_cols and c in money_cols:
                cell.number_format = CURRENCY_FORMAT

    # -- Sheet 1: Ringkasan --

    def _sheet_summary(self, s, month, year):
        ws = self.wb.create_sheet("Ringkasan")
        ws.cell(row=1, column=1, value="LAPORAN DANA CAIR - SHOPEE").font = self.title
        ws.merge_cells("A1:C1")
        ws.cell(row=2, column=1, value=f"Periode: {_bulan(month)} {year}").font = self.sub
        ws.cell(row=3, column=1, value=f"Dibuat: {datetime.now().strftime('%d/%m/%Y %H:%M')}").font = self.sub

        r = 5
        for c, h in enumerate(["Keterangan", "Jumlah", "Nominal (Rp)"], 1):
            ws.cell(row=r, column=c, value=h)
        self._hdr(ws, r, 3)

        rows = [
            ("Pesanan Selesai", s["order_count"], ""),
            ("Total Qty Terjual", f"{s['total_qty_sold']} pcs", ""),
            ("", "", ""),
            ("Dibayar Pembeli", "", s["total_buyer_amount"]),
            ("(-) Komisi Platform", "", -s["total_commission"]),
            ("(-) Biaya Layanan", "", -s["total_service_fee"]),
            ("(-) Biaya Proses", "", -s["total_processing_fee"]),
            ("Ongkir Aktual", "", s["total_shipping"]),
            ("(+) Subsidi Ongkir", "", s["total_shipping_rebate"]),
            ("", "", ""),
            ("DANA CAIR (Escrow)", s["order_count"], s["total_escrow"]),
        ]
        for i, (desc, qty, nom) in enumerate(rows, 1):
            rr = r + i
            ws.cell(row=rr, column=1, value=desc)
            ws.cell(row=rr, column=2, value=qty)
            ws.cell(row=rr, column=3, value=nom)
            if isinstance(nom, (int, float)) and nom != "":
                ws.cell(row=rr, column=3).number_format = CURRENCY_FORMAT
            for c in range(1, 4):
                ws.cell(row=rr, column=c).border = self.border
            if "DANA CAIR" in str(desc):
                for c in range(1, 4):
                    ws.cell(row=rr, column=c).font = self.bold
        self._auto(ws)

    # -- Sheet 2: Daftar Pesanan --

    def _sheet_orders(self, orders):
        ws = self.wb.create_sheet("Daftar Pesanan")
        hdrs = ["No", "Order SN", "Tanggal", "Buyer",
                "Bayar Pembeli (Rp)", "Dana Cair (Rp)",
                "Komisi (Rp)", "Service Fee (Rp)", "Biaya Proses (Rp)",
                "Ongkir (Rp)", "Subsidi Ongkir (Rp)"]
        for c, h in enumerate(hdrs, 1):
            ws.cell(row=1, column=c, value=h)
        self._hdr(ws, 1, len(hdrs))
        money = {5, 6, 7, 8, 9, 10, 11}

        for i, o in enumerate(orders, 1):
            self._row(ws, i + 1, [
                i, o["order_sn"], o["order_date"], o["buyer_username"],
                o["buyer_total_amount"], o["escrow_amount"],
                o["commission_fee"], o["service_fee"], o["seller_processing_fee"],
                o["actual_shipping_fee"], o["shopee_shipping_rebate"],
            ], money)

        if orders:
            tr = len(orders) + 2
            ws.cell(row=tr, column=3, value="TOTAL").font = self.bold
            for ci, k in [(5,"buyer_total_amount"),(6,"escrow_amount"),(7,"commission_fee"),
                          (8,"service_fee"),(9,"seller_processing_fee"),(10,"actual_shipping_fee"),
                          (11,"shopee_shipping_rebate")]:
                c = ws.cell(row=tr, column=ci, value=sum(o[k] for o in orders))
                c.number_format = CURRENCY_FORMAT
                c.font = self.bold
        self._auto(ws)

    # -- Sheet 3: Detail Produk --

    def _sheet_items(self, items):
        ws = self.wb.create_sheet("Detail Produk")
        hdrs = ["No", "Order SN", "Product ID", "SKU",
                "Nama Produk", "Varian", "Qty",
                "Harga Asli (Rp)", "Harga Jual (Rp)", "Harga Diskon (Rp)",
                "Diskon Seller (Rp)", "Diskon Shopee (Rp)"]
        for c, h in enumerate(hdrs, 1):
            ws.cell(row=1, column=c, value=h)
        self._hdr(ws, 1, len(hdrs))
        money = {8, 9, 10, 11, 12}

        for i, it in enumerate(items, 1):
            self._row(ws, i + 1, [
                i, it["order_sn"], it["item_id"], it["model_sku"] or it["sku"],
                it["item_name"], it["model_name"], it["quantity"],
                it["original_price"], it["selling_price"], it["discounted_price"],
                it["seller_discount"], it["shopee_discount"],
            ], money)

        if items:
            tr = len(items) + 2
            ws.cell(row=tr, column=6, value="TOTAL").font = self.bold
            ws.cell(row=tr, column=7, value=sum(i["quantity"] for i in items)).font = self.bold
            for ci, k in [(8,"original_price"),(10,"discounted_price")]:
                c = ws.cell(row=tr, column=ci, value=sum(i[k] for i in items))
                c.number_format = CURRENCY_FORMAT
                c.font = self.bold
        self._auto(ws)

    # -- Sheet 4: Ringkasan Produk --

    def _sheet_product_summary(self, items):
        ws = self.wb.create_sheet("Ringkasan Produk")
        ws.cell(row=1, column=1, value="RINGKASAN PENJUALAN PER PRODUK").font = self.title
        ws.merge_cells("A1:F1")

        hdrs = ["No", "Product ID", "SKU", "Nama Produk", "Varian",
                "Total Qty", "Total Penjualan (Rp)", "Jumlah Transaksi"]
        r = 3
        for c, h in enumerate(hdrs, 1):
            ws.cell(row=r, column=c, value=h)
        self._hdr(ws, r, len(hdrs))

        # Aggregate by SKU
        agg = defaultdict(lambda: {
            "item_id": "", "sku": "", "item_name": "", "model_name": "",
            "qty": 0, "revenue": 0.0, "tx_count": 0,
        })
        for it in items:
            key = it["model_sku"] or it["sku"] or it["item_id"]
            rec = agg[key]
            rec["item_id"] = it["item_id"]
            rec["sku"] = it["model_sku"] or it["sku"]
            rec["item_name"] = it["item_name"]
            rec["model_name"] = it["model_name"]
            rec["qty"] += it["quantity"]
            rec["revenue"] += it["discounted_price"]
            rec["tx_count"] += 1

        sorted_products = sorted(agg.values(), key=lambda x: x["qty"], reverse=True)

        for i, p in enumerate(sorted_products, 1):
            self._row(ws, r + i, [
                i, p["item_id"], p["sku"], p["item_name"], p["model_name"],
                p["qty"], p["revenue"], p["tx_count"],
            ], {7})

        if sorted_products:
            tr = r + len(sorted_products) + 1
            ws.cell(row=tr, column=5, value="TOTAL").font = self.bold
            ws.cell(row=tr, column=6, value=sum(p["qty"] for p in sorted_products)).font = self.bold
            c = ws.cell(row=tr, column=7, value=sum(p["revenue"] for p in sorted_products))
            c.number_format = CURRENCY_FORMAT
            c.font = self.bold
            ws.cell(row=tr, column=8, value=sum(p["tx_count"] for p in sorted_products)).font = self.bold
        self._auto(ws)

    # -- Sheet 5: Rekap Potongan --

    def _sheet_deductions(self, s):
        ws = self.wb.create_sheet("Rekap Potongan")
        ws.cell(row=1, column=1, value="REKAP POTONGAN & BIAYA").font = self.title
        ws.merge_cells("A1:D1")

        r = 3
        hdrs = ["Jenis Potongan", "Total (Rp)", "Rata-rata/Order (Rp)", "% dari Pembeli"]
        for c, h in enumerate(hdrs, 1):
            ws.cell(row=r, column=c, value=h)
        self._hdr(ws, r, 4)

        buyer = s["total_buyer_amount"] or 1
        cnt = s["order_count"] or 1

        fees = [
            ("Komisi Platform", s["total_commission"]),
            ("Biaya Layanan", s["total_service_fee"]),
            ("Biaya Proses Pesanan", s["total_processing_fee"]),
            ("Ongkir Aktual", s["total_shipping"]),
            ("(-) Subsidi Ongkir Shopee", -s["total_shipping_rebate"]),
        ]
        for i, (label, amt) in enumerate(fees, 1):
            rr = r + i
            ws.cell(row=rr, column=1, value=label).border = self.border
            c = ws.cell(row=rr, column=2, value=amt)
            c.number_format = CURRENCY_FORMAT
            c.border = self.border
            c = ws.cell(row=rr, column=3, value=round(amt / cnt))
            c.number_format = CURRENCY_FORMAT
            c.border = self.border
            ws.cell(row=rr, column=4, value=f"{(amt/buyer)*100:.1f}%").border = self.border

        total_fee = s["total_commission"] + s["total_service_fee"] + s["total_processing_fee"]
        net_ship = s["total_shipping"] - s["total_shipping_rebate"]

        tr = r + len(fees) + 2
        ws.cell(row=tr, column=1, value="Total Potongan (excl ongkir)").font = self.bold
        c = ws.cell(row=tr, column=2, value=total_fee)
        c.number_format = CURRENCY_FORMAT
        c.font = self.bold
        ws.cell(row=tr, column=4, value=f"{(total_fee/buyer)*100:.1f}%").font = self.bold

        tr += 1
        ws.cell(row=tr, column=1, value="Ongkir Netto (Aktual - Subsidi)")
        ws.cell(row=tr, column=2, value=net_ship).number_format = CURRENCY_FORMAT

        tr += 2
        ws.cell(row=tr, column=1, value="Dana Cair = Bayar Pembeli - Potongan - Ongkir Netto").font = self.bold
        tr += 1
        ws.cell(row=tr, column=1, value=f"Rp {s['total_escrow']:,.0f} = Rp {buyer:,.0f} - Rp {total_fee:,.0f} - Rp {net_ship:,.0f}")
        self._auto(ws)

    # -- Sheet 6: Analisis Margin --

    def _sheet_margin(self, margin_rows, costs):
        ws = self.wb.create_sheet("Analisis Margin")
        ws.cell(row=1, column=1, value="ANALISIS MARGIN PER PRODUK").font = self.title
        ws.merge_cells("A1:N1")
        ws.cell(row=2, column=1, value="Alur: Harga Beli/pcs + Margin 3% = Harga Modal/pcs. Lalu dibandingkan dengan Dana Cair/pcs dari Shopee.").font = self.sub
        ws.cell(row=3, column=1, value="Biaya Proses Rp 1.250 per order (flat), dibagi rata ke qty. Makin banyak qty per order, makin kecil biaya proses/pcs.").font = self.sub

        hdrs = ["No", "SKU", "Nama Produk", "Varian", "Tier Qty",
                "Jml Order", "Total Qty",
                "Harga Listing/pcs", "Harga Beli/pcs", "Harga Modal/pcs (Beli+3%)",
                "Dana Cair/pcs", "Selisih vs Modal/pcs", "Status", "Keterangan Biaya"]
        r = 5
        for c, h in enumerate(hdrs, 1):
            ws.cell(row=r, column=c, value=h)
        self._hdr(ws, r, len(hdrs))
        money = {8, 9, 10, 11, 12}

        red_fill = PatternFill(start_color="FFC7CE", end_color="FFC7CE", fill_type="solid")
        red_font = Font(name="Calibri", size=10, bold=True, color="9C0006")
        green_font = Font(name="Calibri", size=10, color="006100")

        for i, mr in enumerate(margin_rows, 1):
            sku = mr["sku"]
            cost = costs.get(sku)
            beli = cost["beli_per_pcs"] if cost else 0
            modal = cost["modal_per_pcs"] if cost else 0  # beli + 3%
            selisih = mr["avg_cair"] - modal if modal > 0 else 0

            if modal == 0:
                status = "DATA MODAL TIDAK DITEMUKAN"
            elif selisih < 0:
                status = f"DI BAWAH TARGET (RUGI Rp {abs(selisih):,.0f})"
            else:
                status = f"DI ATAS TARGET +Rp {selisih:,.0f}"

            # Keterangan biaya platform
            ket = f"Komisi Rp {mr['avg_komisi']:,.0f} + Svc Rp {mr['avg_svc']:,.0f} + Proses Rp {mr['avg_proc']:,.0f}"

            self._row(ws, r + i, [
                i, sku or "(tanpa SKU)", mr["item_name"], mr["model_name"], mr["tier"],
                mr["jml_order"], mr["total_qty"],
                mr["avg_listing"], beli, modal, mr["avg_cair"], selisih, status, ket,
            ], money)

            # Highlight
            if modal > 0 and selisih < 0:
                for c in range(1, len(hdrs) + 1):
                    ws.cell(row=r + i, column=c).fill = red_fill
                    ws.cell(row=r + i, column=c).font = red_font
            elif modal > 0:
                ws.cell(row=r + i, column=13).font = green_font

        self._auto(ws)

    # -- Sheet 7: Data Produk (Referensi) --

    def _sheet_product_ref(self, costs):
        ws = self.wb.create_sheet("Data Produk (Referensi)")
        ws.cell(row=1, column=1, value="CARA HITUNG HARGA MODAL PER PRODUK").font = self.title
        ws.merge_cells("A1:H1")
        ws.cell(row=2, column=1, value="Sumber: Google Sheet 'ALL PRODUCT'. Hanya produk yang terjual di periode ini.").font = self.sub
        ws.cell(row=3, column=1, value="Rumus: Harga Beli/Karton (in PPN) / Pcs per Karton = Harga Beli/pcs. Lalu + Margin 3% = Harga Modal/pcs.").font = self.sub

        hdrs = ["No", "SKU", "Nama Barang", "Pcs/Karton",
                "Harga Beli/Karton in PPN (Rp)", "Harga Beli/pcs (Rp)",
                "Harga Modal/Karton +3% (Rp)", "Harga Modal/pcs +3% (Rp)"]
        r = 5
        for c, h in enumerate(hdrs, 1):
            ws.cell(row=r, column=c, value=h)
        self._hdr(ws, r, len(hdrs))
        money = {5, 6, 7, 8}

        sorted_costs = sorted(costs.items(), key=lambda x: x[0])
        for i, (sku, c) in enumerate(sorted_costs, 1):
            self._row(ws, r + i, [
                i, sku, c["nama_sheet"], c["pcs_per_karton"],
                c["beli_per_karton"], c["beli_per_pcs"],
                c["modal_per_karton"], c["modal_per_pcs"],
            ], money)

        self._auto(ws)

    # -- Generate --

    def generate_full_report(self, data, month, year, filename=None):
        s = data["summary"]
        print("\n[EXPORT] Generating XLSX...")
        self._sheet_summary(s, month, year)
        self._sheet_orders(data["orders"])
        self._sheet_items(data["items"])
        self._sheet_product_summary(data["items"])
        self._sheet_deductions(s)
        self._sheet_margin(data["margin_rows"], data["costs"])
        self._sheet_product_ref(data["costs"])

        os.makedirs(OUTPUT_DIR, exist_ok=True)
        fn = filename or f"Laporan_Dana_Cair_Shopee_{_bulan(month)}_{year}.xlsx"
        path = os.path.join(OUTPUT_DIR, fn)
        self.wb.save(path)
        full = os.path.abspath(path)
        print(f"[EXPORT] Saved: {full}")
        return full


def _bulan(m):
    return {1:"Januari",2:"Februari",3:"Maret",4:"April",5:"Mei",6:"Juni",
            7:"Juli",8:"Agustus",9:"September",10:"Oktober",11:"November",12:"Desember"}.get(m, str(m))
