#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Shopee Ads Analysis - Data Loading
==================================
Functions untuk load data dari database (primary) atau CSV files (fallback).
"""

import re
import sqlite3
import pandas as pd
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional, Tuple

from .config import (
    DATA_DIR, COLUMN_MAP, REQUIRED_COLUMNS,
    CSV_SKIP_ROWS, CSV_ENCODING, TENANT_ID, PACKAGE_DIR,
)

# Database path (same as upload script)
DB_PATH = PACKAGE_DIR.parent.parent / "backend" / "config" / "databases" / "yumna_bertigamart.db"


def get_period_from_filename(filename: str) -> Tuple[Optional[datetime], Optional[datetime]]:
    """
    Extract period start and end dates from filename.
    
    Example: Data-+Semua-Iklan-Produk-08_01_2026-15_01_2026.csv
    Returns: (datetime(2026, 1, 8), datetime(2026, 1, 15))
    """
    pattern = r"(\d{2})_(\d{2})_(\d{4})-(\d{2})_(\d{2})_(\d{4})"
    match = re.search(pattern, filename)
    
    if match:
        day1, month1, year1, day2, month2, year2 = match.groups()
        start = datetime(int(year1), int(month1), int(day1))
        end = datetime(int(year2), int(month2), int(day2))
        return start, end
    
    return None, None


def get_period_label(start: datetime, end: datetime) -> str:
    """Generate period label like '2026-W02' from dates."""
    # Use ISO week number from start date
    year, week, _ = start.isocalendar()
    return f"{year}-W{week:02d}"


def load_single_csv(filepath: Path, verbose: bool = True) -> Optional[pd.DataFrame]:
    """
    Load a single Shopee ads CSV file.
    
    Args:
        filepath: Path to CSV file
        verbose: Print progress messages
        
    Returns:
        DataFrame with normalized columns, or None if failed
    """
    def log(msg):
        if verbose:
            print(msg)
    
    try:
        df = pd.read_csv(filepath, skiprows=CSV_SKIP_ROWS, encoding=CSV_ENCODING)
        
        # Rename columns to internal names
        df = df.rename(columns=COLUMN_MAP)
        
        # Extract period from filename
        start, end = get_period_from_filename(filepath.name)
        
        if start and end:
            df["period_start"] = start
            df["period_end"] = end
            df["period_label"] = get_period_label(start, end)
        else:
            log(f"   ⚠️ Could not parse period from {filepath.name}")
            return None
        
        # Parse numeric columns (remove % and commas)
        for col in ["ctr", "conversion_rate", "direct_conversion_rate", "acos", "direct_acos"]:
            if col in df.columns:
                df[col] = df[col].astype(str).str.replace("%", "").str.replace(",", ".").astype(float)
        
        # Ensure product_id is string
        if "product_id" in df.columns:
            df["product_id"] = df["product_id"].astype(str)
        
        return df
        
    except Exception as e:
        log(f"   ❌ Error loading {filepath.name}: {e}")
        return None


def load_from_database(verbose: bool = True) -> pd.DataFrame:
    """
    Load data from SQLite database (primary source).
    
    Returns:
        DataFrame with all product data
    """
    def log(msg):
        if verbose:
            print(msg)
    
    log("📊 Loading Shopee Ads from database...")
    
    conn = sqlite3.connect(str(DB_PATH))
    
    df = pd.read_sql_query("""
        SELECT 
            productId as product_id,
            productName as product_name,
            status,
            biddingMode as bidding_mode,
            placement,
            periodStart as period_start,
            periodEnd as period_end,
            periodLabel as period_label,
            impressions,
            clicks,
            ctr,
            conversions,
            directConversions as direct_conversions,
            conversionRate as conversion_rate,
            directConversionRate as direct_conversion_rate,
            unitsSold as units_sold,
            directUnitsSold as direct_units_sold,
            revenue,
            directRevenue as direct_revenue,
            cost,
            roas,
            directRoas as direct_roas,
            acos,
            directAcos as direct_acos
        FROM ShopeeAdsProductData 
        WHERE tenantId = ? AND cost > 0
        ORDER BY periodStart
    """, conn, params=(TENANT_ID,))
    
    conn.close()
    
    log(f"   ✅ Loaded {len(df):,} records from database")
    
    # Parse dates
    df['period_start'] = pd.to_datetime(df['period_start'])
    df['period_end'] = pd.to_datetime(df['period_end'])
    
    return df


def load_all_csv_files(verbose: bool = True) -> pd.DataFrame:
    """
    Load all CSV files from data directory (fallback).
    
    Returns:
        Combined DataFrame with all periods
    """
    def log(msg):
        if verbose:
            print(msg)
    
    log("📊 Loading Shopee Ads CSV files...")
    
    csv_files = sorted(DATA_DIR.glob("Data-+Semua-Iklan-Produk-*.csv"))
    
    if not csv_files:
        raise FileNotFoundError(f"No CSV files found in {DATA_DIR}")
    
    log(f"   📁 Found {len(csv_files)} CSV files")
    
    dfs = []
    for filepath in csv_files:
        df = load_single_csv(filepath, verbose=False)
        if df is not None:
            dfs.append(df)
    
    if not dfs:
        raise ValueError("No valid CSV files could be loaded")
    
    combined = pd.concat(dfs, ignore_index=True)
    
    log(f"   ✅ Loaded {len(combined):,} records from {len(dfs)} files")
    
    return combined


def load_data(verbose: bool = True, use_database: bool = True) -> pd.DataFrame:
    """
    Main data loading function.
    
    Args:
        verbose: Print progress messages
        use_database: If True, load from database (default). If False, load from CSV.
    
    Returns:
        DataFrame with cleaned and normalized data
    """
    # Try database first if requested
    if use_database and DB_PATH.exists():
        try:
            df = load_from_database(verbose=verbose)
            if len(df) > 0:
                # Calculate profit
                df["profit"] = df["revenue"] - df["cost"]
                return df
        except Exception as e:
            if verbose:
                print(f"   ⚠️ Database load failed: {e}, falling back to CSV")
    
    # Fallback to CSV
    df = load_all_csv_files(verbose=verbose)
    
    # Filter out rows with zero cost (no actual spend)
    df = df[df["cost"] > 0].copy()
    
    # Calculate profit
    df["profit"] = df["revenue"] - df["cost"]
    
    # Calculate ROI if not present
    if "roas" not in df.columns or df["roas"].isna().all():
        df["roas"] = df["revenue"] / df["cost"]
    
    return df


def get_available_periods(df: pd.DataFrame) -> List[str]:
    """Get list of available period labels sorted chronologically."""
    return sorted(df["period_label"].unique().tolist())


def filter_data_by_window(df: pd.DataFrame, n_months: int = 3, verbose: bool = True) -> pd.DataFrame:
    """
    Filter data to only include last n months (rolling window).
    """
    def log(msg):
        if verbose:
            print(msg)
    
    # Get the latest period end date
    latest_date = df["period_end"].max()
    
    # Calculate cutoff date
    from dateutil.relativedelta import relativedelta
    cutoff_date = latest_date - relativedelta(months=n_months)
    
    # Filter
    filtered = df[df["period_start"] >= cutoff_date].copy()
    
    periods_before = df["period_label"].nunique()
    periods_after = filtered["period_label"].nunique()
    
    log(f"   📅 Rolling window: {n_months} months → {periods_after}/{periods_before} periods")
    
    return filtered


def summary_stats(df: pd.DataFrame) -> Dict:
    """
    Get quick summary statistics.
    
    Returns:
        dict: Summary statistics
    """
    return {
        "total_cost": df["cost"].sum(),
        "total_revenue": df["revenue"].sum(),
        "total_profit": df["revenue"].sum() - df["cost"].sum(),
        "roi": df["revenue"].sum() / df["cost"].sum() if df["cost"].sum() > 0 else 0,
        "total_products": df["product_id"].nunique(),
        "total_records": len(df),
        "date_range": f"{df['period_start'].min()} to {df['period_end'].max()}"
    }
