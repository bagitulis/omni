"""
Test Script untuk MCP GitHub
Memastikan semua fitur berjalan dengan benar
"""
import subprocess
import json
import sys
from pathlib import Path

def get_sheets_service():
    """Get Google Sheets service"""
    from google.oauth2.service_account import Credentials
    from googleapiclient.discovery import build
    
    config_path = Path(__file__).parent.parent.parent / "backend/config/static/google/bertigamart1-809b1d8a9ef5.json"
    
    creds = Credentials.from_service_account_file(
        str(config_path),
        scopes=["https://www.googleapis.com/auth/spreadsheets"]
    )
    return build("sheets", "v4", credentials=creds)

SPREADSHEET_ID = "1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA"

def test_connection():
    """Test Google Sheets connection"""
    service = get_sheets_service()
    result = service.spreadsheets().values().get(
        spreadsheetId=SPREADSHEET_ID,
        range="Main!A2:G5"
    ).execute()
    
    print("\n✅ Connection: SUCCESS")
    print(f"   Read {len(result.get('values', []))} rows")
    return True

def test_status_filter():
    """Test filtering by status"""
    service = get_sheets_service()
    result = service.spreadsheets().values().get(
        spreadsheetId=SPREADSHEET_ID,
        range="Main!A:G"
    ).execute()
    
    values = result.get('values', [])
    status_count = {}
    for row in values[2:]:
        if len(row) > 5:
            status = row[5]
            status_count[status] = status_count.get(status, 0) + 1
    
    print("\n✅ Status Filter: SUCCESS")
    print(f"   Distribution: {status_count}")
    return True

def test_code_backup():
    """Test reading backup codes"""
    service = get_sheets_service()
    result = service.spreadsheets().values().get(
        spreadsheetId=SPREADSHEET_ID,
        range="Code_Backup!A:Z"
    ).execute()
    
    values = result.get('values', [])
    emails = values[0] if values else []
    
    print("\n✅ Code Backup: SUCCESS")
    print(f"   {len(emails)} email columns")
    
    test_email = "salmankangchaerul@gmail.com"
    if test_email in emails:
        col_idx = emails.index(test_email)
        codes = [row[col_idx] for row in values[1:] if len(row) > col_idx and row[col_idx]]
        print(f"   Sample: {test_email} has {len(codes)} codes")
    return True

def test_rolling_credentials():
    """Test credential rolling mechanism"""
    config_dir = Path(__file__).parent.parent.parent / "backend/config/static/google"
    cred_files = list(config_dir.glob("*.json"))
    
    print(f"\n✅ Rolling Credentials: {len(cred_files)} files available")
    for f in cred_files:
        print(f"   - {f.name}")
    return len(cred_files) > 0

def main():
    print("="*60)
    print("MCP GitHub - Full Test Suite")
    print("="*60)
    
    tests = [
        ("Connection", test_connection),
        ("Status Filter", test_status_filter),
        ("Code Backup", test_code_backup),
        ("Rolling Credentials", test_rolling_credentials),
    ]
    
    passed = failed = 0
    for name, fn in tests:
        try:
            if fn(): passed += 1
        except Exception as e:
            print(f"\n❌ {name}: FAILED - {e}")
            failed += 1
    
    print("\n" + "="*60)
    print(f"Results: {passed} passed, {failed} failed")
    print("="*60)
    return failed == 0

if __name__ == "__main__":
    sys.exit(0 if main() else 1)
