# PROMETHEUS PLANNING RULES

> **STATUS: MANDATORY - TIDAK BOLEH DIABAIKAN**
>
> File ini WAJIB dibaca dan diterapkan oleh AI Prometheus sebelum membuat planning apapun.
> Pelanggaran terhadap rules ini akan menghasilkan plan yang INVALID.

---

## ACKNOWLEDGMENT WAJIB

Sebelum membuat plan, Prometheus HARUS menyatakan:

```
[PROMETHEUS ACKNOWLEDGMENT]
Saya telah membaca PROMETHEUS_RULES.md dan akan menerapkan:
- [ ] Pre-Planning Checklist (6 items)
- [ ] Planning Template
- [ ] Quality Gates (8 kriteria)
- [ ] Anti-Patterns Avoidance (7 items)
```

---

## 1. PRE-PLANNING STEPS (WAJIB 100%)

Sebelum membuat plan apapun, Prometheus HARUS melakukan langkah-langkah ini:

```
┌─────────────────────────────────────────────────────────────┐
│ LANGKAH PRE-PLANNING (lakukan secara berurutan)             │
├─────────────────────────────────────────────────────────────┤
│ 1. Baca SELURUH AGENTS.md                                   │
│ 2. Identifikasi SEMUA file yang akan dimodifikasi           │
│ 3. Verifikasi apakah ada perubahan database/schema          │
│ 4. Cek apakah task melibatkan multi-tenancy                 │
│ 5. Estimasi jumlah baris per file (HARUS < 300)             │
│ 6. Tentukan naming convention yang akan dipakai             │
└─────────────────────────────────────────────────────────────┘
```

**JIKA ADA LANGKAH YANG DILEWATI → PLAN INVALID**

---

## 2. PLANNING TEMPLATE (WAJIB DIIKUTI)

Setiap plan yang dibuat Prometheus HARUS menggunakan format ini:

```markdown
## Task: [Nama Task]

### PROMETHEUS ACKNOWLEDGMENT

Saya telah membaca PROMETHEUS_RULES.md dan AGENTS.md.

### Pre-Implementation Checks

- [ ] Sudah baca AGENTS.md: Ya
- [ ] Affected files: [list semua file]
- [ ] Database changes required: [Yes/No - jika Yes, perlu migration]
- [ ] Multi-tenant consideration: [Yes/No - jika Yes, wajib validate tenant_id]

### Architecture Compliance

- Layer yang terlibat: [Handler/Service/Repository]
- Pattern: Handler → Service → Repository
- Business logic location: SERVICE ONLY (bukan Handler)

### Naming Convention Applied

| Context             | Convention | Example                |
| ------------------- | ---------- | ---------------------- |
| Go struct fields    | PascalCase | OrderSN, TenantID      |
| JSON tags           | snake_case | json:"order_sn"        |
| Database columns    | snake_case | order_sn, tenant_id    |
| Go local variables  | camelCase  | orderService, tenantID |
| Frontend API types  | snake_case | item_id, product_name  |
| Frontend local vars | camelCase  | itemId, productName    |

### File Size Estimation

| File       | Estimated Lines | Status     |
| ---------- | --------------- | ---------- |
| [filename] | ~[X] lines      | ✅ OK/<300 |

### Security Checklist

- [ ] No default tenant (explicit error if missing)
- [ ] tenant_id validation di setiap protected endpoint
- [ ] No sensitive data in logs (password, token, api_key)
- [ ] Parameterized queries (no SQL injection)

### Implementation Phases

1. **Phase 1: Planning & Analysis** - [tasks]
2. **Phase 2: Implementation** - [tasks]
3. **Phase 3: Cleanup** - DRY, SRP, remove duplicates, <300 lines
4. **Phase 4: Testing** - go build ./... && go test ./...
5. **Phase 5: Docker Apply** - build.py smart/quick-fix

### Success Criteria

1. ✅ go build ./... passes
2. ✅ go test ./... passes (100%)
3. ✅ No false positives (success: true = real success)
4. ✅ All JSON responses use snake_case
5. ✅ All files < 300 lines
6. ✅ No business logic in handlers
```

---

## 3. QUALITY GATES (SEMUA HARUS TERPENUHI)

Plan dianggap VALID hanya jika memenuhi SEMUA kriteria berikut:

```
┌─────────────────────────────────────────────────────────────┐
│ KRITERIA VALIDASI (semua harus terpenuhi)                   │
├─────────────────────────────────────────────────────────────┤
│ 1. Menyebut max 300 lines per file                          │
│ 2. Architecture pattern Handler→Service→Repo benar         │
│ 3. Naming convention dispesifikasi dengan benar             │
│ 4. Testing phase (go build + go test) sudah termasuk        │
│ 5. Tidak ada asumsi default tenant                          │
│ 6. Cleanup phase (DRY, SRP, OOP) sudah termasuk             │
│ 7. Success criteria sudah didefinisikan dengan jelas        │
│ 8. Error handling pattern sudah dispesifikasi               │
└─────────────────────────────────────────────────────────────┘

Jika ada kriteria yang TIDAK terpenuhi → PLAN HARUS DIREVISI
```

---

## 4. ANTI-PATTERNS (DILARANG KERAS)

### ❌ JANGAN PERNAH:

| #   | Anti-Pattern                        | Konsekuensi             |
| --- | ----------------------------------- | ----------------------- |
| 1   | Plan tanpa menyebut file size limit | PLAN INVALID            |
| 2   | Business logic di Handler layer     | ARSITEKTUR SALAH        |
| 3   | Mengabaikan tenant_id validation    | SECURITY BREACH         |
| 4   | camelCase untuk JSON response       | API INCONSISTENT        |
| 5   | Plan tanpa testing phase            | KUALITAS TIDAK TERJAMIN |
| 6   | Mengasumsikan default tenant        | MULTI-TENANCY BROKEN    |
| 7   | Plan tanpa cleanup/refactor phase   | TECHNICAL DEBT          |

### ✅ HARUS SELALU:

| #   | Best Practice                    | Alasan               |
| --- | -------------------------------- | -------------------- |
| 1   | Breakdown task menjadi 5 phases  | Structured execution |
| 2   | Setiap task atomic dan testable  | Easy rollback        |
| 3   | Include rollback strategy        | Risk mitigation      |
| 4   | Specify dependencies antar tasks | Correct order        |
| 5   | Estimate complexity per task     | Realistic planning   |
| 6   | Include validation checkpoints   | Quality assurance    |

---

## 5. CRITICAL RULES REFERENCE (dari AGENTS.md)

### Top 11 Rules yang TIDAK BOLEH DILANGGAR:

```
1. ❌ NO FALSE POSITIVES
   - JANGAN return success: true jika ada error

2. ❌ NO ALIASES FOR WRONG NAMES
   - Fix nama langsung, jangan pakai alias/workaround

3. ❌ NO DEFAULT TENANT
   - Selalu validate tenant_id, error jika missing

4. 🎯 ALL JSON TAGS = snake_case
   - json:"order_sn" ✅ | json:"orderSn" ❌

5. 📏 MAX 300 LINES PER FILE
   - Exceptions: models (500), migrations (unlimited)

6. 🔐 ALWAYS USE context.Context
   - Semua DB/network operations HARUS terima context

7. 🏗️ CLEAN ARCHITECTURE
   - Handler → Service → Repository (strict)

8. 📝 STRUCTURED LOGGING ONLY
   - Pakai zerolog, JANGAN fmt.Printf atau log.Println

9. 🗄️ NO DATABASE CHANGES WITHOUT MIGRATION
   - Semua schema changes via migration script

10. 🧪 100% TEST SUCCESS REQUIRED
    - go build ./... DAN go test ./... HARUS PASS

11. 🔐 GIT OPERATIONS RESTRICTED
    - HANYA boleh: git add, commit, push (hati-hati)
    - DILARANG: reset, rebase, merge, checkout, branch -d,
      stash, revert, cherry-pick, push --force, git config
    - Konfirmasi user jika ragu
```

---

## 6. RESPONSE FORMAT RULES

### Success Response:

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

### Error Response:

```json
{
  "success": false,
  "error": "Human readable message",
  "code": "ERROR_CODE"
}
```

### ❌ SALAH:

```json
{
  "success": true,
  "message": "Failed to get orders" // FALSE POSITIVE!
}
```

---

## 7. ENFORCEMENT MECHANISM

### Bagaimana rules ini di-enforce:

1. **Prometheus WAJIB** menyertakan acknowledgment di awal plan
2. **Sisyphus (executor)** akan menolak plan tanpa acknowledgment
3. **Code review** akan check compliance terhadap rules
4. **Testing phase** akan validate output sesuai rules

### Jika Plan Tidak Comply:

```
┌─────────────────────────────────────────────────────────────┐
│ PLAN REJECTION FLOW                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Plan Created → Quality Gates Check → FAIL?                 │
│                                           ↓                 │
│                                    Return to Prometheus     │
│                                           ↓                 │
│                                    "Plan INVALID karena:    │
│                                     - [list violations]     │
│                                     Revise dan submit ulang"│
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 8. PLAN VALIDATION PSEUDO-CODE

```go
func ValidatePlan(plan Plan) error {
    // 1. Check acknowledgment
    if !plan.HasAcknowledgment {
        return errors.New("INVALID: Missing PROMETHEUS ACKNOWLEDGMENT")
    }

    // 2. Check file size estimates
    for _, file := range plan.AffectedFiles {
        if file.EstimatedLines > 300 && !file.IsModelOrMigration {
            return fmt.Errorf("INVALID: %s exceeds 300 lines (%d)",
                file.Name, file.EstimatedLines)
        }
    }

    // 3. Check architecture compliance
    if plan.HasBusinessLogicInHandler {
        return errors.New("INVALID: Business logic in Handler (must be in Service)")
    }

    // 4. Check naming conventions
    for _, jsonTag := range plan.JSONTags {
        if !isSnakeCase(jsonTag) {
            return fmt.Errorf("INVALID: JSON tag '%s' must be snake_case", jsonTag)
        }
    }

    // 5. Check testing phase
    if !plan.HasTestingPhase {
        return errors.New("INVALID: Missing testing phase (go build + go test)")
    }

    // 6. Check multi-tenancy
    if plan.RequiresAuth && !plan.HasTenantValidation {
        return errors.New("INVALID: Protected endpoint without tenant_id validation")
    }

    // 7. Check cleanup phase
    if !plan.HasCleanupPhase {
        return errors.New("INVALID: Missing cleanup phase (DRY, SRP, OOP)")
    }

    // 8. Check success criteria
    if len(plan.SuccessCriteria) == 0 {
        return errors.New("INVALID: Missing success criteria")
    }

    return nil // Plan is VALID
}
```

---

## 9. KRITERIA KEBERHASILAN (WAJIB)

> **Task TIDAK BOLEH dianggap selesai tanpa memenuhi kriteria ini**

### Dua Bukti Wajib

```
┌─────────────────────────────────────────────────────────────┐
│ BUKTI 1: Log Docker                                          │
├─────────────────────────────────────────────────────────────┤
│ Log yang menampilkan data valid SPESIFIK:                   │
│ • Tanggal/waktu terkait masalah yang diperbaiki             │
│ • Transaksi sukses yang relevan                             │
│ • Data yang membuktikan fix berhasil                        │
│                                                             │
│ ❌ TIDAK VALID: Log "Server started" tanpa konteks masalah  │
│ ✅ VALID: Log "Order TEST123 saved successfully"            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ BUKTI 2: Output Script/Test                                  │
├─────────────────────────────────────────────────────────────┤
│ Output dengan keterangan keberhasilan EKSPLISIT:            │
│ • go build ./... → "Build successful"                       │
│ • go test ./... → "PASS" untuk semua test                   │
│                                                             │
│ ❌ TIDAK VALID: Output tanpa hasil jelas                    │
│ ✅ VALID: "--- PASS: TestOrderService (0.05s)"              │
└─────────────────────────────────────────────────────────────┘
```

### ⚠️ PENTING

```
Bukti HARUS relevan langsung dengan INTI MASALAH yang diselesaikan.
Bukti running umum tanpa konteks spesifik TIDAK DIANGGAP SAH.
```

---

## 10. AI LEARNING REQUIREMENTS

Sebelum mengerjakan task, AI WAJIB mempelajari area-area berikut:

```
┌─────────────────────────────────────────────────────────────┐
│ AREA YANG HARUS DIPELAJARI                                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. BACKEND FLOW                                             │
│    • Request → Middleware → Handler → Service → Repo → DB  │
│    • PostgreSQL multi-tenant schema                         │
│    • GORM query patterns                                    │
│                                                             │
│ 2. PLATFORM SDK (Golang)                                    │
│    • Shopee Open Platform API                               │
│    • Lazada Open Platform API                               │
│    • TikTok Shop Open Platform API                          │
│    • Auth flow, rate limiting, error handling               │
│                                                             │
│ 3. FRONTEND-BACKEND INTEGRATION                             │
│    • Route naming dan mapping                               │
│    • Data format (snake_case API, camelCase internal)       │
│    • Auth flow (JWT → tenant_id extraction)                 │
│    • Error handling di kedua sisi                           │
│                                                             │
│ Referensi lengkap: AGENTS.md section "AI Learning Req."     │
└─────────────────────────────────────────────────────────────┘
```

---

## SUMMARY: PROMETHEUS MUST

```
┌─────────────────────────────────────────────────────────────┐
│ SEBELUM PLAN:                                               │
│ 1. Baca PROMETHEUS_RULES.md (file ini)                      │
│ 2. Baca AGENTS.md                                           │
│ 3. Pelajari area yang relevan (SDK, Flow, Integration)      │
│ 4. Tulis acknowledgment                                     │
├─────────────────────────────────────────────────────────────┤
│ SAAT MEMBUAT PLAN:                                          │
│ 1. Gunakan Planning Template                                │
│ 2. Isi SEMUA section                                        │
│ 3. Validate terhadap Quality Gates                          │
│ 4. Define Success Criteria dengan 2 bukti wajib             │
├─────────────────────────────────────────────────────────────┤
│ SETELAH PLAN:                                               │
│ 1. Self-review dengan validation pseudo-code                │
│ 2. Pastikan semua 10 Quality Gates PASS                     │
│ 3. Submit plan untuk eksekusi                               │
├─────────────────────────────────────────────────────────────┤
│ SETELAH EKSEKUSI:                                           │
│ 1. Sediakan Bukti 1: Docker Log yang relevan                │
│ 2. Sediakan Bukti 2: Test/Script output                     │
│ 3. Jelaskan relevansi bukti dengan masalah                  │
└─────────────────────────────────────────────────────────────┘
```

---

**File Version:** 1.0  
**Last Updated:** 2025-01-29  
**Status:** ACTIVE - MANDATORY COMPLIANCE
