# ⚠️ PENTING: WAJIB DI IMPLEMENTASIKAN OLEH AI. SAYA SEKARANG PAKE BACKEND GOLANG.

**JANGAN EDIT FILE INI TANPA IZIN EKSPLISIT. PELAJARI FILE INI DENGAN SAKSAMA SEBELUM MEMBUAT PERUBAHAN APAPUN.**

**TAHAP 1: PLANNING**

1. **Analisis Kode Existing**
   - Memeriksa seluruh struktur backend Golang
   - Menganalisis schema database Omni-PostgreSQL yang ada
   - Identifikasi pattern arsitektur yang digunakan

2. **Studi Dokumen dan Aturan**
   - Memahami constraint dan requirement sistem

3. **Analisis Integrasi Sistem**
   - Mempelajari flow backend-node + SQLite untuk referensi
   - Analisis SDK TikTok/Lazada Golang jika diperlukan
   - Memahami pemanggilan route dari Frontend

4. **Identifikasi Masalah**
   - Deteksi masalah naming convention (snake_case vs camelCase)
   - Identifikasi kode duplikat dan dead code
   - Analisis pelanggaran prinsip Clean Code dan DRY
   - Susun daftar masalah yang ditemukan lalu cari akar masalah masing-masing

5. **Perencanaan Solusi**
   - Membuat desain todolist yang komprehensif
   - Merencanakan struktur database optimal
   - Menyusun roadmap implementasi bertahap

**TAHAP 2: IMPLEMENTASI**

1. **Setup Environment**
   - Konfigurasi koneksi database dengan credentials (yumna/password123)
   - Persiapan development environment
   - Setup container Docker jika diperlukan

2. **Implementasi Core Features**
   - Bangun struktur todolist dengan layer architecture
   - Implementasi CRUD operations lengkap
   - Integrasi dengan sistem existing

3. **Penerapan Prinsip Kode**
   - Clean Code: Refaktor dan hapus duplikasi
   - DRY: Ekstrak logika umum ke utility functions
   - OOP: Gunakan interfaces dan struct dengan baik
   - SRP: Pisahkan concern ke file-file khusus
   - Maksimal 300 baris dalam 1 file

4. **Quality Control**
   - Patuhi batas 300 baris per file
   - Perbaiki semua error dan warning
   - Optimasi performance queries

5. **Container Management**
   - Jalankan build.bat jika ada masalah container
   - Gunakan Smart Build (build.py smart) atau Quick Apply
   - Tunggu proses build selesai sepenuhnya

**TAHAP 3: EVALUASI**

1. **Build & Initial Test**
   - Build aplikasi dan perbaiki error kompilasi
   - Uji fungsionalitas dasar todolist
   - Verifikasi koneksi database

2. **Perbaikan Berdasarkan Test**
   - Fix bug yang ditemukan selama testing
   - Optimasi kode berdasarkan hasil profiling
   - Perbaiki integration issues

3. **Final Cleanup**
   - Hapus debugging code dan komentar tidak perlu
   - Pastikan semua file <300 baris
   - Eliminasi duplikasi akhir

4. **Final Testing**
   - Uji regresi seluruh fungsionalitas
   - Verifikasi cleanup tidak merusak fitur
   - Test edge cases dan error scenarios

5. **Verifikasi Keberhasilan** = WAJIB SUCCESS TRUE
   - **Bukti 1**: Patikan data real
   - **Bukti 2**: Pastikan output script menunjukkan success

6. **Iterasi Proses**
   - Ulangi tahap jika ditemukan masalah kritis
   - Pastikan kode benar-benar clean dan functional
   - Final review sebelum delivery

user = yumna
pw = password123

**Aturan teknis penting**

- Alias : Tidak memakai ALIAS. Wajib perbaiki jika ada.
- Multi-tenant: tidak ada default tenant; semua akses DB harus pakai `context.Context` dan tenant_id wajib ada.
- Nama: kolom DB snake_case, JSON response snake_case, field Go PascalCase; mapping TableName wajib mengikuti `table_naming.go`.
- Database: schema per tenant; dilarang menambah/rename/hapus schema tanpa izin; tidak boleh raw DDL tanpa migration di `scripts/postgres/`.
- Larangan cache: jangan pakai Redis, in-memory cache, atau cache jenis apa pun; semua read langsung ke DB tenant.
- Batas file 300 baris; hindari duplikasi/dead code; handler hanya parsing/validasi/call service/format response.
- Testing: `cd backend && go build ./...` lalu `go test ./...` (dan lint jika ada) harus dijalankan oleh AI sebelum menyerahkan hasil.
- Security: jangan log data sensitif (password/token); gunakan structured logging; patuhi JWT/CORS/rate limit middleware.
- Env wajib: `GO_ENV`, `PORT`, `DB_DRIVER`, `PG_HOST`, `PG_PORT`, `PG_USER`, `PG_PASSWORD`, `PG_DATABASE`, `JWT_SECRET`, `ENCRYPTION_KEY`.
- Dokumentasi: tidak boleh membuat dokumentasi tanpa persetujuan user; jika wajib ada, simpan di `docs/` atau scripts testing.
- Tahapan kerja: semua tahap wajib dijalani AI; jangan minta user testing sebelum AI selesai menjalankan testing sendiri.
- Testing bisa pakai CURL/Pythonsudah
