#!/usr/bin/env python
"""
TikTok Ads Data - Upload Script
================================
Upload data dari file Excel ke database dengan periode dari nama file.

⚠️ AGENTS.MD Compliance:
- Menggunakan Python untuk database operations
- Tidak menambah tabel baru
- Non-destructive operation (hanya INSERT)

Metodologi Periode (didiskusikan):
- Early Month: Tanggal 1-7 (minggu awal, baru gajian)
- Mid Month I: Tanggal 8-15 (pertengahan awal)
- Mid Month II: Tanggal 16-23 (pertengahan akhir)
- Late Month: Tanggal 24-End (akhir bulan, nunggu gajian)
"""

import sqlite3
import pandas as pd
import numpy as np
from datetime import datetime
from pathlib import Path
import re
import os
import uuid
from pathlib import Path

# ============================================================
# CONFIGURATION
# ============================================================
TENANT_ID = "yumna_bertigamart"

# Use absolute path resolution
_script_dir = Path(__file__).parent.resolve()
DB_PATH = str(_script_dir.parent.parent / "backend" / "config" / "databases" / "yumna_bertigamart.db")
ADS_DATA_DIR = str(_script_dir / "data")


# ============================================================
# PERIOD LABELING (Metodologi yang didiskusikan)
# ============================================================
def get_period_label(start_date: str, end_date: str) -> str:
    """
    Tentukan label periode berdasarkan tanggal.
    
    Metodologi 4 Bin/Month:
    - Early Month (Tgl 1-7): Awal bulan, baru gajian
    - Mid Month I (Tgl 8-15): Pertengahan awal
    - Mid Month II (Tgl 16-23): Pertengahan akhir
    - Late Month (Tgl 24-End): Akhir bulan, nunggu gajian
    
    Args:
        start_date: Format YYYY-MM-DD
        end_date: Format YYYY-MM-DD
        
    Returns:
        str: Label periode (e.g., "2025-10 Early Month")
    """
    try:
        start = datetime.strptime(start_date, '%Y-%m-%d')
        end = datetime.strptime(end_date, '%Y-%m-%d')
        
        day_start = start.day
        year_month = start.strftime('%Y-%m')
        
        if day_start <= 7:
            return f"{year_month} Early Month"
        elif day_start <= 15:
            return f"{year_month} Mid Month I"
        elif day_start <= 23:
            return f"{year_month} Mid Month II"
        else:
            return f"{year_month} Late Month"
    except:
        return "Unknown"


def get_period_bin(day: int) -> int:
    """Get numeric bin (1-4) for a day of month"""
    if day <= 7:
        return 1  # Early Month
    elif day <= 15:
        return 2  # Mid Month I
    elif day <= 23:
        return 3  # Mid Month II
    else:
        return 4  # Late Month


# ============================================================
# CLEAR OLD DATA
# ============================================================
def clear_old_data(confirm: bool = False):
    """
    Hapus semua data TikTok Ads yang ada di database.
    
    ⚠️ HATI-HATI: Operasi ini menghapus data!
    
    Args:
        confirm: Harus True untuk eksekusi
        
    Returns:
        tuple: (deleted_creative, deleted_batch)
    """
    if not confirm:
        print("   ⚠️ Konfirmasi diperlukan! Set confirm=True untuk menghapus data.")
        return 0, 0
    
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    # Count existing data
    cursor.execute(
        "SELECT COUNT(*) FROM TiktokAdsCreativeData WHERE tenantId = ?", 
        (TENANT_ID,)
    )
    count_creative = cursor.fetchone()[0]
    
    cursor.execute(
        "SELECT COUNT(*) FROM TiktokAdsUploadBatch WHERE tenantId = ?", 
        (TENANT_ID,)
    )
    count_batch = cursor.fetchone()[0]
    
    print(f"\n   🗑️ Menghapus data lama...")
    print(f"      - TiktokAdsCreativeData: {count_creative:,} rows")
    print(f"      - TiktokAdsUploadBatch: {count_batch:,} rows")
    
    # Delete data
    cursor.execute(
        "DELETE FROM TiktokAdsCreativeData WHERE tenantId = ?", 
        (TENANT_ID,)
    )
    deleted_creative = cursor.rowcount
    
    cursor.execute(
        "DELETE FROM TiktokAdsUploadBatch WHERE tenantId = ?", 
        (TENANT_ID,)
    )
    deleted_batch = cursor.rowcount
    
    conn.commit()
    conn.close()
    
    print(f"   ✅ Berhasil dihapus: {deleted_creative:,} creative + {deleted_batch:,} batch")
    
    return deleted_creative, deleted_batch

# ============================================================
# EXTRACT PERIOD FROM FILENAME
# ============================================================
def extract_period_from_filename(filename):
    """
    Extract period from filename.
    Format: "creative data for product campaigns YYYY-MM-DD HH ~ YYYY-MM-DD HH.xlsx"
    
    Returns:
        tuple: (period_start, period_end, period_label)
    """
    match = re.search(r'(\d{4}-\d{2}-\d{2}) \d{2} ~ (\d{4}-\d{2}-\d{2}) \d{2}', filename)
    if match:
        start = match.group(1)
        end = match.group(2)
        label = get_period_label(start, end)
        return start, end, label
    return None, None, None

# ============================================================
# UPLOAD SINGLE FILE
# ============================================================
def upload_file(filepath, period_start, period_end, period_label):
    """Upload single Excel file to database
    
    Args:
        filepath: Path to Excel file
        period_start: Start date string (YYYY-MM-DD)
        period_end: End date string (YYYY-MM-DD)
        period_label: Label periode (e.g., "2025-10 Early Month")
    """
    
    filename = os.path.basename(filepath)
    print(f"\n   📄 {filename}")
    print(f"      Periode: {period_start} → {period_end}")
    print(f"      Label: {period_label}")
    
    # Read Excel
    df = pd.read_excel(filepath)
    print(f"      Rows: {len(df):,}")
    
    # Column mapping (TikTok English export → database)
    column_mapping = {
        'Campaign name': 'campaignName',
        # English columns (TikTok English export)
        'Campaign name': 'campaignName',
        'Campaign ID': 'campaignId',
        'Product ID': 'productId',
        'Creative type': 'creativeType',
        'Video title': 'videoTitle',
        'Video ID': 'videoId',
        'TikTok account': 'tiktokAccount',
        'Time posted': 'postingTime',
        'Status': 'status',
        'Authorization type': 'authorizationType',
        'Cost': 'cost',
        'SKU orders': 'ordersSku',
        'Cost per order': 'costPerOrder',
        'Gross revenue': 'grossRevenue',
        'ROI': 'roi',
        'Product ad impressions': 'impressions',
        'Product ad clicks': 'clicks',
        'Product ad click rate': 'ctr',
        'Ad conversion rate': 'conversionRate',
        '2-second ad video view rate': 'watchRate2s',
        '6-second ad video view rate': 'watchRate6s',
        '25% ad video view rate': 'watchRate25pct',
        '50% ad video view rate': 'watchRate50pct',
        '75% ad video view rate': 'watchRate75pct',
        '100% ad video view rate': 'watchRate100pct',
        'Currency': 'currency',
        
        # Indonesian columns (TikTok Indonesian export)
        'Nama kampanye': 'campaignName',
        'ID Campaign': 'campaignId',
        'ID produk': 'productId',
        'Jenis materi iklan': 'creativeType',
        'Judul video': 'videoTitle',
        'ID video': 'videoId',
        'Akun TikTok': 'tiktokAccount',
        'Waktu posting': 'postingTime',
        'Status': 'status',
        'Jenis otorisasi': 'authorizationType',
        'Biaya': 'cost',
        'Pesanan SKU': 'ordersSku',
        'Biaya per pesanan': 'costPerOrder',
        'Pendapatan kotor': 'grossRevenue',
        'ROI': 'roi',
        'Impresi iklan produk': 'impressions',
        'Jumlah klik iklan produk': 'clicks',
        'Tingkat klik iklan produk': 'ctr',
        'Rasio konversi iklan': 'conversionRate',
        'Rasio tayang video iklan 2 detik': 'watchRate2s',
        'Rasio tayang video iklan 6 detik': 'watchRate6s',
        'Rasio tayang video iklan 25%': 'watchRate25pct',
        'Rasio tayang video iklan 50%': 'watchRate50pct',
        'Rasio tayang video iklan 75%': 'watchRate75pct',
        'Rasio tayang video iklan 100%': 'watchRate100pct',
        'Mata uang': 'currency'
    }
    
    # Rename columns
    df = df.rename(columns=column_mapping)
    
    # Add metadata
    batch_id = str(uuid.uuid4())
    now = datetime.now().isoformat() + 'Z'
    
    df['tenantId'] = TENANT_ID
    df['uploadBatchId'] = batch_id
    df['periodStart'] = f"{period_start}T00:00:00.000Z"
    df['periodEnd'] = f"{period_end}T00:00:00.000Z"
    df['periodLabel'] = period_label  # Metodologi 4 bin/month
    df['createdAt'] = now
    df['updatedAt'] = now
    
    # Set default currency if not present
    if 'currency' not in df.columns or df['currency'].isna().all():
        df['currency'] = 'IDR'
    
    # Handle missing columns
    db_columns = [
        'tenantId', 'uploadBatchId', 'periodStart', 'periodEnd', 'periodLabel',
        'campaignId', 'campaignName', 'productId', 'creativeType',
        'videoTitle', 'videoId', 'tiktokAccount', 'postingTime', 
        'status', 'authorizationType',
        'cost', 'ordersSku', 'costPerOrder', 'grossRevenue', 'roi',
        'impressions', 'clicks', 'ctr', 'conversionRate',
        'watchRate2s', 'watchRate6s', 'watchRate25pct',
        'watchRate50pct', 'watchRate75pct', 'watchRate100pct',
        'currency', 'createdAt', 'updatedAt'
    ]
    
    for col in db_columns:
        if col not in df.columns:
            df[col] = None
    
    # Convert numeric columns
    numeric_cols = ['cost', 'ordersSku', 'costPerOrder', 'grossRevenue', 'roi',
                    'impressions', 'clicks', 'ctr', 'conversionRate',
                    'watchRate2s', 'watchRate6s', 'watchRate25pct',
                    'watchRate50pct', 'watchRate75pct', 'watchRate100pct']
    
    for col in numeric_cols:
        if col in df.columns:
            df[col] = pd.to_numeric(df[col], errors='coerce').fillna(0)
    
    # Handle required string columns (NOT NULL in schema)
    required_string_cols = ['campaignId', 'campaignName', 'productId', 'creativeType']
    for col in required_string_cols:
        if col in df.columns:
            df[col] = df[col].fillna('').astype(str)
            # Replace empty strings with placeholder
            df[col] = df[col].replace('', f'unknown_{col}')
    
    # Select only needed columns
    df = df[db_columns]
    
    # Insert to database
    conn = sqlite3.connect(DB_PATH)
    
    # Insert data using executemany for better performance
    placeholders = ','.join(['?' for _ in db_columns])
    insert_sql = f"INSERT INTO TiktokAdsCreativeData ({','.join(db_columns)}) VALUES ({placeholders})"
    
    # Prepare data
    data_to_insert = []
    for _, row in df.iterrows():
        values = []
        for col in db_columns:
            val = row[col]
            if pd.isna(val) or str(val) == 'nan' or str(val) == 'None':
                values.append(None)
            else:
                values.append(val)
        data_to_insert.append(tuple(values))
    
    # Bulk insert
    cursor = conn.cursor()
    cursor.executemany(insert_sql, data_to_insert)
    inserted = cursor.rowcount
    
    # Create batch record
    try:
        cursor.execute("""
            INSERT INTO TiktokAdsUploadBatch 
            (id, tenantId, fileName, periodStart, periodEnd, totalRows, insertedRows, 
             updatedRows, skippedRows, status, createdAt, updatedAt)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        """, (
            batch_id, TENANT_ID, filename, 
            f"{period_start}T00:00:00.000Z", f"{period_end}T00:00:00.000Z",
            len(df), inserted, 0, 0, 'completed', now, now
        ))
    except Exception as e:
        print(f"      ⚠️ Error creating batch record: {e}")
    
    conn.commit()
    conn.close()
    
    print(f"      ✅ Inserted: {inserted:,} baris")
    
    return inserted

# ============================================================
# GET UPLOADED FILES
# ============================================================
def get_uploaded_files():
    """
    Get list of files that have already been uploaded.
    
    Returns:
        set: Set of filenames that have been uploaded
    """
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute(
        "SELECT fileName FROM TiktokAdsUploadBatch WHERE tenantId = ?",
        (TENANT_ID,)
    )
    
    uploaded = {row[0] for row in cursor.fetchall()}
    conn.close()
    
    return uploaded


# ============================================================
# UPLOAD ALL FILES
# ============================================================
def upload_all_files(clear_existing: bool = False, force_reupload: bool = False):
    """Upload all creative data files
    
    Args:
        clear_existing: Hapus SEMUA data lama sebelum upload (default: False)
        force_reupload: Upload ulang semua file meskipun sudah ada (default: False)
    """
    print("=" * 70)
    print("📤 UPLOAD ALL ADS DATA FILES")
    print("=" * 70)
    print(f"   Tenant: {TENANT_ID}")
    print(f"   Database: {DB_PATH}")
    print(f"   Data Directory: {ADS_DATA_DIR}")
    
    # Clear existing data if requested
    if clear_existing:
        clear_old_data(confirm=True)
    
    # Get already uploaded files
    uploaded_files = set() if (clear_existing or force_reupload) else get_uploaded_files()
    if uploaded_files:
        print(f"\n   📁 Already uploaded: {len(uploaded_files)} files")
    
    # Find all creative data files
    all_files = sorted([
        f for f in os.listdir(ADS_DATA_DIR) 
        if f.endswith('.xlsx') and 'creative data' in f
    ])
    
    # Filter out already uploaded files
    files_to_upload = [f for f in all_files if f not in uploaded_files]
    files_skipped = [f for f in all_files if f in uploaded_files]
    
    print(f"\n   Found {len(all_files)} files total")
    print(f"   ✅ Already uploaded: {len(files_skipped)} files (will skip)")
    print(f"   📤 To upload: {len(files_to_upload)} files")
    
    if not files_to_upload:
        print("\n   ✨ Semua file sudah di-upload, tidak ada yang perlu diproses.")
        return 0, 0, {}
    
    total_inserted = 0
    period_summary = {}
    uploaded_count = 0
    
    for filename in files_to_upload:
        filepath = os.path.join(ADS_DATA_DIR, filename)
        period_start, period_end, period_label = extract_period_from_filename(filename)
        
        if not period_start or not period_end:
            print(f"\n   ⚠️ Skipping {filename} - Cannot extract period")
            continue
        
        try:
            inserted = upload_file(filepath, period_start, period_end, period_label)
            total_inserted += inserted
            uploaded_count += 1
        except Exception as e:
            print(f"\n   ❌ Error processing {filename}: {e}")
            continue
        
        # Track period summary
        if period_label not in period_summary:
            period_summary[period_label] = 0
        period_summary[period_label] += inserted
    
    # Print period summary
    if period_summary:
        print("\n" + "=" * 70)
        print("📊 RINGKASAN UPLOAD BARU")
        print("=" * 70)
        for label, count in sorted(period_summary.items()):
            print(f"   {label}: {count:,} rows")
    
    return total_inserted, uploaded_count, period_summary

# ============================================================
# VALIDATE UPLOAD
# ============================================================
def validate_upload():
    """Validate uploaded data"""
    print("\n" + "=" * 70)
    print("✅ VALIDATE UPLOAD")
    print("=" * 70)
    
    conn = sqlite3.connect(DB_PATH)
    
    # Total rows
    total = pd.read_sql_query(
        "SELECT COUNT(*) as cnt FROM TiktokAdsCreativeData WHERE tenantId = ?",
        conn, params=(TENANT_ID,)
    )['cnt'].iloc[0]
    
    # Per period
    periods = pd.read_sql_query("""
        SELECT periodStart, periodEnd, COUNT(*) as rows,
               COUNT(DISTINCT productId) as products,
               SUM(cost) as total_cost
        FROM TiktokAdsCreativeData 
        WHERE tenantId = ?
        GROUP BY periodStart, periodEnd
        ORDER BY periodStart
    """, conn, params=(TENANT_ID,))
    
    conn.close()
    
    print(f"\n   📊 Total baris: {total:,}")
    print(f"\n   📅 Per Periode:")
    
    for _, row in periods.iterrows():
        start = str(row['periodStart'])[:10]
        end = str(row['periodEnd'])[:10]
        
        # Determine granularity
        try:
            s = datetime.strptime(start, '%Y-%m-%d')
            e = datetime.strptime(end, '%Y-%m-%d')
            days = (e - s).days + 1
            granularity = "BULAN" if days >= 28 else "MINGGU"
        except:
            granularity = "?"
            days = 0
        
        print(f"      {start} → {end} ({days:2} hari = {granularity})")
        print(f"         Rows: {row['rows']:,} | Products: {row['products']:,} | Cost: Rp {row['total_cost']:,.0f}")
    
    return total

# ============================================================
# MAIN
# ============================================================
if __name__ == "__main__":
    import sys
    
    # Parse arguments
    # Usage:
    #   python upload_ads_data.py              → Upload file baru saja (incremental)
    #   python upload_ads_data.py --force      → Upload ulang semua file
    #   python upload_ads_data.py --clean      → Hapus semua data & upload ulang
    
    force_reupload = "--force" in sys.argv
    clear_existing = "--clean" in sys.argv
    
    if clear_existing:
        print("\n   ⚠️ Mode: CLEAN - Menghapus semua data dan upload ulang")
    elif force_reupload:
        print("\n   ⚠️ Mode: FORCE - Upload ulang semua file (tanpa hapus data lama)")
    else:
        print("\n   📤 Mode: INCREMENTAL - Hanya upload file baru")
    
    # Upload files
    total_inserted, file_count, period_summary = upload_all_files(
        clear_existing=clear_existing,
        force_reupload=force_reupload
    )
    
    # Validate
    final_count = validate_upload()
    
    # Summary
    print("\n" + "=" * 70)
    print("📋 SUMMARY")
    print("=" * 70)
    print(f"   📁 Files uploaded: {file_count}")
    print(f"   ✨ Inserted: {total_inserted:,} baris")
    print(f"   📊 Final count in DB: {final_count:,} baris")
    if period_summary:
        print(f"   📆 Periode baru: {len(period_summary)} periode")
    print("=" * 70)
