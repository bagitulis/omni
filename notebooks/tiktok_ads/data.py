#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Data Loading
==================================
Functions untuk load data dari database dan Excel.
"""

from pathlib import Path

import pandas as pd
import psycopg2

from .config import (
    TENANT_ID,
    EXCEL_PRODUCT_REF,
    PG_HOST,
    PG_PORT,
    PG_USER,
    PG_PASSWORD,
    PG_DATABASE,
    PG_SCHEMA,
)


def load_product_names(verbose=True):
    """Load product names from Excel reference file
    
    Args:
        verbose: Print progress messages
        
    Returns:
        dict: Mapping of product_id -> product_name
    """
    def log(msg):
        if verbose:
            print(msg)
    
    try:
        df = pd.read_excel(str(EXCEL_PRODUCT_REF), skiprows=3)
        # Rename columns based on actual structure
        if 'ID Produk' in df.columns:
            df = df.rename(columns={'ID Produk': 'product_id', 'Nama Produk': 'product_name'})
        elif len(df.columns) >= 2:
            df.columns = ['product_id', 'platform', 'category', 'brand', 'product_name'] + list(df.columns[5:])
        
        # Create mapping
        product_map = {}
        for _, row in df.iterrows():
            pid = str(row.get('product_id', '')).strip()
            pname = str(row.get('product_name', '')).strip()
            if pid and pname and pid not in ['V3', 'ID Produk', 'Wajib', 'Tidak dapat diedit']:
                product_map[pid] = pname[:80]  # Limit name length
        
        log(f"   ✅ Loaded {len(product_map)} product names from Excel")
        return product_map
    except Exception as e:
        log(f"   ⚠️ Warning: Could not load Excel ({e})")
        return {}


def load_data(verbose=True):
    """Load data from database
    
    Args:
        verbose: Print progress messages
        
    Returns:
        pd.DataFrame: Ads data with parsed dates and period labels
    """
    def log(msg):
        if verbose:
            print(msg)
    
    log("📊 Loading data...")
    
    conn = psycopg2.connect(
        host=PG_HOST,
        port=PG_PORT,
        user=PG_USER,
        password=PG_PASSWORD,
        dbname=PG_DATABASE,
    )
    with conn.cursor() as cur:
        cur.execute("SET search_path TO %s, public", (PG_SCHEMA,))
    df = pd.read_sql_query(
        """
        SELECT 
            product_id, campaign_name, creative_type, video_title,
            period_start, period_end, period_label,
            cost, gross_revenue, orders_sku, roi,
            impressions, clicks, ctr, conversion_rate
        FROM tiktok_ads_creative_data 
        WHERE tenant_id = %s AND cost > 0
        ORDER BY period_start
        """,
        conn,
        params=(TENANT_ID,),
    )
    conn.close()
    
    log(f"   ✅ Loaded {len(df):,} records from database")
    
    # Parse dates
    df['periodStart'] = pd.to_datetime(df['periodStart'])
    df['periodEnd'] = pd.to_datetime(df['periodEnd'])
    
    return df


def get_product_display_name(product_id, campaign_name, video_title, product_map):
    """Get best available product name
    
    Priority:
    1. Excel mapping (most accurate)
    2. Video title (if contains 'Nusseyba')
    3. Campaign name (fallback)
    """
    pid = str(product_id)
    
    # 1. Try Excel mapping first
    if pid in product_map:
        return product_map[pid]
    
    # 2. Try video title if meaningful
    if video_title and str(video_title) not in ['-', 'None', 'nan', '']:
        title = str(video_title).strip()
        if len(title) > 10 and 'Nusseyba' in title:
            return title[:80]
    
    # 3. Fallback to campaign name
    return f"Campaign: {campaign_name[:60]}" if campaign_name else f"Product-{pid[-6:]}"


def summary_stats(df):
    """Get quick summary statistics
    
    Args:
        df: DataFrame from load_data()
        
    Returns:
        dict: Summary statistics
    """
    return {
        'total_cost': df['cost'].sum(),
        'total_revenue': df['grossRevenue'].sum(),
        'total_profit': df['grossRevenue'].sum() - df['cost'].sum(),
        'roi': df['grossRevenue'].sum() / df['cost'].sum() if df['cost'].sum() > 0 else 0,
        'total_products': df['productId'].nunique(),
        'total_records': len(df),
        'date_range': f"{df['periodStart'].min()} to {df['periodEnd'].max()}"
    }
