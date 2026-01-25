#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Product Resolver
================
Maps product IDs to names using JSON (primary) and Excel (fallback).
Following AGENTS.MD: Clean Code, DRY, SRP.
"""

from pathlib import Path
from typing import Dict, Optional, Set
import pandas as pd
import json


class ProductResolver:
    """
    Resolves product IDs to human-readable names.
    
    Priority:
    1. JSON cache file (product_names.json) - Primary, most reliable
    2. Excel reference file - Fallback
    
    JSON format supports:
    - Flat: {"id": "name"}
    - Nested: {"id": {"name": "...", "category": "..."}}
    """
    
    # Default paths - data is at tiktok_ads/data (same level as intelligence)
    DEFAULT_JSON_PATH = Path(__file__).parent.parent / "data" / "product_names.json"
    DEFAULT_EXCEL_PATH = Path(__file__).parent.parent / "data" / "Tiktoksellercenter_batchedit_20260114_basic_information_template.xlsx"
    
    def __init__(self, excelPath: str = None, jsonCachePath: str = None):
        # Use provided paths or defaults
        self.jsonCachePath = Path(jsonCachePath) if jsonCachePath else self.DEFAULT_JSON_PATH
        self.excelPath = Path(excelPath) if excelPath else self.DEFAULT_EXCEL_PATH
        self._productMap: Dict[str, str] = {}
        self._categoryMap: Dict[str, str] = {}  # Store categories too
        self._discontinuedProducts: Set[str] = set()
        self._loaded = False
    
    def load(self) -> None:
        """Load product mappings from sources."""
        if self._loaded:
            return
        
        # Try JSON first (primary source - more reliable)
        if self.jsonCachePath and self.jsonCachePath.exists():
            self._loadFromJson()
        
        # Then Excel as fallback (may have additional products)
        if self.excelPath and self.excelPath.exists():
            self._loadFromExcel()
        
        self._loaded = True
    
    def _loadFromJson(self) -> None:
        """Load from JSON cache file (primary source)."""
        try:
            with open(self.jsonCachePath, 'r', encoding='utf-8') as f:
                data = json.load(f)
                
            for pid, value in data.items():
                # Skip invalid IDs
                if not pid or pid in ['V3', 'ID Produk', 'Wajib', 'Tidak dapat diedit']:
                    continue
                    
                if isinstance(value, dict):
                    # Nested format: {"name": "...", "category": "..."}
                    name = value.get('name', '')
                    category = value.get('category', '')
                    if name:
                        self._productMap[pid] = name[:80]
                        if category:
                            self._categoryMap[pid] = category
                else:
                    # Flat format: "name"
                    if value:
                        self._productMap[pid] = str(value)[:80]
                        
        except Exception as e:
            print(f"Warning: Could not load JSON ({e})")
    
    def _loadFromExcel(self) -> None:
        """Load from Excel reference file (fallback)."""
        try:
            # TikTok Seller Center export format:
            # Row 0: English headers (product_id, sale_platforms, category, brand, product_name, ...)
            # Row 1: Version info (V3, Basic_Information, ...)
            # Row 2: Indonesian headers (ID Produk, Platform, Kategori, Merek, Nama Produk, ...)
            # Row 3: Required/Optional labels (Wajib, Wajib, Wajib, Opsional, Wajib, ...)
            # Row 4: Edit instructions (Tidak dapat diedit, Bisa diubah, ...)
            # Row 5+: Actual data
            
            df = pd.read_excel(str(self.excelPath), header=None, skiprows=5)
            
            # Column indices based on TikTok format
            # [0] product_id, [1] sale_platforms, [2] category, [3] brand, [4] product_name
            
            for _, row in df.iterrows():
                pid = str(row.iloc[0]).strip() if pd.notna(row.iloc[0]) else ''
                pname = str(row.iloc[4]).strip() if len(row) > 4 and pd.notna(row.iloc[4]) else ''
                category = str(row.iloc[2]).strip() if len(row) > 2 and pd.notna(row.iloc[2]) else ''
                
                # Skip invalid rows
                if not pid or not pname:
                    continue
                if pid in ['V3', 'ID Produk', 'Wajib', 'Tidak dapat diedit', 'nan']:
                    continue
                if not pid.isdigit():
                    continue
                    
                # Only add if not already in map (JSON takes priority)
                if pid not in self._productMap:
                    self._productMap[pid] = pname[:80]
                    if category and pid not in self._categoryMap:
                        self._categoryMap[pid] = category
                        
        except Exception as e:
            print(f"Warning: Could not load Excel ({e})")
    
    def resolve(self, productId: str, campaignName: str = "",
                videoTitle: str = "") -> str:
        """
        Resolve product ID to display name.
        
        Priority:
        1. JSON/Excel mapping (most accurate - actual product names)
        2. Video title (if contains 'Nusseyba' - likely product name)
        3. Campaign name (fallback - for discontinued/unmapped products)
        4. Product ID suffix (last resort)
        
        Args:
            productId: Product ID to resolve
            campaignName: Campaign name as fallback
            videoTitle: Video title as fallback
        
        Returns:
            Human-readable product name
        """
        self.load()
        
        pid = str(productId).strip()
        
        # 1. Try product mapping (JSON + Excel)
        if pid in self._productMap:
            return self._productMap[pid]
        
        # 2. Try video title (often contains actual product name)
        if videoTitle and str(videoTitle) not in ['-', 'None', 'nan', '', 'NaN']:
            title = str(videoTitle).strip()
            # Video titles with Nusseyba are likely product names
            if len(title) > 10 and 'Nusseyba' in title:
                return title[:80]
        
        # 3. Use campaign name (for discontinued/old products)
        if campaignName and str(campaignName) not in ['-', 'None', 'nan', '', 'NaN']:
            campaign = str(campaignName).strip()
            # Mark as discontinued/unmapped
            return f"[Discontinued] {campaign[:55]}"
        
        # 4. Last resort: Product ID suffix
        return f"[Unknown] Product-{pid[-8:]}" if len(pid) > 8 else f"[Unknown] Product-{pid}"
    
    def getCategory(self, productId: str) -> str:
        """Get product category if available."""
        self.load()
        return self._categoryMap.get(str(productId).strip(), "")
    
    def isDiscontinued(self, productId: str) -> bool:
        """
        Check if product is discontinued (not in Excel reference).
        These products are kept for historical analysis only.
        """
        self.load()
        return str(productId).strip() not in self._productMap
    
    def getDiscontinuedProducts(self, productIds: list) -> Set[str]:
        """Get set of discontinued product IDs from a list."""
        self.load()
        return {pid for pid in productIds 
                if str(pid).strip() not in self._productMap}
    
    def getActiveProducts(self, productIds: list) -> Set[str]:
        """Get set of active product IDs from a list."""
        self.load()
        return {pid for pid in productIds 
                if str(pid).strip() in self._productMap}
    
    def resolveMany(self, productIds: list, 
                    campaignNames: Dict[str, str] = None,
                    videoTitles: Dict[str, str] = None) -> Dict[str, str]:
        """
        Resolve multiple product IDs at once.
        
        Args:
            productIds: List of product IDs
            campaignNames: Dict mapping productId -> campaignName
            videoTitles: Dict mapping productId -> videoTitle
        
        Returns:
            Dict mapping productId -> resolved name
        """
        campaignNames = campaignNames or {}
        videoTitles = videoTitles or {}
        
        result = {}
        for pid in productIds:
            campaign = campaignNames.get(pid, "")
            video = videoTitles.get(pid, "")
            result[pid] = self.resolve(pid, campaign, video)
        
        return result
    
    def saveToJson(self, path: str = None) -> None:
        """Save current product map to JSON cache."""
        savePath = Path(path) if path else self.jsonCachePath
        if savePath:
            with open(savePath, 'w', encoding='utf-8') as f:
                json.dump(self._productMap, f, ensure_ascii=False, indent=2)
