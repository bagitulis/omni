#!/usr/bin/env python3
"""
Fix auto_functions_config schema - Remove tenant_id references
Since each tenant has separate database, tenant_id column is redundant
"""

import sqlite3
import os
from pathlib import Path

def fix_auto_functions_schema(db_path):
    """Remove tenant_id column from auto_functions_config if exists"""
    print(f"📂 Checking: {db_path}")
    
    if not os.path.exists(db_path):
        print(f"   ⚠️  Database not found, skipping")
        return
    
    try:
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()
        
        # Check if table exists
        cursor.execute("""
            SELECT name FROM sqlite_master 
            WHERE type='table' AND name='auto_functions_config'
        """)
        
        if not cursor.fetchone():
            print(f"   ℹ️  Table auto_functions_config doesn't exist, skipping")
            conn.close()
            return
        
        # Get current schema
        cursor.execute("PRAGMA table_info(auto_functions_config)")
        columns = cursor.fetchall()
        column_names = [col[1] for col in columns]
        
        print(f"   📋 Current columns: {', '.join(column_names)}")
        
        # Check if tenant_id exists
        if 'tenant_id' not in column_names:
            print(f"   ✅ Schema already correct (no tenant_id)")
            conn.close()
            return
        
        print(f"   🔧 Removing tenant_id column...")
        
        # SQLite doesn't support DROP COLUMN directly
        # Need to recreate table without tenant_id
        
        # 1. Create new table without tenant_id
        cursor.execute("""
            CREATE TABLE auto_functions_config_new (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL UNIQUE,
                enabled BOOLEAN DEFAULT 0,
                interval_minutes INTEGER DEFAULT 30,
                start_time TEXT,
                end_time TEXT,
                last_executed DATETIME,
                next_scheduled_execution DATETIME,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        """)
        
        # 2. Copy data (excluding tenant_id)
        cursor.execute("""
            INSERT INTO auto_functions_config_new 
            (id, name, enabled, interval_minutes, start_time, end_time, 
             last_executed, next_scheduled_execution, created_at, updated_at)
            SELECT id, name, enabled, interval_minutes, start_time, end_time,
                   last_executed, next_scheduled_execution, created_at, updated_at
            FROM auto_functions_config
        """)
        
        # 3. Drop old table
        cursor.execute("DROP TABLE auto_functions_config")
        
        # 4. Rename new table
        cursor.execute("ALTER TABLE auto_functions_config_new RENAME TO auto_functions_config")
        
        # 5. Recreate indexes
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_auto_functions_enabled 
            ON auto_functions_config(enabled)
        """)
        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_auto_functions_name 
            ON auto_functions_config(name)
        """)
        
        conn.commit()
        print(f"   ✅ Schema fixed successfully")
        
        # Verify
        cursor.execute("PRAGMA table_info(auto_functions_config)")
        new_columns = cursor.fetchall()
        new_column_names = [col[1] for col in new_columns]
        print(f"   📋 New columns: {', '.join(new_column_names)}")
        
        conn.close()
        
    except Exception as e:
        print(f"   ❌ Error: {e}")
        if 'conn' in locals():
            conn.rollback()
            conn.close()
        raise

def main():
    """Fix schema for all tenant databases"""
    print("🔧 Fixing auto_functions_config schema for all tenants\n")
    
    # Path to tenant databases
    backend_dir = Path(__file__).parent.parent
    db_dir = backend_dir / "config" / "databases"
    
    print(f"📁 Database directory: {db_dir}\n")
    
    if not db_dir.exists():
        print(f"❌ Database directory not found: {db_dir}")
        return
    
    # Find all tenant job databases
    job_dbs = list(db_dir.glob("*_jobs.db"))
    
    if not job_dbs:
        print("⚠️  No tenant job databases found")
        return
    
    print(f"Found {len(job_dbs)} tenant job databases:\n")
    
    success_count = 0
    error_count = 0
    
    for db_path in job_dbs:
        try:
            fix_auto_functions_schema(str(db_path))
            success_count += 1
        except Exception as e:
            print(f"   ❌ Failed: {e}")
            error_count += 1
        print()
    
    print("=" * 60)
    print(f"✅ Successfully fixed: {success_count}")
    print(f"❌ Failed: {error_count}")
    print(f"📊 Total: {len(job_dbs)}")
    print("=" * 60)

if __name__ == "__main__":
    main()
