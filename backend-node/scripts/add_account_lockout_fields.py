"""
Add account lockout fields to User table in all tenant databases
Multi-tenant aware - updates all databases in backend/config/databases/
"""
import sqlite3
import os
from pathlib import Path

# Database directory
DB_DIR = Path(__file__).parent.parent / "config" / "databases"

def add_account_lockout_fields(db_path):
    """Add account lockout fields to User table"""
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    # Check if User table exists
    cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='User'")
    if not cursor.fetchone():
        print(f"  ⚠️  No User table found, skipping")
        conn.close()
        return False
    
    # Check existing columns
    cursor.execute("PRAGMA table_info(User)")
    columns = {row[1] for row in cursor.fetchall()}
    
    fields_added = []
    
    # Add failedLoginAttempts if not exists
    if 'failedLoginAttempts' not in columns:
        cursor.execute("ALTER TABLE User ADD COLUMN failedLoginAttempts INTEGER DEFAULT 0")
        fields_added.append('failedLoginAttempts')
    
    # Add accountLockedUntil if not exists
    if 'accountLockedUntil' not in columns:
        cursor.execute("ALTER TABLE User ADD COLUMN accountLockedUntil TEXT")
        fields_added.append('accountLockedUntil')
    
    # Add lastFailedLogin if not exists
    if 'lastFailedLogin' not in columns:
        cursor.execute("ALTER TABLE User ADD COLUMN lastFailedLogin TEXT")
        fields_added.append('lastFailedLogin')
    
    conn.commit()
    conn.close()
    
    if fields_added:
        print(f"  ✅ Added fields: {', '.join(fields_added)}")
        return True
    else:
        print(f"  ✓  Fields already exist")
        return False

def main():
    print("🔐 Adding Account Lockout Fields to User Table")
    print("=" * 60)
    
    if not DB_DIR.exists():
        print(f"❌ Database directory not found: {DB_DIR}")
        return
    
    # Find all .db files
    db_files = list(DB_DIR.glob("*.db"))
    
    if not db_files:
        print(f"❌ No database files found in {DB_DIR}")
        return
    
    print(f"Found {len(db_files)} database(s)\n")
    
    updated_count = 0
    skipped_count = 0
    
    for db_file in sorted(db_files):
        # Skip journal files
        if db_file.suffix in ['.db-shm', '.db-wal']:
            continue
        
        print(f"📁 {db_file.name}")
        
        try:
            if add_account_lockout_fields(db_file):
                updated_count += 1
            else:
                skipped_count += 1
        except Exception as e:
            print(f"  ❌ Error: {e}")
            skipped_count += 1
        
        print()
    
    print("=" * 60)
    print(f"✅ Updated: {updated_count} database(s)")
    print(f"⚠️  Skipped: {skipped_count} database(s)")
    print("\n✅ Migration complete!")

if __name__ == "__main__":
    main()
