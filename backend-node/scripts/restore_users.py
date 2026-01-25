"""
Restore Users Script
Restore essential users after accidental database reset

Per AGENTS.MD: Use Python for database operations
"""

import sqlite3
import os
import hashlib
import secrets
from pathlib import Path
from datetime import datetime

# Database paths
BACKEND_DIR = Path(__file__).parent.parent
DATABASES_DIR = BACKEND_DIR / "config" / "databases"

# Users to ensure exist
USERS = {
    "yumna_bertigamart.db": [
        {
            "id": "cmjnyo05e0000hb9l1lhvf1oa",
            "username": "yumna",
            "password": "$2a$10$rKN3zYvZj.VJWs5VhS7mVOoZlXdAqLqR4mHq5ZdB6jV2mJLq5.Iqe",  # password123
            "role": "owner",
        },
        {
            "id": "tester001",
            "username": "tester",
            "password": "$2a$10$rKN3zYvZj.VJWs5VhS7mVOoZlXdAqLqR4mHq5ZdB6jV2mJLq5.Iqe",  # password123
            "role": "tester",
        },
    ],
    "tika_nusseyba.db": [
        {
            "id": "tika001",
            "username": "tika",
            "password": "$2a$10$rKN3zYvZj.VJWs5VhS7mVOoZlXdAqLqR4mHq5ZdB6jV2mJLq5.Iqe",  # password123
            "role": "owner",
        },
    ],
}


def user_exists(cursor, username: str) -> bool:
    """Check if user exists"""
    cursor.execute("SELECT id FROM User WHERE username = ?", (username,))
    return cursor.fetchone() is not None


def ensure_users(db_path: str, users: list) -> int:
    """Ensure users exist in database"""
    db_name = os.path.basename(db_path)
    print(f"\n📁 Processing: {db_name}")
    
    if not os.path.exists(db_path):
        print(f"   ❌ Database not found")
        return 0
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    created = 0
    
    try:
        for user in users:
            if user_exists(cursor, user["username"]):
                print(f"   ℹ️  User '{user['username']}' already exists")
                continue
            
            # Password is already bcrypt hashed
            hashed = user["password"]
            now = datetime.utcnow().isoformat() + "Z"
            
            cursor.execute("""
                INSERT INTO User (id, username, password, role, createdAt, updatedAt)
                VALUES (?, ?, ?, ?, ?, ?)
            """, (user["id"], user["username"], hashed, user["role"], now, now))
            
            print(f"   ✅ Created user: {user['username']} (role: {user['role']})")
            created += 1
        
        conn.commit()
        return created
        
    except Exception as e:
        print(f"   ❌ Error: {e}")
        conn.rollback()
        return 0
    finally:
        conn.close()


def main():
    print("=" * 50)
    print("Restore Users Script")
    print("=" * 50)
    
    total_created = 0
    
    for db_name, users in USERS.items():
        db_path = DATABASES_DIR / db_name
        created = ensure_users(str(db_path), users)
        total_created += created
    
    print("\n" + "=" * 50)
    print(f"Complete: {total_created} users created")
    print("=" * 50)


if __name__ == "__main__":
    main()
