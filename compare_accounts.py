#!/usr/bin/env python3
"""
Script untuk membandingkan email dari enowxai accounts list dengan tes.txt
dan menyimpan email yang belum ada di test_cek.txt
"""

import subprocess
import re

def get_enowxai_accounts():
    """Ambil list email dari enowxai accounts list"""
    try:
        result = subprocess.run(['enowxai', 'accounts', 'list'], 
                              capture_output=True, text=True)
        if result.returncode != 0:
            print(f"Error running enowxai command: {result.stderr}")
            return set()
        
        # Parse output dan ambil email yang unik
        emails = set()
        for line in result.stdout.split('\n'):
            # Skip header dan separator
            if line.startswith('EMAIL') or line.startswith('-') or not line.strip():
                continue
            if 'Total:' in line:
                break
            
            # Extract email (first column)
            parts = line.split()
            if parts and '@enxapi.site' in parts[0]:
                emails.add(parts[0])
        
        return emails
    except Exception as e:
        print(f"Error: {e}")
        return set()

def get_tes_accounts(filepath):
    """Ambil list email dari tes.txt"""
    emails = set()
    try:
        with open(filepath, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                # Format: number: email:password
                match = re.search(r'([a-zA-Z0-9]+@enxapi\.site)', line)
                if match:
                    emails.add(match.group(1))
        return emails
    except FileNotFoundError:
        print(f"File not found: {filepath}")
        return set()

def main():
    print("Mengambil data dari enowxai accounts list...")
    enowxai_emails = get_enowxai_accounts()
    
    print("Membaca tes.txt...")
    tes_emails = get_tes_accounts('D:\\Project\\omni\\tes.txt')
    
    # Cari email yang ada di tes.txt tapi tidak ada di enowxai
    missing_in_enowxai = tes_emails - enowxai_emails
    
    # Baca tes.txt untuk ambil password
    tes_data = {}
    try:
        with open('D:\\Project\\omni\\tes.txt', 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                match = re.search(r'([a-zA-Z0-9]+@enxapi\.site):(\S+)', line)
                if match:
                    email, password = match.groups()
                    tes_data[email] = password
    except:
        pass
    
    # Simpan ke test_cek.txt
    output_file = 'D:\\Project\\omni\\test_cek.txt'
    try:
        with open(output_file, 'w') as f:
            for email in sorted(missing_in_enowxai):
                password = tes_data.get(email, 'qwertyui')
                f.write(f"{email}:{password}\n")
        
        print(f"Total email yang belum ada di enowxai: {len(missing_in_enowxai)}")
        print(f"Hasil disimpan ke: {output_file}")
        
        # Eksekusi enowxai accounts add untuk menambahkan akun yang belum ada
        if missing_in_enowxai:
            print("\n" + "="*60)
            print("Menambahkan akun ke enowxai...")
            print("="*60 + "\n")
            subprocess.run(['enowxai', 'accounts', 'add', output_file])
        else:
            print("\nTidak ada akun baru yang perlu ditambahkan.")
                
    except Exception as e:
        print(f"Error writing to file: {e}")

if __name__ == '__main__':
    main()
