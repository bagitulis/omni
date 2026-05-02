# OMNI Backup Sync Guide

## Overview

OMNI menggunakan **hybrid sync mechanism** yang menggabungkan:
- **Git**: Untuk `manifest.json` dan `_schema/*.sql` (kecil, diffable, tracked)
- **OneDrive**: Untuk binary data `.sql.gz` (44MB, auto-synced, tidak di git)

## Workflow

### PC Lama (Backup + Sync)

```powershell
# 1. Backup database
python build.py backup

# 2. Sync ke OneDrive
python build.py backup --sync
# Atau manual:
.\backups\db-tools\sync-backup.ps1 to-onedrive

# 3. Commit manifest + schema ke git
git add backups/smart/manifest.json backups/smart/_schema/
git commit -m "backup: $(Get-Date -Format 'yyyy-MM-dd HH:mm')"
git push
```

### PC Baru (Clone + Sync + Restore)

```powershell
# 1. Clone repo (kecil, ~50MB)
git clone https://github.com/bagitulis/omni.git
cd omni

# 2. Sync dari OneDrive
.\backups\db-tools\sync-backup.ps1 from-onedrive

# 3. Restore database
python build.py restore

# Atau satu command:
python build.py sync-restore
```

## Struktur Backup

```
backups/smart/                    # Lokal (44MB, TIDAK di git)
├── manifest.json                    # DI GIT (DB state, diffable)
├── _schema/
│   ├── system.sql                   # DI GIT (DDL, diffable)
│   ├── system.sql.gz                # TIDAK di git (binary)
│   └── tenant_*.sql / .sql.gz       # Mixed (plain in git, gzip not)
├── system/
│   └── *.sql.gz                     # TIDAK di git (binary data)
└── tenant_*/
    ├── *.sql.gz                     # TIDAK di git (binary data)
    └── *.chunk*.sql.gz              # TIDAK di git (binary data)

OneDrive/Data/omni-backup/        # Cloud sync (44MB)
└── (mirror dari backups/smart/)
```

## Commands

| Command | Description |
|---------|-------------|
| `python build.py backup` | Backup only |
| `python build.py backup --sync` | Backup + sync to OneDrive |
| `python build.py sync-restore` | Sync from OneDrive + restore |
| `python build.py restore` | Restore only (local backup) |
| `.\backups\db-tools\sync-backup.ps1 to-onedrive` | Sync to OneDrive |
| `.\backups\db-tools\sync-backup.ps1 from-onedrive` | Sync from OneDrive |
| `.\backups\db-tools\sync-backup.ps1 status` | Check sync status |

## Git Size

- **Sebelum cleanup**: 926 MB (binary files tracked)
- **Setelah cleanup**: ~50 MB (hanya source + manifest + schema)
- **Binary data**: 44 MB di OneDrive (auto-synced)

## Troubleshooting

### OneDrive tidak ditemukan
```
[ERROR] OneDrive not found
```
Pastikan OneDrive terinstall dan signed in. Sync script mencari di:
- `%USERPROFILE%\OneDrive`
- `%USERPROFILE%\OneDrive - Personal`

### Sync conflict (PC A dan B backup bersamaan)
OneDrive akan handle conflict dengan membuat versi file. Manifest.json di git akan menunjukkan mana yang lebih baru (timestamp).

### Restore gagal setelah sync
Pastikan PostgreSQL container running:
```powershell
docker ps  # check omni-postgres
docker compose up -d postgres  # start if needed
python build.py restore
```

## Keuntungan vs Extensions

| Aspek | Extensions (NDJSON) | OMNI (Hybrid) |
|-------|---------------------|---------------|
| Format | NDJSON (text) | pg_dump SQL (binary) |
| Data size | ~650 KB | 44 MB |
| Git size | 18 MB | ~50 MB (setelah cleanup) |
| Sync mechanism | Git push/pull | OneDrive auto-sync |
| Multi-PC | Git clone | OneDrive sync + git pull |
| Offline access | Yes (git history) | Yes (OneDrive local cache) |
| Speed | Fast (small data) | Fast (compressed binary) |
| Best for | Small SQLite | Large PostgreSQL |

## Future Improvements

1. **Git cleanup**: Setelah mekanisme ini stabil 2-3 siklus, jalankan `git filter-repo` untuk shrink repo dari 926MB → 50MB
2. **Compression**: OneDrive sudah compress, tapi bisa tambah zip untuk transfer lebih cepat
3. **Incremental**: Pertimbangkan WAL archiving untuk incremental backup (bukan full dump)
