# Implementation Plan — FINAL (Verified & Approved)

## Keputusan Desain (Hasil Re-evaluasi)

| Keputusan | Pilihan | Alasan |
|-----------|---------|--------|
| Real-time method | **Polling 30 detik** | Tidak ada SSE/WebSocket infrastructure di codebase |
| Notification storage | **Database (tabel `notifications`)** | Persist across devices, tidak hilang saat clear cache |
| Notification types | **`notification_types.go` konstanta global** | Konsistensi backend-frontend |
| Toast behavior | **Hybrid**: toast + simpan ke DB | Muncul instan + tersimpan di panel |
| Read state | **Tetap muncul tapi meredup** | Seperti Facebook/Instagram |
| Retention | **Setting UI**: 7/30/90 hari atau manual | Per-tenant setting |
| Audio | **Visual only** | Tidak ada suara |

## Urutan Eksekusi

1. Backend: Model + Migration (`notification.go`, `analytics_cache_model.go`, `notification_types.go`)
2. Backend: Notification service + handler + routes
3. Backend: Fix `cache_service.go` (MV refresh background context + relispopulated check)
4. Backend: Fix `inventory_settings_service.go` (atomic upsert)
5. Backend: Fix `master_product_repository_links.go` + `import_service_shopee.go` (ON CONFLICT upsert)
6. Backend: Job system (zombie recovery, atomic claim, recordHistory fix, push notif)
7. Frontend: Polling hook + notificationStore update + NotificationDropdown enhancement
8. Frontend: Settings > Notifications tab
9. Frontend: All toast→notify conversions (50+ files)
10. Build + Test verification









# implementation_plan.md (Versi Ultimate)
## 🎯 Target Akhir
Sistem aplikasi yang stabil (bebas PostgreSQL error), background job yang efisien (paralel), dan sistem notifikasi real-time ala Medsos (Facebook/Instagram) yang masuk ke database.
## 🛠️ Proposed Changes
---
### [LAYER 1] Database & Backend Stability (Anti-Error)
**Target**: Membersihkan semua error di `postgress.txt`.
#### 1.1 Tabel `analytics_cache_metadata`
*   **Masalah**: Error "table missing" di log.
*   **Solusi**: Buat model GORM dan daftarkan ke sistem AutoMigrate agar tabel dibuat otomatis.
#### 1.2 Materialized View Smart Refresh
*   **Masalah**: Dashboard analytics kosong (blank) setelah server restart karena MV tidak terisi.
*   **Solusi**: Tambahkan cek status `relispopulated` sebelum refresh dan jalankan di background agar tidak terputus request user.
#### 1.3 Atomic Upsert Pattern
*   **Masalah**: Error duplikat di `inventory_settings` dan `platform_links`.
*   **Solusi**: Ganti pola *baca-baru-tulis* (SELECT-then-INSERT) menjadi satu query **Atomic Upsert** (`ON CONFLICT`).
---
### [LAYER 2] Job System Optimization (Parallelism)
**Target**: Menjamin tenant tidak saling menunggu dan job tersimpan rapi.
#### 2.1 Perbaikan MultiTenantExecutor
*   **Masalah**: Saat ini tenant antri satu-satu.
*   **Solusi**: Implementasi **Worker Pool/Goroutine** per tenant agar proses sync berjalan paralel (serempak).
#### 2.2 Zombie Job Recovery
*   **Masalah**: Job yang "nyangkut" status 'running' setelah server mati.
*   **Solusi**: Saat startup, otomatis reset job lama yang menggantung menjadi status 'failed'.
#### 2.3 Kamus Notifikasi (`notifications_types.go`) [NEW]
*   **Tujuan**: Membuat satu standar kategori (Sync, Order, Product, Inventory, System, Auth, Export) untuk seluruh aplikasi.
---
### [LAYER 3] Facebook-Style Notification System
**Target**: Notifikasi real-time, masuk DB, dan ada pengaturan masa simpan (Retention).
#### 3.1 Backend: Real-Time SSE (Server-Sent Events)
*   **Fitur**: Membuat "pipa" komunikasi langsung dari server ke browser agar notif muncul instan tanpa refresh.
*   **Event**: Dipicu setiap kali Background Job selesai atau ada aksi penting lainnya.
#### 3.2 Backend: Notification Service & API
*   **Tabel `notifications`**: Menyimpan permanen notif di DB (id, type, category, title, message, is_read, action_url).
*   **Retention Worker**: Proses harian untuk hapus notif lama sesuai setting user (7/30 hari).
#### 3.3 Frontend: Perombakan Total `notify.x()`
*   **Audit 50+ Hook & Page**: Mengganti semua `toast` yang ephemeral (hilang) menjadi `notify` yang masuk ke database.
*   **UI Bell Enhancement**: Badge angka merah, hover list ala Facebook, dan indikator kategori.
*   **Settings UI**: Tab baru untuk mengatur berapa lama notifikasi disimpan.
---
## 🧪 Rencana Verifikasi
1.  Check `postgress.txt` kembali: Harusnya tidak ada error baru muncul.
2.  Test Background Job: Jalankan sync di 2 tenant berbeda serentak (pastikan berjalan paralel).
3.  Test Medsos Notif: Jalankan job, matikan browser, buka lagi → Notif harus tetap ada di daftar.
4.  Test Retention: Set ke "1 hari" dan pastikan notif lama terhapus secara terjadwal.
---
## ❓ Open Questions
*   **Audio**: Apakah mau ada suara notifikasi kustom atau cukup visual saja?
*   **Mode Medsos**: Apakah notif "Sudah Dibaca" harus tetap muncul (seperti FB) atau otomatis hilang dari list? (Default: Tetap muncul tapi meredup).



Pertanyaan Terakhir: Apakah Anda ingin saya membuat satu file notifications_types.go global agar kategori notifikasi (Sync, Order, dll) konsisten di seluruh aplikasi? = visual aja
*   **Mode Medsos**: Apakah notif "Sudah Dibaca" harus tetap muncul (seperti FB) atau otomatis hilang dari list? (Default: Tetap muncul tapi meredup). = ya tetap muncul tapi meredup


silahkan reevaluasi ulang plan dan edgecase lagi secara mendalam cek semua referensi2 yang terkait dan juga verifikasi ulang, ika ada pertanyaan silahkan pertanyakan dulu.
 apabila suda sudah yakin dan sangat matang baru dieksekusi dengan catatan AI wajib yakin dan berhati2 . 