#!/usr/bin/env python3
"""
Add WholesaleSettings table to all tenant databases
Non-destructive: Only adds table if not exists
"""

import sqlite3
import os
from pathlib import Path

# Schema for WholesaleSettings table
CREATE_TABLE_SQL = """
CREATE TABLE IF NOT EXISTS WholesaleSettings (
    id TEXT PRIMARY KEY,
    tenantId TEXT NOT NULL UNIQUE,
    platform TEXT DEFAULT 'shopee',
    adminFee INTEGER DEFAULT 1500,
    maxOrderTier3 INTEGER DEFAULT 1000,
    minOrder1 INTEGER DEFAULT 2,
    maxOrder1 INTEGER DEFAULT 3,
    createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    updatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
);
"""

CREATE_INDEX_SQL = """
CREATE UNIQUE INDEX IF NOT EXISTS WholesaleSettings_tenantId_key 
ON WholesaleSettings(tenantId);
"""


def add_wholesale_settings_table(db_path: str) -> bool:
    """Add WholesaleSettings table to a database if not exists."""
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # Check if table already exists
        cursor.execute("""
            SELECT name FROM sqlite_master 
            WHERE type='table' AND name='WholesaleSettings'
        """)
        
        if cursor.fetchone():
            print(f"  ⏭️  Table already exists, skipping")
            conn.close()
            return True
        
        # Create table
        cursor.execute(CREATE_TABLE_SQL)
        cursor.execute(CREATE_INDEX_SQL)
        conn.commit()
        conn.close()
        
        print(f"  ✅ WholesaleSettings table created")
        return True
        
    except Exception as e:
        print(f"  ❌ Error: {e}")
        return False


def main():
    # Database directory
    db_dir = Path(__file__).parent.parent / "config" / "databases"
    
    if not db_dir.exists():
        print(f"❌ Database directory not found: {db_dir}")
        return
    
    print("\n" + "=" * 60)
    print("🔧 ADDING WholesaleSettings TABLE TO TENANT DATABASES")
    print("=" * 60 + "\n")
    
    # Find all .db files (exclude jobs and system databases)
    db_files = list(db_dir.glob("*.db"))
    
    # Filter to only tenant databases (contain bertigamart or nusseyba)
    tenant_dbs = [
        db for db in db_files 
        if "bertigamart" in db.name or "nusseyba" in db.name or "developer" in db.name
    ]
    
    success_count = 0
    fail_count = 0
    
    for db_path in tenant_dbs:
        print(f"\n📁 Processing: {db_path.name}")
        
        if add_wholesale_settings_table(str(db_path)):
            success_count += 1
        else:
            fail_count += 1
    
    print("\n" + "=" * 60)
    print(f"✅ Complete: {success_count} success, {fail_count} failed")
    print("=" * 60 + "\n")


if __name__ == "__main__":
    main()
