# Fix GitHub Copilot Model Format Configuration

## TL;DR

> **Quick Summary**: Perbaiki format nama model yang salah di konfigurasi Oh My OpenCode untuk GitHub Copilot. Model `gemini-3-pro-preview` tidak ada di Copilot - yang benar adalah `gemini-3-pro`.
>
> **Deliverables**:
>
> - File `oh-my-opencode-copilot.json` dengan format model yang benar
>
> **Estimated Effort**: Quick (5 menit)
> **Parallel Execution**: NO - sequential single task
> **Critical Path**: Fix → Verify

---

## Context

### Original Request

User ingin mengoptimalkan konfigurasi Oh My OpenCode untuk GitHub Copilot Pro. Pada analisis ditemukan bahwa format nama model `gemini-3-pro-preview` tidak sesuai dengan yang tersedia di GitHub Copilot.

### Interview Summary

**Key Discussions**:

- User menyediakan screenshot daftar model di GitHub Copilot
- Teridentifikasi bahwa model "Gemini 3 Pro (Preview)" di dropdown = `gemini-3-pro` dalam konfigurasi

**Research Findings**:

- Model `gemini-3-pro-preview` kemungkinan dari provider Google langsung (Gemini CLI), bukan Copilot
- Format Copilot: `github-copilot/<model-name>` tanpa suffix `-preview`

---

## Work Objectives

### Core Objective

Memperbaiki 6 instance nama model yang salah dari `gemini-3-pro-preview` menjadi `gemini-3-pro`.

### Concrete Deliverables

- `opencode-configs/oh-my-opencode-copilot.json` dengan format yang benar

### Definition of Done

- [ ] Semua instance `gemini-3-pro-preview` diganti dengan `gemini-3-pro`
- [ ] File JSON valid (no syntax errors)

### Must Have

- Format model sesuai dengan yang tersedia di GitHub Copilot

### Must NOT Have (Guardrails)

- JANGAN ubah model lain yang sudah benar (claude-_, gpt-_)
- JANGAN hapus atau tambah agent/category baru
- Ini hanya fix format, bukan optimalisasi model

---

## Verification Strategy

### Test Decision

- **Infrastructure exists**: NO (tidak perlu test framework)
- **User wants tests**: Manual-only
- **Framework**: none

### Manual Verification Procedure

```bash
# Verify JSON is valid after edit
cat opencode-configs/oh-my-opencode-copilot.json | jq '.'

# Verify no gemini-3-pro-preview remains
grep -c "gemini-3-pro-preview" opencode-configs/oh-my-opencode-copilot.json
# Expected: 0

# Verify gemini-3-pro exists (should be 6)
grep -c "gemini-3-pro" opencode-configs/oh-my-opencode-copilot.json
# Expected: 6
```

---

## TODOs

- [ ] 1. Replace all `gemini-3-pro-preview` with `gemini-3-pro`

  **What to do**:
  - Open `opencode-configs/oh-my-opencode-copilot.json`
  - Find and replace ALL instances:
    - `github-copilot/gemini-3-pro-preview` → `github-copilot/gemini-3-pro`
  - Locations to change (6 total):
    - Line 20: agents.librarian.model
    - Line 27: agents.multimodal-looker.model
    - Line 33: agents.momus.model
    - Line 44: categories.visual-engineering.model
    - Line 47: categories.artistry.model
    - Line 65: categories.review.model

  **Must NOT do**:
  - Jangan ubah model lain
  - Jangan ubah struktur file

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: `[]` (no special skills needed)

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Sequential (single task)
  - **Blocks**: None (this is the only task)
  - **Blocked By**: None

  **References**:
  - `opencode-configs/oh-my-opencode-copilot.json:20,27,33,44,47,65` - Lines to modify
  - Screenshot dari user showing "Gemini 3 Pro (Preview)" as available model

  **Acceptance Criteria**:

  ```bash
  # Agent runs:
  grep -c "gemini-3-pro-preview" opencode-configs/oh-my-opencode-copilot.json
  # Assert: Output is "0" (no instances remain)

  grep "gemini-3-pro" opencode-configs/oh-my-opencode-copilot.json | wc -l
  # Assert: Output is "6" (all 6 instances replaced)

  cat opencode-configs/oh-my-opencode-copilot.json | jq '.' > /dev/null && echo "Valid JSON"
  # Assert: Output is "Valid JSON"
  ```

  **Commit**: YES
  - Message: `fix(config): correct gemini model name format for github-copilot`
  - Files: `opencode-configs/oh-my-opencode-copilot.json`

---

## Success Criteria

### Verification Commands

```bash
# Must show 0
grep -c "gemini-3-pro-preview" opencode-configs/oh-my-opencode-copilot.json

# Must show 6
grep -c "gemini-3-pro" opencode-configs/oh-my-opencode-copilot.json

# Must be valid JSON
cat opencode-configs/oh-my-opencode-copilot.json | jq '.' > /dev/null
```

### Final Checklist

- [ ] No `gemini-3-pro-preview` in file
- [ ] 6 instances of `gemini-3-pro` present
- [ ] JSON is valid
- [ ] No other changes made
