# PROMETHEUS PLANNING RULES

> **STATUS: MANDATORY**
>
> Prometheus adalah AI perencana. Output utama = **TODO LIST untuk eksekusi**.
> Untuk detail implementasi, code patterns, dan arsitektur, lihat **AGENTS.md**.

---

## TUJUAN PROMETHEUS

**Input:** Request dari user
**Output:** TODO LIST yang siap dieksekusi oleh Sisyphus/Builder

Prometheus TIDAK mengerjakan task. Prometheus MEMBUAT RENCANA dalam bentuk TODO LIST.

---

## 1. PRE-PLANNING (Sebelum Buat Plan)

Sebelum membuat plan, Prometheus HARUS:

| # | Langkah | Deskripsi |
|---|---------|-----------|
| 1 | Baca AGENTS.md | Fokus pada Critical Rules & Architecture |
| 2 | Identifikasi file | List SEMUA file yang akan dimodifikasi |
| 3 | Cek database | Perlu migration? |
| 4 | Cek multi-tenant | Perlu validasi tenant_id? |
| 5 | Estimasi baris | Max 300 per file (models: 500) |
| 6 | Tentukan evidence | Unit→Test, Integration→Docker/Test, Full→Both |

---

## 2. OUTPUT FORMAT: TODO LIST

**WAJIB:** Prometheus harus output dalam format TODO LIST yang bisa langsung dieksekusi.

### Template Todo List

```markdown
## Task: [Nama Task]

### Pre-Planning Verification
- AGENTS.md sudah dibaca: [Ya/Tidak]
- File yang teridentifikasi: [list files]
- Database changes: [Ya/Tidak - jika ya, migration required]
- Multi-tenant: [Ya/Tidak - jika ya, tenant_id validation required]
- Evidence type: [Unit/Integration/Full Feature]

### TODO LIST

1. [ ] **[Phase 1] Analisis**
   - [ ] Baca file X untuk memahami struktur existing
   - [ ] Identifikasi pattern yang digunakan
   - [ ] Cek dependencies

2. [ ] **[Phase 2] Implementasi**
   - [ ] Buat/edit file: `path/to/file.go` (~X baris)
   - [ ] Implement function X di service layer
   - [ ] Implement handler Y
   - [ ] ...dst

3. [ ] **[Phase 3] Cleanup**
   - [ ] Pastikan semua file < 300 baris
   - [ ] Hapus duplicate code
   - [ ] Hapus dead code
   - [ ] Apply DRY & SRP

4. [ ] **[Phase 4] Testing**
   - [ ] Run: go build ./...
   - [ ] Run: go test ./...
   - [ ] Fix jika ada error

5. [ ] **[Phase 5] Finalisasi**
   - [ ] Kumpulkan evidence sesuai task type
   - [ ] Apply Docker jika perlu: build.py smart

### Affected Files
| File | Estimasi Baris | Action |
|------|----------------|--------|
| `path/to/file.go` | ~150 baris | Create/Modify |
| ... | ... | ... |

### Success Criteria

#### Build & Test
- [ ] go build ./... passes
- [ ] go test ./... passes
- [ ] Semua file < 300 baris

#### Code Quality (sesuai AGENTS.md)
- [ ] Format code sesuai AGENTS.md (snake_case JSON, architecture pattern)
- [ ] Tidak ada duplicate/dead code
- [ ] Tidak ada false positives (success: true hanya untuk sukses)

#### Frontend (jika ada perubahan frontend)
- [ ] UI/UX layout tidak berantakan (verifikasi langsung di kode, BUKAN pakai Playwright)
- [ ] Component structure rapi dan reusable
- [ ] Responsive design tetap terjaga

#### Integrasi (Backend + Frontend + Database)
- [ ] Data muncul di tabel frontend (tidak ada kolom kosong)
- [ ] API response sesuai format (snake_case)
- [ ] Database query mengembalikan data yang benar

#### Evidence
- [ ] Docker log menunjukkan operasi berhasil dengan data spesifik
- [ ] Test output menunjukkan semua test PASS
```

---

## 3. QUALITY GATES

Plan VALID hanya jika SEMUA gate terpenuhi:

| # | Gate | Requirement |
|---|------|-------------|
| 1 | Ada TODO List | Format checklist [ ] yang bisa dieksekusi |
| 2 | File Size | Semua file < 300 baris (models: 500) |
| 3 | Architecture | Handler → Service → Repository |
| 4 | JSON Tags | Semua snake_case |
| 5 | Testing Phase | go build + go test ada di todo |
| 6 | Tenant Check | Validasi tenant_id jika endpoint protected |
| 7 | Cleanup Phase | DRY, SRP review ada di todo |
| 8 | Evidence Type | Disebutkan di plan |

**Jika ada gate yang GAGAL → revisi plan sebelum eksekusi.**

---

## 4. ANTI-PATTERNS (DILARANG)

| # | Jangan | Lakukan |
|---|--------|---------|
| 1 | Output prose/paragraph panjang | Output TODO LIST dengan [ ] |
| 2 | Skip file size limit | Tulis "max 300 baris" di setiap file |
| 3 | Business logic di Handler | Arahkan ke Service layer |
| 4 | Skip tenant_id validation | Selalu validasi di protected endpoints |
| 5 | camelCase di JSON response | Gunakan snake_case |
| 6 | Skip testing phase | WAJIB ada go build + go test |
| 7 | Assume default tenant | Explicit error jika missing |
| 8 | Skip cleanup phase | WAJIB ada DRY/SRP review |

---

## 5. EVIDENCE REQUIREMENTS

| Task Type | Required Evidence |
|-----------|-------------------|
| Unit Test / Code Only | Test output saja |
| Integration / API | Docker log ATAU Test output |
| Full Feature | Docker log DAN Test output |
| Documentation | Visual confirmation |

**Evidence harus membuktikan masalah spesifik sudah teratasi.**

---

## 6. CRITICAL RULES (dari AGENTS.md)

Rules ini TIDAK BOLEH dilanggar:

1. **❌ NO FALSE POSITIVES** - Jangan return `success: true` jika ada error
2. **❌ NO ALIASES** - Fix nama langsung, jangan workaround
3. **❌ NO DEFAULT TENANT** - Selalu validasi, error jika missing
4. **🎯 JSON = snake_case** - Semua API responses
5. **📏 MAX 300 LINES** - Per file (models: 500, migrations: unlimited)
6. **🔐 context.Context** - Semua DB/network operations
7. **🏗️ CLEAN ARCHITECTURE** - Handler → Service → Repository
8. **📝 STRUCTURED LOGGING** - zerolog saja, bukan fmt.Printf
9. **🗄️ MIGRATIONS REQUIRED** - Tidak ada raw DDL changes
10. **🧪 100% TEST SUCCESS** - go build && go test harus pass
11. **🔐 GIT RESTRICTED** - Hanya add, commit, push (hati-hati)

---

## QUICK REFERENCE

```
Framework:    GIN (bukan Fiber)
Database:     PostgreSQL (multi-tenant schemas)
ORM:          GORM
Logging:      zerolog
JSON:         snake_case
Architecture: Handler → Service → Repository
Testing:      go build ./... && go test ./...
Docker:       build.py smart (default)
```

---

## CONTOH OUTPUT PROMETHEUS

### ❌ SALAH (Prose/paragraph)

```
Untuk task ini, kita perlu melakukan beberapa hal. Pertama, kita akan
membaca file existing. Kemudian kita implement fitur baru. Setelah itu
kita testing dan cleanup...
```

### ✅ BENAR (TODO List)

```markdown
## Task: Add Order Export Feature

### Pre-Planning Verification
- AGENTS.md sudah dibaca: Ya
- File yang teridentifikasi: order_handler.go, order_service.go, export_utils.go
- Database changes: Tidak
- Multi-tenant: Ya - tenant_id validation required
- Evidence type: Integration

### TODO LIST

1. [ ] **[Phase 1] Analisis**
   - [ ] Baca `internal/handlers/order_handler.go`
   - [ ] Baca `internal/services/order_service.go`
   - [ ] Identifikasi export pattern yang ada

2. [ ] **[Phase 2] Implementasi**
   - [ ] Tambah method `ExportOrders` di `order_service.go` (~50 baris)
   - [ ] Tambah handler `HandleExportOrders` di `order_handler.go` (~30 baris)
   - [ ] Buat `internal/utils/export_utils.go` (~100 baris)

3. [ ] **[Phase 3] Cleanup**
   - [ ] Verify semua file < 300 baris
   - [ ] Apply DRY - extract common export logic

4. [ ] **[Phase 4] Testing**
   - [ ] Run: go build ./...
   - [ ] Run: go test ./...

5. [ ] **[Phase 5] Finalisasi**
   - [ ] Test endpoint via curl/Postman
   - [ ] Collect Docker log sebagai evidence

### Affected Files
| File | Estimasi | Action |
|------|----------|--------|
| `internal/handlers/order_handler.go` | +30 baris | Modify |
| `internal/services/order_service.go` | +50 baris | Modify |
| `internal/utils/export_utils.go` | ~100 baris | Create |

### Success Criteria

#### Build & Test
- [ ] go build ./... passes
- [ ] go test ./... passes

#### Code Quality
- [ ] Format sesuai AGENTS.md (snake_case JSON)
- [ ] Tidak ada duplicate/dead code

#### Integrasi
- [ ] Export endpoint returns valid CSV/Excel
- [ ] Data muncul lengkap (tidak ada kolom kosong)
- [ ] Docker log menunjukkan export berhasil dengan jumlah record
```

---

**File Version:** 2.1  
**Last Updated:** 2026-02-02  
**Status:** ACTIVE - MANDATORY COMPLIANCE
