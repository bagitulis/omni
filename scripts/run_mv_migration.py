#!/usr/bin/env python3
"""
Run Materialized Views Migration for Analytics
"""
import subprocess
import sys

def main():
    sql_file = r'C:\Users\PC\Desktop\Project\omni\scripts\postgres\create-analytics-materialized-views.sql'
    
    print("Reading SQL file...")
    with open(sql_file, 'r', encoding='utf-8') as f:
        sql_content = f.read()
    
    print(f"SQL file size: {len(sql_content)} bytes")
    print("Executing migration via docker...")
    
    proc = subprocess.run(
        ['docker', 'exec', '-i', 'omni-postgres', 'psql', '-U', 'omni', '-d', 'omni_main'],
        input=sql_content,
        capture_output=True,
        text=True,
        timeout=300
    )
    
    if proc.stdout:
        print("=== STDOUT ===")
        print(proc.stdout)
    
    if proc.stderr:
        print("=== STDERR ===")
        print(proc.stderr)
    
    print(f"\nReturn code: {proc.returncode}")
    
    if proc.returncode == 0:
        print("\n[SUCCESS] Migration completed!")
    else:
        print("\n[ERROR] Migration failed!")
        sys.exit(1)

if __name__ == '__main__':
    main()
