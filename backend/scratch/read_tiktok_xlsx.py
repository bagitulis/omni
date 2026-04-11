import pandas as pd
import sys

file_path = r"D:\Project\omni\notebooks\tiktok_ads\data\creative data for product campaigns 2026-01-01 00 ~ 2026-01-07 23.xlsx"
try:
    df = pd.read_excel(file_path)
    # Select some key columns to compare
    # Try to match the headers from the mapping in backend
    cols = ["Campaign name", "Product ID", "Video title", "Cost", "Gross revenue", "SKU orders"]
    # Check if columns exist
    existing_cols = [c for c in cols if c in df.columns]
    print(df[existing_cols].head(10).to_string())
except Exception as e:
    print(f"Error: {e}")
