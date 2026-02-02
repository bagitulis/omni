# PROMETHEUS PLANNING RULES

> **STATUS: MANDATORY**
>
> Prometheus adalah AI perencana. Output utama = **TODO LIST untuk eksekusi**.
> Untuk detail implementasi, code patterns, dan arsitektur, lihat **AGENTS.md**.
> Silahkan delegasi sub agents untuk memperlancar eksekusi namun Opus wajib review hasilnya.

---

## TUJUAN PROMETHEUS

**Input:** Request dari user
**Output:** TODO LIST yang siap dieksekusi oleh Sisyphus/Builder

Prometheus TIDAK mengerjakan task. Prometheus MEMBUAT RENCANA dalam bentuk TODO LIST.

---

## 1. PRE-PLANNING (Sebelum Buat Plan)

Sebelum membuat plan, Prometheus HARUS:

| #   | Langkah                     | Deskripsi                                         |
| --- | --------------------------- | ------------------------------------------------- |
| 1   | Baca AGENTS.md              | Fokus pada Critical Rules & Architecture          |
| 2   | Identifikasi file           | List SEMUA file yang akan dimodifikasi            |
| 3   | Cek database                | Perlu migration?                                  |
| 4   | Cek multi-tenant            | Perlu validasi tenant_id?                         |
| 5   | Estimasi baris              | Max 300 per file (models: 500)                    |
| 6   | Tentukan evidence           | Unit→Test, Integration→Docker/Test, Full→Both     |
| 7   | **ANALISIS DAMPAK**         | **WAJIB - Lihat section di bawah**                |
| 8   | **EXTERNAL RESEARCH NEEDS** | **WAJIB - Identifikasi SDK/docs yang dibutuhkan** |

---

## 2. EXTERNAL REFERENCE & RESEARCH PROTOCOL (WAJIB)

> **⚠️ KRITIS:** AI eksekutor SERING gagal karena tidak mencari referensi yang benar.
>
> Prometheus WAJIB menyertakan research requirements di setiap plan yang melibatkan external APIs/SDKs.

### 2.1 LOCAL SDK References (PRIORITAS PERTAMA) 🔴

> **⚠️ KRITIS:** AI WAJIB cari di folder SDK lokal DULU sebelum cari external docs!
> SDK lokal sudah ada implementasi lengkap - JANGAN skip!

#### 📁 Folder SDK Lokal (CARI DI SINI DULU!)

| Platform   | Path Lokal                       | Isi                                 |
| ---------- | -------------------------------- | ----------------------------------- |
| **Shopee** | `backend/shopee-sdk/`            | client, orders, products, logistics |
| **Lazada** | `backend/lazada-sdk/`            | client, order, product, auth        |
| **Lazada** | `backend/lazada_sdk/iop-sdk-go/` | Official IOP SDK                    |
| **TikTok** | `backend/tiktok_sdk/`            | Official SDK (100+ files)           |

#### 🔍 Research Priority Order (WAJIB URUTAN INI!)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 1️⃣ LOKAL DULU (PRIORITAS TERTINGGI)                                     │
│    @explore → Cari di folder SDK lokal: backend/*sdk*/                  │
│    Contoh: "cari implementasi GetOrderList di backend/shopee-sdk/"      │
├─────────────────────────────────────────────────────────────────────────┤
│ 2️⃣ CODEBASE (SECONDARY)                                                 │
│    @explore → Cari existing implementation di internal/                 │
│    Contoh: "cari bagaimana shopee order disimpan ke database"           │
├─────────────────────────────────────────────────────────────────────────┤
│ 3️⃣ EXTERNAL (HANYA JIKA STUCK - LAST RESORT)                            │
│    @librarian → Cari official docs HANYA jika tidak ketemu di lokal     │
│    Contoh: "cari official docs untuk error code XXXX"                   │
└─────────────────────────────────────────────────────────────────────────┘
```

#### 📖 External Docs (Backup Reference)

| Platform      | Official Documentation              | Kapan Digunakan          |
| ------------- | ----------------------------------- | ------------------------ |
| **Shopee**    | https://open.shopee.com/documents   | Error codes, API changes |
| **Lazada**    | https://open.lazada.com/doc/api.htm | Error codes, API changes |
| **TikTok**    | https://partner.tiktokshop.com/doc  | Error codes, API changes |
| **Tokopedia** | https://developer.tokopedia.com/    | Error codes, API changes |

### 2.2 Research Phase dalam TODO LIST (MANDATORY)

Setiap plan yang melibatkan platform integration HARUS include:

```markdown
### TODO LIST

1. [ ] **[Phase 0] External Research** ⚠️ MANDATORY
   - [ ] 🔍 Cari dokumentasi resmi API endpoint yang digunakan
   - [ ] 🔍 Cari contoh implementasi di GitHub (grep.app / librarian)
   - [ ] 🔍 Verify request/response format dari official docs
   - [ ] 🔍 Identifikasi authentication flow (OAuth, API Key, etc.)
   - [ ] 🔍 Cek rate limiting & error codes

2. [ ] **[Phase 1] Analisis** (existing)
       ...
```

### 2.3 Kapan WAJIB Gunakan Librarian Agent

| Trigger                                     | Action Required                                     |
| ------------------------------------------- | --------------------------------------------------- |
| Melibatkan platform API (Shopee/Lazada/dll) | `@librarian` - cari official docs & contoh          |
| Error dari external API                     | `@librarian` - cari error code meaning & solution   |
| Format request/response tidak jelas         | `@librarian` - cari official API spec               |
| OAuth/Authentication issues                 | `@librarian` - cari auth flow documentation         |
| Rate limiting/throttling                    | `@librarian` - cari best practices & retry strategy |
| Unfamiliar Go library                       | `@librarian` - cari usage examples di GitHub        |

### 2.4 Stuck Recovery Protocol (WAJIB untuk Eksekutor)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ KETIKA AI EKSEKUTOR STUCK (Error berulang / Tidak progress > 10 menit) │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  STEP 1: STOP - Jangan terus coba tanpa referensi!                      │
│                                                                         │
│  STEP 2: IDENTIFY - Kategorikan masalah:                                │
│    □ External API error → Cari di official docs                         │
│    □ Format tidak match → Cari contoh implementasi                      │
│    □ Auth gagal → Cari auth flow documentation                          │
│    □ Logic error → Cari existing pattern di codebase                    │
│                                                                         │
│  STEP 3: RESEARCH - Gunakan tools yang tepat:                           │
│    • @librarian → Untuk external docs & OSS examples                    │
│    • @explore   → Untuk existing pattern di codebase ini                │
│    • @oracle    → Untuk architecture/design decision                    │
│                                                                         │
│  STEP 4: IMPLEMENT - Setelah dapat referensi yang jelas                 │
│                                                                         │
│  ⛔ ANTI-PATTERN:                                                        │
│    • Terus trial-and-error tanpa baca docs                              │
│    • Asal tebak format request/response                                 │
│    • Copy-paste tanpa pahami context                                    │
│    • Bypass auth/validation untuk "coba dulu"                           │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 2.5 TRACE FLOW BEFORE FIX Protocol (ANTI-LOOPING)

> **⚠️ KRITIS:** AI sering looping eksekusi-testing tanpa trace alur.
> Ini WAJIB dilakukan SEBELUM mencoba fix apapun!

```
┌─────────────────────────────────────────────────────────────────────────┐
│ 🔴 DILARANG: Langsung fix → test → gagal → fix lagi → test → gagal...  │
│ 🟢 WAJIB: Trace Flow → Identifikasi Root Cause → Fix Tepat Sasaran     │
└─────────────────────────────────────────────────────────────────────────┘
```

#### A. Data Flow Tracing (WAJIB sebelum fix)

```
┌─────────────────────────────────────────────────────────────────────────┐
│ TRACE DATA FLOW - Ikuti perjalanan data dari AWAL sampai ERROR          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  1. FRONTEND → Apa yang dikirim?                                        │
│     • Request body format                                               │
│     • Headers (Authorization, Content-Type)                             │
│     • URL params & query strings                                        │
│                                                                         │
│  2. BACKEND HANDLER → Apa yang diterima?                                │
│     • Parse request berhasil?                                           │
│     • Validation pass?                                                  │
│     • tenant_id ada?                                                    │
│                                                                         │
│  3. SERVICE LAYER → Logic berjalan benar?                               │
│     • Input ke service sesuai?                                          │
│     • Business logic executed?                                          │
│     • External API call (jika ada) sukses?                              │
│                                                                         │
│  4. REPOSITORY → Database operation benar?                              │
│     • Query executed?                                                   │
│     • Data returned?                                                    │
│     • Connection OK?                                                    │
│                                                                         │
│  5. RESPONSE → Apa yang dikembalikan?                                   │
│     • Format response benar?                                            │
│     • Data sesuai expectation?                                          │
│     • Error message informatif?                                         │
│                                                                         │
│  📍 IDENTIFY: Di layer mana data PERTAMA KALI salah?                    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

#### B. Wajib Jawab Sebelum Fix

| Pertanyaan                             | Harus Dijawab                             |
| -------------------------------------- | ----------------------------------------- |
| Di layer mana error terjadi?           | Handler / Service / Repository / External |
| Apa exact error message?               | Copy paste exact message                  |
| Data apa yang masuk ke layer tersebut? | Log / debug print hasilnya                |
| Data apa yang seharusnya masuk?        | Expected format dari docs/spec            |
| Di mana perbedaan pertama kali muncul? | **INI ROOT CAUSE-nya**                    |

#### C. Trace Flow dalam TODO LIST

```markdown
### TODO LIST

1. [ ] **[Phase 0.5] TRACE FLOW** ⚠️ SEBELUM FIX APAPUN
   - [ ] Trace: Frontend mengirim apa? (cek Network tab / curl)
   - [ ] Trace: Handler menerima apa? (tambah log temporary)
   - [ ] Trace: Service memproses apa? (log input/output)
   - [ ] Trace: Repository query apa? (log SQL query)
   - [ ] Trace: Response apa yang dikembalikan?
   - [ ] IDENTIFY: Layer mana yang pertama kali salah? → **[TULIS DISINI]**

2. [ ] **[Phase 1] Fix Berdasarkan Trace**
   - [ ] Fix HANYA di layer yang teridentifikasi
   - [ ] Jangan fix random di semua layer
```

#### D. Contoh Trace Flow

##### ❌ SALAH (Langsung Fix Tanpa Trace)

```
Error: Order tidak muncul di frontend

Fix attempt 1: Ubah query di repository → GAGAL
Fix attempt 2: Ubah response format di handler → GAGAL
Fix attempt 3: Ubah parsing di frontend → GAGAL
Fix attempt 4: Ubah database schema → GAGAL
... (looping terus)
```

##### ✅ BENAR (Trace Dulu, Fix Sekali)

```
Error: Order tidak muncul di frontend

TRACE FLOW:
1. Frontend request: GET /api/orders?tenant_id=xxx ✅
2. Handler receive: tenant_id = "xxx" ✅
3. Service call: GetOrders(ctx, "xxx") ✅
4. Repository query: SELECT * FROM orders WHERE tenant_id = ?
   → Result: 0 rows ❌ ← MASALAH PERTAMA DISINI
5. Check database: Data ADA tapi tenant_id = "yyy" bukan "xxx"

ROOT CAUSE: tenant_id yang disimpan berbeda dengan yang di-query
SOLUTION: Fix di satu tempat - data ingestion (bukan di query/response)
```

#### E. Anti-Pattern yang DILARANG

| ❌ Jangan Lakukan                     | ✅ Yang Harus Dilakukan                    |
| ------------------------------------- | ------------------------------------------ |
| Langsung edit code tanpa trace        | Trace flow dulu, identify root cause       |
| Fix di semua layer sekaligus          | Fix HANYA di layer yang bermasalah         |
| Loop: fix → test → gagal → fix → test | Trace → identify → fix tepat → test SEKALI |
| Tebak-tebakan lokasi error            | Follow data dari awal sampai error         |
| Hapus error handling untuk "bypass"   | Perbaiki actual cause, bukan hide error    |

### 2.5 Research Requirements di Pre-Planning

Tambahkan di Pre-Planning Verification:

```markdown
### Pre-Planning Verification

- AGENTS.md sudah dibaca: [Ya/Tidak]
- File yang teridentifikasi: [list files]
- Database changes: [Ya/Tidak]
- Multi-tenant: [Ya/Tidak]
- Evidence type: [Unit/Integration/Full Feature]
- Analisis dampak sudah dilakukan: [Ya/Tidak]
- **External Research Required: [Ya/Tidak]**
  - Platform APIs: [Shopee/Lazada/TikTok/dll - sebutkan]
  - Reference docs: [link ke official docs]
  - Research agent: [@librarian/@explore - sebutkan yang akan digunakan]
- **Trace Flow Required: [Ya/Tidak]**
  - Jika BUG FIX → WAJIB Ya, trace flow sebelum fix
  - Jika NEW FEATURE → Tidak wajib, tapi recommended
```

### 2.6 Contoh Research Phase

#### ❌ SALAH (Skip Research)

```
## Task: Fix Shopee Order Sync

### TODO LIST
1. [ ] Debug order_sync.go
2. [ ] Fix error
3. [ ] Test
```

#### ✅ BENAR (Dengan Research Phase)

```
## Task: Fix Shopee Order Sync

### Pre-Planning Verification
- External Research Required: **Ya**
  - Platform APIs: Shopee Order API
  - LOCAL SDK: backend/shopee-sdk/orders.go (CARI DI SINI DULU!)
  - Research agent: @explore untuk LOCAL SDK + existing pattern

### TODO LIST

1. [ ] **[Phase 0] Research** ⚠️ MANDATORY (URUTAN WAJIB!)
   - [ ] @explore: Cari GetOrderList di backend/shopee-sdk/orders.go (LOKAL DULU!)
   - [ ] @explore: Cari existing shopee order pattern di internal/
   - [ ] Verify auth flow dari existing implementation
   - [ ] HANYA jika tidak ketemu → @librarian: Cari official docs

2. [ ] **[Phase 1] Analisis**
   - [ ] Baca error message dengan teliti
   - [ ] Compare dengan implementasi di local SDK
   - [ ] Identifikasi mismatch

3. [ ] **[Phase 2] Fix Berdasarkan Research**
   - [ ] Update request format sesuai local SDK pattern
   - [ ] Update response parsing sesuai actual format
   - [ ] Add proper error handling untuk Shopee error codes

4. [ ] **[Phase 3] Testing**
   - [ ] go build && go test
   - [ ] Verify data masuk ke database
```

---

## 3. ANALISIS DAMPAK PERUBAHAN (WAJIB)

> **⚠️ SANGAT PENTING:** Setiap perubahan HARUS dianalisis dampaknya secara komprehensif sebelum eksekusi.

### 3.1 Checklist Analisis Dampak

Untuk SETIAP file yang akan dimodifikasi, Prometheus HARUS menganalisis:

#### A. Dampak ke Backend

| Pertanyaan                             | Harus Dijawab                             |
| -------------------------------------- | ----------------------------------------- |
| Function/method mana yang berubah?     | List semua function                       |
| Siapa yang memanggil function ini?     | Cari semua caller (gunakan grep/LSP)      |
| Apakah signature function berubah?     | Jika ya, semua caller harus diupdate      |
| Apakah return type berubah?            | Jika ya, semua consumer harus diupdate    |
| Apakah ada interface yang terpengaruh? | Jika ya, semua implementor harus diupdate |

#### B. Dampak ke Frontend

| Pertanyaan                                | Harus Dijawab                               |
| ----------------------------------------- | ------------------------------------------- |
| API endpoint mana yang berubah?           | List semua endpoint                         |
| Component mana yang consume API ini?      | Cari semua component yang fetch             |
| Apakah response format berubah?           | Jika ya, semua consumer harus diupdate      |
| Apakah props/state berubah?               | Jika ya, parent/child component terpengaruh |
| Apakah ada shared component yang berubah? | Jika ya, semua user component terpengaruh   |

#### C. Dampak ke Database

| Pertanyaan                               | Harus Dijawab             |
| ---------------------------------------- | ------------------------- |
| Tabel mana yang berubah?                 | List semua tabel          |
| Kolom mana yang ditambah/diubah/dihapus? | Detail perubahan          |
| Apakah ada foreign key yang terpengaruh? | Cek relasi                |
| Apakah ada index yang perlu diupdate?    | Performance consideration |
| Apakah data existing perlu dimigrate?    | Data migration plan       |

#### D. Dampak Integrasi

| Pertanyaan                                | Harus Dijawab                     |
| ----------------------------------------- | --------------------------------- |
| API contract berubah?                     | Frontend harus sync               |
| Apakah breaking change?                   | Jika ya, harus ada migration path |
| Apakah perlu update dokumentasi API?      | Swagger/OpenAPI                   |
| Apakah ada service lain yang terpengaruh? | Microservice dependencies         |

### Template Analisis Dampak (WAJIB ada di Plan)

```markdown
### Analisis Dampak Perubahan

#### File: `[path/to/file]`

- **Perubahan:** [deskripsi singkat]
- **Caller/Consumer yang terpengaruh:**
  - `file1.go` - function X memanggil function yang diubah
  - `Component.vue` - consume API yang diubah
- **Breaking change:** [Ya/Tidak]
- **Action required:**
  - [ ] Update caller di file1.go
  - [ ] Update Component.vue untuk handle response baru
```

### Contoh Analisis Dampak

#### ❌ SALAH (Tanpa Analisis Dampak)

```
## Task: Ubah format response order

### TODO LIST
1. [ ] Ubah response di order_handler.go
2. [ ] Done
```

#### ✅ BENAR (Dengan Analisis Dampak)

```
## Task: Ubah format response order

### Analisis Dampak Perubahan

#### File: `internal/handlers/order_handler.go`
- **Perubahan:** Ubah field `orderSn` menjadi `order_sn` (snake_case)
- **Caller/Consumer yang terpengaruh:**
  - `frontend/src/api/order.ts` - parsing response
  - `frontend/src/views/OrderList.vue` - display di tabel
  - `frontend/src/views/OrderDetail.vue` - display detail
- **Breaking change:** Ya - frontend expect `orderSn`
- **Action required:**
  - [ ] Update order.ts interface
  - [ ] Update OrderList.vue template binding
  - [ ] Update OrderDetail.vue template binding
  - [ ] Verify tabel tidak ada kolom kosong setelah perubahan

### TODO LIST
1. [ ] **[Phase 1] Analisis**
   - [ ] Grep semua penggunaan `orderSn` di frontend
   - [ ] List semua component yang terpengaruh

2. [ ] **[Phase 2] Backend**
   - [ ] Ubah response format di order_handler.go

3. [ ] **[Phase 3] Frontend**
   - [ ] Update interface di order.ts
   - [ ] Update OrderList.vue
   - [ ] Update OrderDetail.vue

4. [ ] **[Phase 4] Testing**
   - [ ] Verify data muncul di tabel (tidak ada kolom kosong)
   - [ ] go build && go test
```

---

## 3. OUTPUT FORMAT: TODO LIST

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
- **Analisis dampak sudah dilakukan: [Ya/Tidak]**

### Analisis Dampak Perubahan

[Wajib diisi - lihat template di Section 2]

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

| File              | Estimasi Baris | Action        |
| ----------------- | -------------- | ------------- |
| `path/to/file.go` | ~150 baris     | Create/Modify |
| ...               | ...            | ...           |

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

## 4. QUALITY GATES

Plan VALID hanya jika SEMUA gate terpenuhi:

| #   | Gate                | Requirement                                       |
| --- | ------------------- | ------------------------------------------------- |
| 1   | Ada TODO List       | Format checklist [ ] yang bisa dieksekusi         |
| 2   | **Analisis Dampak** | **WAJIB ada untuk setiap file yang dimodifikasi** |
| 3   | File Size           | Semua file < 300 baris (models: 500)              |
| 4   | Architecture        | Handler → Service → Repository                    |
| 5   | JSON Tags           | Semua snake_case                                  |
| 6   | Testing Phase       | go build + go test ada di todo                    |
| 7   | Tenant Check        | Validasi tenant_id jika endpoint protected        |
| 8   | Cleanup Phase       | DRY, SRP review ada di todo                       |
| 9   | Evidence Type       | Disebutkan di plan                                |

**Jika ada gate yang GAGAL → revisi plan sebelum eksekusi.**

---

## 5. ANTI-PATTERNS (DILARANG)

| #   | Jangan                              | Lakukan                                    |
| --- | ----------------------------------- | ------------------------------------------ |
| 1   | Output prose/paragraph panjang      | Output TODO LIST dengan [ ]                |
| 2   | **Skip analisis dampak**            | **WAJIB analisis dampak setiap perubahan** |
| 3   | Skip file size limit                | Tulis "max 300 baris" di setiap file       |
| 4   | Business logic di Handler           | Arahkan ke Service layer                   |
| 5   | Skip tenant_id validation           | Selalu validasi di protected endpoints     |
| 6   | camelCase di JSON response          | Gunakan snake_case                         |
| 7   | Skip testing phase                  | WAJIB ada go build + go test               |
| 8   | Assume default tenant               | Explicit error jika missing                |
| 9   | Skip cleanup phase                  | WAJIB ada DRY/SRP review                   |
| 10  | Ubah API tanpa cek frontend         | Cek semua consumer di frontend             |
| 11  | Ubah DB schema tanpa migration plan | Selalu sertakan migration steps            |

---

## 6. EVIDENCE REQUIREMENTS

| Task Type             | Required Evidence           |
| --------------------- | --------------------------- |
| Unit Test / Code Only | Test output saja            |
| Integration / API     | Docker log ATAU Test output |
| Full Feature          | Docker log DAN Test output  |
| Documentation         | Visual confirmation         |

**Evidence harus membuktikan masalah spesifik sudah teratasi.**

---

## 7. CRITICAL RULES (dari AGENTS.md)

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

| File                                 | Estimasi   | Action |
| ------------------------------------ | ---------- | ------ |
| `internal/handlers/order_handler.go` | +30 baris  | Modify |
| `internal/services/order_service.go` | +50 baris  | Modify |
| `internal/utils/export_utils.go`     | ~100 baris | Create |

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

## 8. DELEGATION RULES (MEMPERCEPAT EKSEKUSI)

> **Delegasi ke sub-agent untuk parallel processing dan expertise yang lebih fokus.**
> **MAKSIMAL 2 delegasi paralel** untuk menghindari overload.

### 8.1 Kapan WAJIB Delegate

| Situasi                           | Delegate Ke                                    | Alasan                          |
| --------------------------------- | ---------------------------------------------- | ------------------------------- |
| Cari SDK docs / official API      | `@librarian`                                   | Expertise di external reference |
| Cari existing pattern di codebase | `@explore`                                     | Faster contextual grep          |
| Architecture/design question      | `@oracle`                                      | High-IQ reasoning               |
| UI/UX / Frontend work             | `delegate_task(category="visual-engineering")` | Frontend specialist             |
| Complex logic problem             | `delegate_task(category="ultrabrain")`         | Deep reasoning                  |
| Quick/trivial fix                 | `delegate_task(category="quick")`              | Fast execution                  |

### 8.2 Delegation Strategy

```
┌─────────────────────────────────────────────────────────────────────────┐
│ PARALLEL DELEGATION (Mempercepat Research)                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Ketika butuh research, fire 2 agent PARALEL:                           │
│                                                                         │
│  // Contoh: Fix Shopee API error                                        │
│  @librarian: "Cari Shopee GetOrderList API docs, request/response"      │
│  @explore: "Cari existing shopee API pattern di codebase ini"           │
│                                                                         │
│  → Keduanya jalan paralel, hasil digabung untuk fix                     │
│                                                                         │
│  ⚠️ MAKSIMAL 2 paralel untuk menghindari overload                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 8.3 Delegation dalam TODO LIST

```markdown
### TODO LIST

1. [ ] **[Phase 0] Parallel Research** ⚠️ DELEGATE
   - [ ] 🔀 PARALLEL #1: @librarian → Cari [specific docs]
   - [ ] 🔀 PARALLEL #2: @explore → Cari [existing pattern]
   - [ ] ⏳ Tunggu hasil, gabungkan findings

2. [ ] **[Phase 0.5] Trace Flow** (jika bug fix)
   - [ ] Trace berdasarkan research result
3. [ ] **[Phase 1] Implementasi**
   - [ ] Fix berdasarkan research + trace
```

### 8.4 Delegation Format

```markdown
## Delegation Request

**Agent:** @librarian / @explore / @oracle
**Task:** [specific question/search]
**Expected Output:** [apa yang diharapkan]
**Context:** [background info yang relevan]
```

### 8.5 Contoh Delegation yang Efektif

#### ❌ SALAH (Tidak Delegate, Semua Sendiri)

```
Task: Fix Shopee order sync

*coba fix sendiri*
*gagal*
*coba lagi*
*gagal*
*coba lagi*
... (wasting time tanpa reference)
```

#### ✅ BENAR (Delegate untuk Research)

```
Task: Fix Shopee order sync

## Fix Attempt #1
**Failure Count:** 1
*gagal - error: invalid signature*

## Fix Attempt #2 - TRACE FLOW + DELEGATION
**Failure Count:** 2

**Parallel Delegation:**
🔀 @librarian: "Cari Shopee API signature generation docs,
               termasuk parameter order dan hash algorithm"
🔀 @explore: "Cari existing shopee signature generation di codebase"

**Results:**
- @librarian: Signature = SHA256(base_string + secret),
              base_string harus sorted by key
- @explore: File `internal/shopee/auth.go` line 45 ada existing impl

**Root Cause:** Parameter tidak di-sort sebelum hash
**Fix:** Update signature generation sesuai docs

*test* → BERHASIL
```

### 8.6 Delegation Decision Tree

```
┌─────────────────────────────────────────────────────────────────────────┐
│ KAPAN DELEGATE vs KERJAKAN SENDIRI                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  DELEGATE jika:                                                         │
│  ├─ Butuh external docs (SDK, API) → @librarian                         │
│  ├─ Butuh cari pattern di codebase → @explore                           │
│  ├─ Butuh architecture decision → @oracle                               │
│  ├─ Frontend/UI work → delegate_task(visual-engineering)                │
│  ├─ Failure >= 2 dan butuh research → @librarian + @explore             │
│  └─ Task bisa di-parallelkan → fire 2 agent sekaligus                   │
│                                                                         │
│  KERJAKAN SENDIRI jika:                                                 │
│  ├─ Simple edit yang sudah jelas                                        │
│  ├─ Sudah punya reference yang cukup                                    │
│  ├─ Task trivial (typo fix, formatting)                                 │
│  └─ Overhead delegasi > benefit                                         │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 9. FAILURE COUNTER RULE (UNTUK EKSEKUTOR)

> **⚠️ RULE INI WAJIB DIIKUTI OLEH AI EKSEKUTOR (Sisyphus/Builder)**
>
> AI sering tidak sadar sedang stuck/looping. Rule ini MEMAKSA awareness.

### 9.1 Definisi Failure

| Kondisi                                 | Count |
| --------------------------------------- | ----- |
| `go build` gagal setelah edit           | +1    |
| `go test` gagal setelah fix             | +1    |
| Error yang sama muncul lagi setelah fix | +1    |
| API call gagal dengan error yang sama   | +1    |
| Fix tidak menyelesaikan masalah         | +1    |

### 9.2 Mandatory Action by Failure Count

```
┌─────────────────────────────────────────────────────────────────────────┐
│ FAILURE COUNT → MANDATORY ACTION (TIDAK BISA DIABAIKAN)                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  FAILURE = 1 (Pertama kali gagal)                                       │
│  → Boleh coba fix langsung                                              │
│  → TAPI catat: "Failure #1: [error message]"                            │
│                                                                         │
│  FAILURE = 2 (Gagal kedua kali) ⚠️ TRACE FLOW ACTIVATED                 │
│  → STOP! Jangan langsung fix lagi                                       │
│  → WAJIB: Trace flow dari frontend → database                           │
│  → WAJIB: Identify di layer mana root cause                             │
│  → Tulis: "Failure #2 - TRACE FLOW protocol activated"                  │
│                                                                         │
│  FAILURE >= 3 (Gagal 3x atau lebih) 🚨 RESEARCH ACTIVATED               │
│  → STOP TOTAL! Tidak boleh edit code tanpa research                     │
│  → WAJIB: @librarian untuk cari SDK docs / official API docs            │
│  → WAJIB: @explore untuk cari existing pattern di codebase              │
│  → WAJIB: @oracle jika architecture/design issue                        │
│  → Tulis: "Failure #3+ - RESEARCH protocol activated"                   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 9.3 Format Wajib di Setiap Fix Attempt

```markdown
## Fix Attempt #[N]

**Failure Count:** [current count]
**Previous Error:** [error sebelumnya]
**Hypothesis:** [kenapa ini akan berhasil]
**Action:** [apa yang akan dilakukan]

[Jika failure >= 2, WAJIB tambah:]
**Trace Flow Result:**

- Frontend sends: [apa]
- Handler receives: [apa]
- Service processes: [apa]
- Repository queries: [apa]
- Root cause identified: [di layer mana]

[Jika failure >= 3, WAJIB tambah:]
**Research Result:**

- @librarian found: [hasil research]
- @explore found: [existing pattern]
- Reference: [link/source]
```

---

## 10. EXECUTOR QUICK REFERENCE (CHEAT SHEET)

> **Print ini dan ikuti setiap eksekusi task**

```
┌─────────────────────────────────────────────────────────────────────────┐
│ EXECUTOR CHECKLIST - SEBELUM MULAI                                      │
├─────────────────────────────────────────────────────────────────────────┤
│ [ ] Baca AGENTS.md (Critical Rules section)                             │
│ [ ] Baca PROMETHEUS_RULES.md (Section 2: Research & Trace Flow)         │
│ [ ] Identify task type: BUG FIX atau NEW FEATURE?                       │
│     • Bug fix → Siap-siap trace flow jika failure >= 2                  │
│     • New feature dengan external API → Research dulu (delegate!)       │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ EXECUTOR CHECKLIST - SAAT EKSEKUSI                                      │
├─────────────────────────────────────────────────────────────────────────┤
│ [ ] Track failure count (WAJIB!)                                        │
│ [ ] Failure = 1 → Boleh fix langsung, CATAT error                       │
│ [ ] Failure = 2 → STOP, trace flow + delegate @explore                  │
│ [ ] Failure >= 3 → STOP, delegate @librarian + @explore (PARALEL)       │
│ [ ] Jangan loop fix-test tanpa trace/research!                          │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ 🔴 LOCAL SDK PRIORITY (CARI DI SINI DULU!)                              │
├─────────────────────────────────────────────────────────────────────────┤
│ Shopee  → backend/shopee-sdk/     (orders.go, products.go, client.go)  │
│ Lazada  → backend/lazada-sdk/     (order.go, product.go, auth.go)      │
│ Lazada  → backend/lazada_sdk/iop-sdk-go/  (Official IOP SDK)           │
│ TikTok  → backend/tiktok_sdk/     (100+ files, comprehensive!)         │
│                                                                         │
│ URUTAN WAJIB:                                                           │
│ 1️⃣ @explore: "Cari di backend/*sdk*/" → LOKAL DULU                      │
│ 2️⃣ @explore: "Cari di internal/" → EXISTING PATTERN                     │
│ 3️⃣ @librarian: External docs → HANYA JIKA STUCK                         │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ DELEGATION RULES (MEMPERCEPAT)                                          │
├─────────────────────────────────────────────────────────────────────────┤
│ ⚠️ MAKSIMAL 2 DELEGASI PARALEL                                          │
│                                                                         │
│ @explore    → Existing pattern di codebase + LOCAL SDK                  │
│ @librarian  → External docs (LAST RESORT - HANYA JIKA TIDAK DI LOKAL)  │
│ @oracle     → Architecture decision                                     │
│                                                                         │
│ delegate_task(category="visual-engineering") → Frontend/UI work         │
│ delegate_task(category="ultrabrain") → Complex logic                    │
│ delegate_task(category="quick") → Trivial tasks                         │
│                                                                         │
│ CONTOH PARALEL:                                                         │
│ 🔀 @explore: "Cari implementasi GetOrderList di backend/shopee-sdk/"    │
│ 🔀 @explore: "Cari existing shopee order pattern di internal/"          │
│ → Tunggu hasil → Gabungkan → Fix                                        │
│                                                                         │
│ HANYA JIKA TIDAK KETEMU:                                                │
│ 🔀 @librarian: "Cari official Shopee API docs untuk error code X"       │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ EXECUTOR CHECKLIST - SEBELUM SELESAI                                    │
├─────────────────────────────────────────────────────────────────────────┤
│ [ ] go build ./... passes                                               │
│ [ ] go test ./... passes                                                │
│ [ ] Tidak ada looping (max 2 fix attempts tanpa trace)                  │
│ [ ] Evidence sesuai task type sudah dikumpulkan                         │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│ PLATFORM SDK REFERENCES                                                 │
├─────────────────────────────────────────────────────────────────────────┤
│ Shopee:    https://open.shopee.com/documents                            │
│ Lazada:    https://open.lazada.com/doc/api.htm                          │
│ TikTok:    https://partner.tiktokshop.com/doc                           │
│ Tokopedia: https://developer.tokopedia.com/                             │
└─────────────────────────────────────────────────────────────────────────┘
```

---

**File Version:** 3.0  
**Last Updated:** 2026-02-02  
**Status:** ACTIVE - MANDATORY COMPLIANCE
**Changes:** Added Delegation Rules, Failure Counter Rule, Trace Flow Protocol, Research Protocol, Executor Quick Reference
