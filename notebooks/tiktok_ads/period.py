#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
TikTok Ads Analysis - Period Functions
======================================
Period bin, rolling window, dan dynamic period handling.
"""

import pandas as pd
from datetime import timedelta
from dateutil.relativedelta import relativedelta


def get_period_bin(date):
    """
    Get period bin (1-4) from date.
    
    Metodologi 4 Bins per Month:
    - Early Month (Tgl 1-7): Awal bulan, baru gajian
    - Mid Month I (Tgl 8-15): Pertengahan awal
    - Mid Month II (Tgl 16-23): Pertengahan akhir
    - Late Month (Tgl 24-End): Akhir bulan
    
    Returns:
        tuple: (bin_number, bin_name)
    """
    if isinstance(date, str):
        date = pd.to_datetime(date)
    
    day = date.day
    if day <= 7:
        return 1, "Early Month"
    elif day <= 15:
        return 2, "Mid Month I"
    elif day <= 23:
        return 3, "Mid Month II"
    else:
        return 4, "Late Month"


def get_period_label(date):
    """Get full period label (YYYY-MM Bin Name)"""
    if isinstance(date, str):
        date = pd.to_datetime(date)
    
    year_month = date.strftime('%Y-%m')
    _, bin_name = get_period_bin(date)
    return f"{year_month} {bin_name}"


def get_available_periods(df, verbose=False):
    """
    Otomatis detect semua periode yang tersedia di database.
    Tidak ada hardcode - membaca langsung dari data.
    
    Returns:
        list: Sorted list of period labels
    """
    periods = sorted(df['periodLabel'].unique().tolist())
    if verbose:
        print(f"   📅 Detected {len(periods)} periods: {periods[0]} to {periods[-1]}")
    return periods


def get_rolling_window_periods(df, n_months=3, reference_date=None):
    """
    Get periods for rolling window analysis.
    Otomatis adjust berdasarkan tanggal - TIDAK PERLU SETTING MANUAL.
    
    Misalnya:
    - Jika data terakhir Januari 2026, ambil 3 bulan terakhir (Nov 2025 - Jan 2026)
    - Jika data terakhir April 2026, ambil 3 bulan terakhir (Feb - Apr 2026)
    
    Args:
        df: DataFrame dengan periodLabel
        n_months: Jumlah bulan untuk window (default: 3 = quarterly)
        reference_date: Tanggal referensi (default: tanggal terbaru di data)
    
    Returns:
        dict: Window information
    """
    # Auto-detect latest date from data
    if reference_date is None:
        reference_date = df['periodEnd'].max()
    
    # Convert to pandas timestamp and remove timezone if present
    if isinstance(reference_date, str):
        reference_date = pd.to_datetime(reference_date)
    
    # Make timezone-naive for comparison
    if hasattr(reference_date, 'tz') and reference_date.tz is not None:
        reference_date = reference_date.tz_localize(None)
    
    reference_date = pd.Timestamp(reference_date)
    
    # End date is end of the reference month
    end_date = reference_date.replace(day=1) + relativedelta(months=1) - timedelta(days=1)
    
    # Start date is n_months back (inclusive)
    start_date = reference_date.replace(day=1) - relativedelta(months=n_months-1)
    
    # Get quarter label based on the data range
    end_month = end_date.month
    quarter = (end_month - 1) // 3 + 1
    quarter_label = f"Q{quarter} {end_date.year}"
    
    # Alternative: use descriptive label
    start_str = start_date.strftime('%b %Y')
    end_str = end_date.strftime('%b %Y')
    range_label = f"{start_str} - {end_str}"
    
    # Filter periods within window by parsing period labels
    all_periods = get_available_periods(df, verbose=False)
    filtered_periods = []
    
    for period in all_periods:
        try:
            period_ym = period.split()[0]  # "2025-10"
            period_date = pd.Timestamp(period_ym + "-01")
            # Check if period falls within our window (comparing months)
            if (start_date.year * 12 + start_date.month <= 
                period_date.year * 12 + period_date.month <= 
                end_date.year * 12 + end_date.month):
                filtered_periods.append(period)
        except Exception:
            continue
    
    return {
        'start_date': start_date,
        'end_date': end_date,
        'periods': filtered_periods,
        'n_periods': len(filtered_periods),
        'quarter_label': quarter_label,
        'range_label': range_label,
        'window_months': n_months
    }


def filter_data_by_window(df, n_months=3, reference_date=None, verbose=True):
    """
    Filter data untuk rolling window analysis.
    
    Args:
        df: Full DataFrame
        n_months: Window size (default: 3)
        reference_date: Reference date (default: latest in data)
        
    Returns:
        pd.DataFrame: Filtered data for window
    """
    window = get_rolling_window_periods(df, n_months, reference_date)
    
    if verbose:
        print(f"   🔄 Rolling Window: {window['quarter_label']}")
        print(f"      Period: {window['start_date'].strftime('%Y-%m-%d')} to {window['end_date'].strftime('%Y-%m-%d')}")
        print(f"      Periods: {window['n_periods']} ({len(window['periods'])} period labels)")
    
    filtered_df = df[df['periodLabel'].isin(window['periods'])].copy()
    return filtered_df


def analyze_period_performance(df, verbose=True):
    """
    Analisis performa per periode dengan metodologi 4 bin/month.
    
    Returns:
        list[dict]: Summary per periode
    """
    def log(msg):
        if verbose:
            print(msg)
    
    log("📅 Analyzing period performance...")
    
    period_summary = df.groupby('periodLabel').agg({
        'productId': 'nunique',
        'cost': 'sum',
        'grossRevenue': 'sum',
        'ordersSku': 'sum'
    }).reset_index()
    
    period_summary['roi'] = period_summary['grossRevenue'] / period_summary['cost']
    period_summary['profit'] = period_summary['grossRevenue'] - period_summary['cost']
    
    # Sort by period
    period_summary = period_summary.sort_values('periodLabel')
    
    result = []
    for _, row in period_summary.iterrows():
        result.append({
            'period': row['periodLabel'],
            'products': int(row['productId']),
            'cost': int(row['cost']),
            'revenue': int(row['grossRevenue']),
            'profit': int(row['profit']),
            'roi': round(row['roi'], 2),
            'orders': int(row['ordersSku'])
        })
    
    log(f"   ✅ Analyzed {len(result)} periods")
    return result
