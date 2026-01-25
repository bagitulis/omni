#!/usr/bin/env python
"""
Shopee Ads Data - Upload Script
================================
Upload data dari file CSV ke database dengan periode dari nama file.

Format: Data-+Semua-Iklan-Produk-DD_MM_YYYY-DD_MM_YYYY.csv
Period label: YYYY-Www (ISO week number)
"""

import os
import re
import uuid
from datetime import datetime, timezone
from pathlib import Path

import pandas as pd
import psycopg2
from psycopg2.extras import execute_values

from upload_config import (
    TENANT_ID, ADS_DATA_DIR,
    PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_DATABASE, PG_SCHEMA,
    CSV_SKIP_ROWS, CSV_ENCODING,
    CSV_TO_DB_COLUMN_MAP, DB_COLUMNS, NUMERIC_COLUMNS, PERCENTAGE_COLUMNS
)

def extract_period_from_filename(filename: str) -> tuple:
    """Extract period from filename. Returns (period_start, period_end, period_label)."""
    pattern = r"(\d{2})_(\d{2})_(\d{4})-(\d{2})_(\d{2})_(\d{4})"
    match = re.search(pattern, filename)
    if not match:
        return None, None, None
    
    day1, month1, year1, day2, month2, year2 = match.groups()
    start = datetime(int(year1), int(month1), int(day1))
    end = datetime(int(year2), int(month2), int(day2))
    year, week, _ = start.isocalendar()
    return start.strftime("%Y-%m-%d"), end.strftime("%Y-%m-%d"), f"{year}-W{week:02d}"

def clear_old_data(confirm: bool = False):
    """Hapus semua data Shopee Ads yang ada di database."""
    if not confirm:
        print("   ⚠️ Set confirm=True untuk menghapus data.")
        return 0, 0
    
    conn = psycopg2.connect(
        host=PG_HOST,
        port=PG_PORT,
        user=PG_USER,
        password=PG_PASSWORD,
        dbname=PG_DATABASE,
    )
    cursor = conn.cursor()
    cursor.execute("SET search_path TO %s, public", (PG_SCHEMA,))
    
    cursor.execute("SELECT COUNT(*) FROM shopee_ads_product_data WHERE tenant_id = %s", (TENANT_ID,))
    count_product = cursor.fetchone()[0]
    cursor.execute("SELECT COUNT(*) FROM shopee_ads_upload_batches WHERE tenant_id = %s", (TENANT_ID,))
    count_batch = cursor.fetchone()[0]
    
    print(f"\n   🗑️ Menghapus: {count_product:,} product data + {count_batch:,} batch")
    
    cursor.execute("DELETE FROM shopee_ads_product_data WHERE tenant_id = %s", (TENANT_ID,))
    cursor.execute("DELETE FROM shopee_ads_upload_batches WHERE tenant_id = %s", (TENANT_ID,))
    conn.commit()
    conn.close()
    
    print(f"   ✅ Berhasil dihapus")
    return count_product, count_batch


# ============================================================
# UPLOAD SINGLE FILE
# ============================================================
def upload_file(filepath: Path, period_start: str, period_end: str, period_label: str):
    """Upload single CSV file to database."""
    
    filename = filepath.name
    print(f"\n   📄 {filename}")
    print(f"      Periode: {period_start} → {period_end}")
    print(f"      Label: {period_label}")
    
    # Read CSV
    df = pd.read_csv(filepath, skiprows=CSV_SKIP_ROWS, encoding=CSV_ENCODING)
    print(f"      Rows: {len(df):,}")
    
    # Rename columns
    df = df.rename(columns=CSV_TO_DB_COLUMN_MAP)

    # Add metadata
    batch_id = str(uuid.uuid4())
    now = datetime.now(timezone.utc)
    
    df['tenant_id'] = TENANT_ID
    df['upload_batch_id'] = batch_id
    df['period_start'] = f"{period_start}T00:00:00Z"
    df['period_end'] = f"{period_end}T00:00:00Z"
    df['period_label'] = period_label
    df['created_at'] = now
    df['updated_at'] = now
    
    # Parse percentage columns (remove %)
    for col in PERCENTAGE_COLUMNS:
        if col in df.columns:
            df[col] = df[col].astype(str).str.replace('%', '').str.replace(',', '.').astype(float)
    
    # Parse startDate (campaign start date)
    if 'start_date' in df.columns:
        df['start_date'] = pd.to_datetime(df['start_date'], format='%d/%m/%Y %H:%M:%S', errors='coerce')
        df['start_date'] = df['start_date'].dt.strftime('%Y-%m-%dT%H:%M:%SZ')
    
    # Ensure product_id is string
    df['product_id'] = df['product_id'].astype(str)
    
    # Ensure all DB columns exist
    for col in DB_COLUMNS:
        if col not in df.columns:
            df[col] = None
    
    # Convert numeric columns
    for col in NUMERIC_COLUMNS:
        if col in df.columns:
            df[col] = pd.to_numeric(df[col], errors='coerce').fillna(0)
    
    # Insert to database
    conn = psycopg2.connect(
        host=PG_HOST,
        port=PG_PORT,
        user=PG_USER,
        password=PG_PASSWORD,
        dbname=PG_DATABASE,
    )
    cursor = conn.cursor()
    cursor.execute("SET search_path TO %s, public", (PG_SCHEMA,))

    # Create batch record first
    cursor.execute(
        """
        INSERT INTO shopee_ads_upload_batches 
        (id, tenant_id, file_name, period_start, period_end, period_label, total_rows, status, created_at, updated_at)
        VALUES (%s, %s, %s, %s, %s, %s, %s, 'processing', %s, %s)
        """,
        (batch_id, TENANT_ID, filename, f"{period_start}T00:00:00Z", f"{period_end}T00:00:00Z", period_label, len(df), now, now),
    )

    # Prepare product data rows
    inserted = len(df)
    skipped = 0

    records = []
    for _, row in df.iterrows():
        try:
            records.append(
                (
                    row['tenant_id'], row['upload_batch_id'], row['period_start'], row['period_end'], row['period_label'],
                    row['product_id'], row['product_name'], row['status'], row['bidding_mode'], row['placement'],
                    row['start_date'], row['end_date'],
                    int(row['impressions']), int(row['clicks']), float(row['ctr']),
                    int(row['conversions']), int(row['direct_conversions']), float(row['conversion_rate']), float(row['direct_conversion_rate']),
                    float(row['cost_per_conversion']), float(row['cost_per_direct_conversion']),
                    int(row['units_sold']), int(row['direct_units_sold']), float(row['revenue']), float(row['direct_revenue']),
                    float(row['cost']), float(row['roas']), float(row['direct_roas']), float(row['acos']), float(row['direct_acos']),
                    now, now,
                )
            )
        except Exception as e:
            skipped += 1
            inserted -= 1
            print(f"      ⚠️ Skip row: {e}")

    if records:
        execute_values(
            cursor,
            """
            INSERT INTO shopee_ads_product_data (
                tenant_id, upload_batch_id, period_start, period_end, period_label,
                product_id, product_name, status, bidding_mode, placement,
                start_date, end_date, impressions, clicks, ctr,
                conversions, direct_conversions, conversion_rate, direct_conversion_rate,
                cost_per_conversion, cost_per_direct_conversion,
                units_sold, direct_units_sold, revenue, direct_revenue,
                cost, roas, direct_roas, acos, direct_acos,
                created_at, updated_at
            ) VALUES %s
            """,
            records,
        )

    # Update batch status
    cursor.execute(
        """
        UPDATE shopee_ads_upload_batches 
        SET status = 'completed', inserted_rows = %s, skipped_rows = %s, updated_rows = 0, updated_at = %s
        WHERE id = %s
        """,
        (inserted, skipped, datetime.now(timezone.utc), batch_id),
    )

    conn.commit()
    conn.close()

    print(f"      ✅ Inserted: {inserted:,}, Skipped: {skipped:,}")

    return inserted, skipped


# ============================================================
# UPLOAD ALL FILES
# ============================================================
def upload_all_files(clear_first: bool = False):
    """Upload all CSV files from data directory."""
    
    print("\n" + "="*60)
    print("📊 SHOPEE ADS DATA UPLOAD")
    print("="*60)
    print(f"📁 Data Directory: {ADS_DATA_DIR}")
    print(f"🗄️ Database: postgres://{PG_USER}@{PG_HOST}:{PG_PORT}/{PG_DATABASE} (schema {PG_SCHEMA})")
    print(f"🏢 Tenant: {TENANT_ID}")
    
    # Find all CSV files
    csv_files = sorted(ADS_DATA_DIR.glob("Data-+Semua-Iklan-Produk-*.csv"))
    
    if not csv_files:
        print("\n❌ No CSV files found!")
        return
    
    print(f"\n📄 Found {len(csv_files)} CSV files")
    
    # Clear old data if requested
    if clear_first:
        clear_old_data(confirm=True)
    
    # Process each file
    total_inserted = 0
    total_skipped = 0
    
    for filepath in csv_files:
        period_start, period_end, period_label = extract_period_from_filename(filepath.name)
        
        if period_start and period_end:
            inserted, skipped = upload_file(filepath, period_start, period_end, period_label)
            total_inserted += inserted
            total_skipped += skipped
        else:
            print(f"\n   ⚠️ Skipping {filepath.name} - could not parse period")
    
    print("\n" + "="*60)
    print(f"✅ UPLOAD COMPLETE")
    print(f"   Total Inserted: {total_inserted:,}")
    print(f"   Total Skipped: {total_skipped:,}")
    print("="*60)


# ============================================================
# MAIN
# ============================================================
if __name__ == "__main__":
    import sys
    
    clear_first = "--clear" in sys.argv or "-c" in sys.argv
    upload_all_files(clear_first=clear_first)
