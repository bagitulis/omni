Saya menggunakan backend Golang dengan database Omni-PostgreSQL. Buatkan todolist lengkap dengan implementasi detail. Kode yang banyak tidak masalah.

Informasi Kredensial:  
Username: yumna  
Password: password123

Panduan Wajib (Baca agents.md dengan ketat dan review setelah selesai):

1. Prinsip yang HARUS diterapkan:
   - Clean Code (hapus duplicate/dead code)
   - DRY (Don't Repeat Yourself)
   - OOP (Object-Oriented Programming)
   - SRP (Single Responsibility Principle)
2. Batas File: Maksimal 300 baris per file.
3. Quality Control: Perbaiki SEMUA error, unused code, duplikasi, warning.

PROSES PENGEMBANGAN (WAJIB URUTAN INI):

1. Analisis Kode & Database: Periksa semua kode existing dan schema Omni-PostgreSQL
2. Perbaikan Masalah Fungsional: Fix bug dan masalah utama
3. Build & Test Dasar: Build aplikasi dan uji fungsionalitas
4. Perbaikan Lanjut: Fix berdasarkan hasil test
5. CLEANUP AKHIR: Refaktor semua kode sesuai agents.md (SRP, DRY, OOP, <300 baris, hapus duplikasi)
6. Testing Final: Pastikan cleanup tidak merusak fungsionalitas
7. Iterasi: Ulangi sampai berhasil

Kriteria Keberhasilan (Dua Bukti Wajib):
Bukti 1: Log Docker yang menampilkan data valid spesifik (misal: tanggal terkait masalah, transaksi sukses).
Bukti 2: Output script dengan keterangan keberhasilan yang eksplisit terkait penyelesaian masalah.
Catatan: Bukti harus relevan langsung dengan inti masalah yang didiskusikan. Bukti running umum tanpa konteks spesifik tidak dianggap sah.

AI Harus: Pelajari flow backend-node + SQLite, SDK TikTok/Lazada Golang, dan pemanggilan route dari Frontend. Perhatikan juga masalah penggunaan snake_case dan camelCase.

---
