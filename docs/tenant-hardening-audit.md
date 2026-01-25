## Audit Isolasi Tenant (Go backend)
Baseline: AGENTS.md melarang default tenant dan mewajibkan context per-request. Node punya `tenantMiddleware` global dengan allowlist kecil (`/api/health`, `/api/status`, `/api/auth/*` publik, `/api/captcha`, `/api/webhooks`, `/api/platform-auth`, `/api/n8n`, `/api/order-manager`). Go sekarang hanya mengandalkan middleware per-route sehingga ada celah.

### 1) Titik lemah enforcement tenant di Go routes
- **Tidak ada global tenant guard**: `cmd/server/main.go` hanya pasang `CORS` + `Logger`. Coverage bergantung setiap group. Risiko: route baru tanpa `Tenant()` bisa bocor lintas tenant; berbeda dengan guard global di Node.
- **`/api/google/service-accounts/*` (List/Stats/Switch/DetectSheet)**: Hanya `Auth()` di `routes/google_routes.go`. Handler sudah baca `tenantID` (DetectSheet mengembalikan 401 jika kosong) dan Sheets harus tenant-scoped. Tambah `Tenant()` pada group agar selaras Node.
- **`/api/monitoring/metrics` & `/api/monitoring/health/detailed`**: Hanya `Auth()`. `MonitoringHandler` memanggil `h.db` tanpa set schema; jika `h.db` tenant connection, search_path yang aktif bisa salah (tenant bleed). Pilih kebijakan: (a) tandai sebagai system-only (pakai system DB eksplisit dan beri catatan), atau (b) tambahkan `Tenant()` dan set schema sebelum ping DB.
- **`/api/status`**: Endpoint publik (di Node masuk allowlist). Di Go bisa mengembalikan token status jika ada header `x-tenant-id` tanpa auth/tenant middleware; berpotensi bocor status token lintas tenant jika header dipalsukan. Solusi: jadikan health-only (hapus detail token) atau wajibkan `Auth()+Tenant()` untuk data token.

### 2) Area yang bergantung pada tenant dari caller tanpa jaring pengaman middleware
- **Protected auth routes** (`/api/auth/logout`, `change-password`, `me`, `tenants`, `switch-tenant`) hanya `Auth()`. Saat ini beroperasi di system schema; tetap dokumentasikan sebagai system-scoped agar tidak memakai tenant DB tanpa middleware di masa depan.
- **Google quota routes** memang publik; biarkan sebagai pengecualian resmi.

### 3) Langkah penguatan (urut prioritas)
1. **Pasang tenant middleware global** (mirror allowlist Node) di router; pertahankan `Tenant()` per-group saat transisi untuk hindari regresi. Ini menutup gap otomatis untuk route baru.  
2. **Tambah `Tenant()` pada `/api/google/service-accounts/*`** di `routes/google_routes.go` karena handler sudah expect `tenantID`.  
3. **Tetapkan kebijakan `/api/monitoring/*`**: pilih system-only (pakai system DB eksplisit + komentar) atau Tenant+schema.  
4. **Kunci `/api/status`**: opsi A – tetap publik tapi hanya health (tanpa token detail); opsi B – pisah `/api/health` (public) dan `/api/status` (Auth+Tenant) untuk token visibility.  
5. **Dokumentasikan allowlist path publik** di Go (samakan dengan Node TENANT_EXEMPT_PATHS) supaya endpoint baru tidak skip tenant secara tidak sengaja.

### 4) Cross-check singkat
- Node `tenantMiddleware` punya allowlist, Go belum ada guard global.
- Service-account routes Auth-only padahal handler butuh tenant.  
- Monitoring routes Auth-only dan menyentuh DB tanpa schema eksplisit.  
- `/api/status` publik dan bisa membeberkan token info jika header disuplai.

Menjalankan langkah di atas akan menyelaraskan Go backend dengan kontrak multi-tenant Node dan menutup peluang spoofing/kebocoran tenant.  
