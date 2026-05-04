# Finance Reports - Shopee Dana Cair

Script untuk menarik data dana yang sudah cair dari Shopee dan menghasilkan laporan XLSX yang rinci.

## Quick Start

```bash
cd scripts/finance-reports
pip install -r requirements.txt
python main.py
```

## Output

File XLSX dengan 5 sheet:

| Sheet | Isi |
|-------|-----|
| **Ringkasan** | Summary: total dana cair, potongan, net income |
| **Transaksi Dana Cair** | Semua transaksi yang sudah cair (per order) |
| **Detail Escrow per Order** | Breakdown biaya per order (komisi, service fee, ongkir, dll) |
| **Detail Produk per Item** | Product ID, SKU, nama produk, qty, harga, diskon per item |
| **Rekap Potongan** | Total semua jenis potongan/biaya |

## Options

```bash
python main.py --month 4 --year 2026              # April 2026 (default)
python main.py --month 3 --year 2026              # Maret 2026
python main.py --output "Laporan_Custom.xlsx"     # Custom filename
python main.py --base-url http://localhost:3000    # Local development
```

## File Structure

```
finance-reports/
├── main.py              # Entry point - run this
├── config.py            # Configuration (URL, credentials, constants)
├── api_client.py        # API client with auth management
├── shopee_finance.py    # Shopee data fetcher (wallet + escrow)
├── xlsx_exporter.py     # XLSX report generator
├── requirements.txt     # Python dependencies
├── README.md            # This file
└── output/              # Generated reports (auto-created)
```

## Data Flow

```
Login (yumna/password123)
    → GET /api/shopee/wallet/report (month=4, year=2026)
        → Returns: all wallet transactions for April
    → Filter: ESCROW_RELEASED + SETTLEMENT only (= dana cair)
    → Extract unique order_sn list
    → POST /api/shopee/wallet/escrow-detail-batch
        → Returns: per-order fee breakdown + per-item detail
    → Generate XLSX with all data
```

## Reuse

Script ini bisa di-reuse untuk bulan/tahun lain. Cukup ubah parameter `--month` dan `--year`.
