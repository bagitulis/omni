# 📊 TikTok Ads Analysis - Notebooks

## 🚀 Quick Start

### Via Copilot (Recommended)

Cukup ketik ke Copilot:

```
"Generate laporan TikTok Ads"
```

Copilot akan otomatis:

1. Membaca semua data
2. Menjalankan analisis ML
3. Generate laporan HTML
4. Membuka laporan di browser

### Via Command Line

```bash
cd notebooks/tiktok_ads/scripts
python generate_report.py
```

**Output:** `tiktok_ads/output/TIKTOK_ADS_REPORT_[timestamp].html`

---

## 📁 Struktur Folder (v3.0)

```
notebooks/
├── core/                      # 🆕 Generic library (platform-agnostic)
│   ├── statistics/            # Statistical methods
│   │   ├── trend.py           # Mann-Kendall, linear regression
│   │   ├── momentum.py        # Momentum analysis
│   │   ├── variation.py       # CV analysis
│   │   ├── confidence.py      # Confidence intervals
│   │   └── correlation.py     # Pearson, Spearman
│   ├── intelligence/          # ML/AI analysis
│   │   ├── scoring.py         # Composite scorer
│   │   ├── elasticity.py      # Budget elasticity
│   │   ├── churn_risk.py      # Churn risk calculator
│   │   ├── unified_scorer.py  # Unified scoring engine
│   │   └── ...
│   ├── calendar/              # Calendar utilities
│   │   └── indonesian.py      # Indonesian events (payday, holidays)
│   └── reports/               # Report templates
│       └── base.py            # Base report generator
│
├── tiktok_ads/                # TikTok-specific implementation
│   ├── config.py              # TikTok thresholds & config
│   ├── stats.py               # Wrapper → core/statistics
│   ├── correlation.py         # Wrapper → core/intelligence
│   ├── confidence.py          # Wrapper → core/statistics
│   ├── elasticity.py          # Wrapper → core/intelligence
│   ├── analysis.py            # Product analysis
│   ├── data.py                # Data loading
│   ├── reports/               # 🆕 Report modules
│   │   ├── report.py          # Main report generator
│   │   ├── report_executive.py
│   │   └── report_html.py
│   ├── scripts/               # 🆕 Executable scripts
│   │   ├── generate_report.py # Main script
│   │   └── exec_*.py          # Helper scripts
│   ├── intelligence/          # ML analysis engine
│   │   ├── unified_scorer.py  # TikTok scorer (wrapper)
│   │   ├── comprehensive_engine.py
│   │   └── reports/           # Intelligence reports
│   └── output/                # Generated reports
│
└── README.md                  # Dokumentasi ini
```

### Arsitektur: Core + Wrapper Pattern

```
┌─────────────────────────────────────────────────────────┐
│  tiktok_ads/stats.py (Wrapper)                          │
│  - Import dari core/statistics/trend.py                 │
│  - Tambah TikTok-specific emoji & thresholds            │
└────────────────────┬────────────────────────────────────┘
                     │ import
                     ▼
┌─────────────────────────────────────────────────────────┐
│  core/statistics/trend.py (Generic)                     │
│  - mann_kendall_test()                                  │
│  - linear_trend_analysis()                              │
│  - Platform-agnostic, reusable                          │
└─────────────────────────────────────────────────────────┘
```

---

## 🆕 Cara Menambah Fitur Baru

### 1. Menambah Report Mingguan

```python
# File: tiktok_ads/scripts/generate_weekly_report.py

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from tiktok_ads.intelligence import ComprehensiveAnalysisEngine
from tiktok_ads.reports.report_html import generate_html
from datetime import datetime, timedelta

def generate_weekly_report():
    """Generate report untuk 7 hari terakhir."""
    engine = ComprehensiveAnalysisEngine()

    # Filter data 7 hari terakhir
    end_date = datetime.now()
    start_date = end_date - timedelta(days=7)

    analyses = engine.runAnalysis(
        minCost=10000,
        startDate=start_date,
        endDate=end_date
    )

    # Generate report
    html = generate_html(analyses, title=f"Weekly Report {start_date:%d %b} - {end_date:%d %b %Y}")

    output_path = Path(__file__).parent.parent / "output" / f"WEEKLY_REPORT_{end_date:%Y%m%d}.html"
    output_path.write_text(html, encoding='utf-8')
    print(f"✅ Weekly report saved: {output_path}")

if __name__ == "__main__":
    generate_weekly_report()
```

### 2. Sinkronisasi Nama Produk

```python
# File: tiktok_ads/scripts/sync_product_names.py

import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

import pandas as pd
from tiktok_ads.config import DB_PATH, EXCEL_PRODUCT_REF

def sync_product_names():
    """Sinkronisasi nama produk dari Excel ke database."""
    import sqlite3

    # Load product names dari Excel
    df_names = pd.read_excel(EXCEL_PRODUCT_REF)

    # Connect ke database
    conn = sqlite3.connect(DB_PATH)

    # Update nama produk
    for _, row in df_names.iterrows():
        product_id = row['product_id']
        product_name = row['product_name']

        conn.execute("""
            UPDATE TiktokAdsPerformance
            SET productName = ?
            WHERE productId = ?
        """, (product_name, product_id))

    conn.commit()
    conn.close()
    print(f"✅ Synced {len(df_names)} product names")

if __name__ == "__main__":
    sync_product_names()
```

### 3. Menambah Platform Baru (Shopee/Lazada)

Karena logic generic sudah di `core/`, tinggal buat wrapper:

```python
# File: shopee_ads/__init__.py

from core.statistics import mann_kendall_test, calculate_momentum
from core.intelligence import UnifiedScorer, ChurnRiskCalculator

# Shopee-specific thresholds
SHOPEE_ROI_GOOD = 3.0  # Shopee margin lebih tinggi
SHOPEE_ROI_BAD = 1.5

class ShopeeScorer(UnifiedScorer):
    """Shopee-specific scorer dengan threshold berbeda."""

    def __init__(self):
        super().__init__(
            roiThresholdGood=SHOPEE_ROI_GOOD,
            roiThresholdBad=SHOPEE_ROI_BAD
        )
```

---

## 🔄 Rolling Window (Quarterly Analysis)

**FITUR BARU:** Kode otomatis menyesuaikan dengan data baru!

### Cara Kerja

- Jika data terakhir Januari 2026 → Analisis Nov 2025 - Jan 2026
- Jika data terakhir April 2026 → Analisis Feb - Apr 2026
- **TIDAK PERLU ubah kode apapun!**

### Contoh Penggunaan

```python
from tiktok_ads_core import run_full_analysis, run_quarterly_analysis

# Analisis semua data
result = run_full_analysis()

# Analisis hanya 3 bulan terakhir (auto-adjust)
result = run_quarterly_analysis()

# Custom window (misalnya 6 bulan)
result = run_full_analysis(use_rolling_window=True, window_months=6)
```

---

## 📊 Data Quality Labeling

**FITUR BARU:** Produk dengan data terbatas akan diberi label!

### Quality Levels

| Level           | Periode | Confidence | Keterangan                   |
| --------------- | ------- | ---------- | ---------------------------- |
| ✅ HIGH         | ≥8      | 90%+       | Analisis valid dan reliable  |
| 🔶 MEDIUM       | 4-7     | 60%+       | Cukup valid, perlu hati-hati |
| ⚠️ LOW          | 2-3     | 30%        | Rekomendasi tentatif         |
| ❌ INSUFFICIENT | 1       | 0%         | Data tidak cukup             |

### Contoh Output

```
- Produk XYZ: Score 75, 🟢 LANJUTKAN
  Data Quality: ⚠️ LOW (n=3)
  Action: [DATA TERBATAS] Scale up budget 30-50%
```

**⚠️ PENTING:** Produk dengan label `[DATA TERBATAS]` memerlukan verifikasi manual sebelum mengambil keputusan besar.

---

## 📋 MCP Tools (untuk Copilot)

| Tool                | Fungsi                                 |
| ------------------- | -------------------------------------- |
| `analyze_ads`       | Statistik ringkas (cost, revenue, ROI) |
| `get_top_products`  | Produk terbaik untuk scale up          |
| `get_stop_products` | Produk yang harus dihentikan           |
| `generate_insights` | Insight lengkap dengan rekomendasi     |
| `generate_report`   | Generate laporan HTML lengkap          |

### Contoh Prompt ke Copilot

- "Analisis performa TikTok Ads saya"
- "Produk mana yang harus di-scale up?"
- "Produk mana yang harus dihentikan?"
- "Generate laporan TikTok Ads"

---

## 📈 Metodologi Analisis

Sistem menggunakan pendekatan **Bertingkat (Layered Approach)**:

1. **Fundamental Layer:** Menentukan Status Produk (Lanjut/Pantau/Stop) menggunakan Composite Score.
2. **Optimization Layer:** Menentukan Strategi Budgeting (Aggressive/Cautious) menggunakan Elasticity & mROI.

### 1. Metode Statistik Dasar (Fundamental)

| Metode                   | Tujuan               | Min Data  |
| ------------------------ | -------------------- | --------- |
| Mann-Kendall Test        | Deteksi trend        | 4 periode |
| Linear Regression        | Slope & R²           | 2 periode |
| Momentum Analysis        | Recent vs Historical | 3 periode |
| Coefficient of Variation | Konsistensi          | 2 periode |

### 2. Metode Lanjutan (Optimization) 🆕

Metode ini digunakan untuk menjawab: _"Apakah budget harus dinaikkan 10% atau 30%?"_

| Metode                  | Tujuan               | Logika                                                                                                                                      |
| :---------------------- | :------------------- | :------------------------------------------------------------------------------------------------------------------------------------------ |
| **Ad Spend Elasticity** | Responsivitas Budget | Mengukur % kenaikan Revenue per 1% kenaikan Budget. <br>• `E > 1.0`: Elastic (Bagus/Stage 2) <br>• `E < 1.0`: Inelastic (Normal/Stage 1)    |
| **Marginal ROI (mROI)** | Efisiensi Terakhir   | Menghitung profitabilitas dari _penambahan budget_ terakhir. <br>• `mROI < 0`: **BLOCKER** (Jangan naikkan budget jika test terakhir rugi). |

### Composite Score (100 poin)

| Komponen          | Bobot | Formula               |
| ----------------- | ----- | --------------------- |
| ROI Score         | 30%   | min(ROI × 10, 100)    |
| Profit Score      | 25%   | 100 jika profit > 0   |
| Momentum Score    | 20%   | 50 + (change% / 2)    |
| Consistency Score | 15%   | 100 - CV%             |
| Trend Score       | 10%   | Based on Mann-Kendall |

## 🧠 Strategi Keputusan & Budget

### 1. Kategori Dasar (Status)

| Kategori     | Kriteria                                 |
| ------------ | ---------------------------------------- |
| 🟢 LANJUTKAN | Score ≥ 60 AND ROI ≥ 2x AND Profit > 0   |
| 🔴 HENTIKAN  | Score < 40 OR ROI < 1x OR Rugi > Rp 100k |
| 🟡 PANTAU    | Sisanya                                  |

### 2. Strategi Scaling 2-Tahap (Smart Budgeting) 🆕

Hanya berlaku untuk kategori **🟢 LANJUTKAN**.

| Strategi                 | Kenaikan | Syarat Kondisi                              | Insight                                                                                    |
| :----------------------- | :------- | :------------------------------------------ | :----------------------------------------------------------------------------------------- |
| **Stage 2 (Aggressive)** | **+30%** | ROI Tinggi (>2x) **DAN** Elastisitas > 1.0  | Market masih "lapar". Setiap Rp 1 budget ekstra menghasilkan >Rp 1 revenue ekstra.         |
| **Stage 1 (Cautious)**   | **+10%** | ROI Tinggi (>2x) **TAPI** Elastisitas < 1.0 | Market mulai jenuh (diminishing returns). Naikkan pelan-pelan.                             |
| **HOLD (Risk)**          | **0%**   | ROI Tinggi **TAPI** mROI < 0                | **Safety Brake:** Penambahan budget periode lalu terbukti merugi. Jangan ulangi kesalahan. |

---

## 🗓️ Periode Methodology (4 Bins per Month)

Pembagian berdasarkan perilaku belanja konsumen:

| Bin          | Tanggal | Karakteristik                      |
| ------------ | ------- | ---------------------------------- |
| Early Month  | 1-7     | Baru gajian, purchase power tinggi |
| Mid Month I  | 8-15    | Uang mulai menipis                 |
| Mid Month II | 16-23   | Purchase power rendah              |
| Late Month   | 24-31   | Akhir bulan, nunggu gajian         |

---

## 📥 Upload Data Baru

1. Letakkan file Excel di folder `ads_data/`
2. Format nama: `CreativePerformance_YYYYMMDD.xlsx`
3. Jalankan:

```bash
python upload_ads_data.py
```

Kode akan otomatis:

- Detect file baru
- Parse tanggal dari nama file
- Assign period label (4 bins/month)
- Upload ke database

---

## 🔧 Troubleshooting

### Error: "Module not found"

```bash
pip install pandas numpy scipy python-dateutil
```

### Error: "Database not found"

Pastikan path database benar di `tiktok_ads_core.py`:

```python
DB_PATH = Path(__file__).parent / "../backend/config/databases/yumna_bertigamart.db"
```

### Data tidak muncul

1. Cek file Excel ada di `ads_data/`
2. Jalankan `python upload_ads_data.py`
3. Pastikan format kolom sesuai (Indonesian atau English)

---

_Last updated: January 2026_
_Using: Mann-Kendall Test, Linear Regression, Momentum Analysis, CV Analysis_
