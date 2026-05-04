# Finance Reports - Shopee Dana Cair

Script untuk menarik data dana yang sudah cair dari Shopee dan menghasilkan laporan XLSX lengkap dengan analisis margin per produk.

## Quick Start

```bash
cd scripts/finance-reports
pip install -r requirements.txt
python main.py                          # Default: April 2026
python main.py --month 5 --year 2026    # Bulan lain
```

## Prasyarat

- Docker running (untuk akses PostgreSQL via `docker exec`)
- Backend OMNI running (untuk trigger sync)
- Google Sheets service account (`bertigahemat-f1bd6932b229.json`) di `backend/config/static/google/`
- Data escrow di-sync via web UI (`/report/shopee` > Sync Escrow) atau script trigger otomatis

## Output XLSX (7 Sheet)

| # | Sheet | Isi |
|---|-------|-----|
| 1 | **Ringkasan** | Total pesanan, qty terjual, bayar pembeli, potongan, dana cair |
| 2 | **Daftar Pesanan** | Per order: Order SN, tanggal, buyer, dana cair, komisi, service fee, biaya proses, ongkir |
| 3 | **Detail Produk** | Per item: Product ID, SKU, nama produk, varian, qty, harga asli/jual/diskon |
| 4 | **Ringkasan Produk** | Per SKU: total qty terjual, total penjualan, jumlah transaksi (sorted bestseller) |
| 5 | **Rekap Potongan** | Rincian biaya platform: komisi, service fee, biaya proses, ongkir, persentase |
| 6 | **Analisis Margin** | Per SKU per tier qty: Harga Beli > Modal+3% > Dana Cair > Selisih (DI ATAS/BAWAH TARGET) |
| 7 | **Data Produk (Referensi)** | Import dari Sheet ALL PRODUCT: cara hitung harga modal dari kartonan ke per pcs |

## Alur Perhitungan Harga

```
DARI SHEET "ALL PRODUCT" (Google Sheets):
  Harga Beli/Karton (in PPN)  /  Pcs per Karton  =  Harga Beli/pcs
  Harga Beli/pcs  +  Margin 3%  =  HARGA MODAL/pcs (target minimum profit)

DI SHOPEE (dari data escrow):
  Harga Listing/pcs (yang dipasang seller)
    - Komisi Platform (8-9%)
    - Service Fee (4-5%)
    - Biaya Proses (Rp 1.250 flat per order, dibagi rata ke qty)
    = DANA CAIR/pcs (yang masuk ke wallet seller)

PERBANDINGAN:
  Dana Cair/pcs  -  Harga Modal/pcs  =  SELISIH
  Selisih > 0  =  DI ATAS TARGET (untung di atas margin 3%)
  Selisih < 0  =  DI BAWAH TARGET (RUGI, cair di bawah modal)
```

## Catatan Penting

- **Tier Qty**: Biaya Proses Rp 1.250/order (flat). 1 pcs = Rp 1.250/pcs. 6 pcs = Rp 208/pcs.
- **Analisis Margin** hanya dari single-item order (1 jenis produk per order) supaya akurat.
- **Harga Listing** bisa beda dari Sheet karena seller set harga sendiri di Shopee.
- **Produk tanpa SKU**: Jika `model_sku` kosong, script fallback ke field `sku`.

## Data Flow

```
1. Login OMNI backend
2. POST /api/analytics/shopee/sync {month, year}
   -> Backend call Shopee API untuk semua order COMPLETED
   -> Simpan ke DB: shopee_escrow_orders + shopee_escrow_items
3. Query DB via Docker (psql)
   -> Orders: dana cair, komisi, service fee, ongkir per order
   -> Items: product ID, SKU, qty, harga per item
   -> Margin: avg cair/pcs per SKU per tier qty
4. Baca Google Sheets "ALL PRODUCT"
   -> Harga beli/karton, pcs/karton, harga modal+3%
5. Generate XLSX 7 sheet
```

## File Structure

```
finance-reports/
  main.py              # Entry point
  config.py            # Konfigurasi
  api_client.py        # API client + auth
  shopee_finance.py    # Data fetcher (DB + Sheets)
  xlsx_exporter.py     # XLSX generator
  requirements.txt     # Dependencies
  README.md            # Dokumentasi ini
  output/              # Hasil report
```

## Reuse

```bash
python main.py --month 3 --year 2026    # Maret
python main.py --month 4 --year 2026    # April
python main.py --month 5 --year 2026    # Mei
```
