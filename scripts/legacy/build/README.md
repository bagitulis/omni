# Build System - Auto-Fix Edition

Sistem build yang modular dan robust dengan kemampuan auto-fix untuk **SEMUA** error umum.

## Quick Start

```bash
# Interactive menu (recommended)
.\build.bat

# Direct commands
.\build.bat quick        # Restart containers only
.\build.bat incremental  # Rebuild changed code
.\build.bat full         # Clean rebuild from scratch

# Diagnose Docker issues
cd scripts\build
.\docker-diagnose.ps1       # Check status only
.\docker-diagnose.ps1 -Fix  # Check and fix issues
```

## 🔧 Fitur Auto-Fix (22 Error Types)

Build system ini secara otomatis mendeteksi dan memperbaiki **semua error umum**:

### Docker Errors

| Error                             | Auto-Fix                              |
| --------------------------------- | ------------------------------------- |
| `no matching manifest`            | Switch Docker ke Linux mode           |
| Docker not running                | Restart Docker Desktop                |
| **500 Internal Server Error**     | Full Docker engine reset (new!)       |
| **Docker pipe/npipe errors**      | Reset Docker services + WSL (new!)    |
| WSL mount errors                  | Restart WSL + wait                    |
| **WSL2 kernel errors**            | Update WSL + restart Docker (new!)    |
| Network errors                    | Prune Docker networks                 |
| BuildKit errors                   | Clear builder cache                   |
| Container start failed            | Remove orphans + recreate             |
| Layer corruption                  | Clear image cache                     |
| Context canceled                  | Restart WSL                           |
| **Hyper-V/virtualization errors** | Restart vmcompute/hns services (new!) |

### System Resource Errors

| Error               | Auto-Fix                        |
| ------------------- | ------------------------------- |
| Disk space low      | Progressive cleanup (3 levels)  |
| Out of Memory (OOM) | Stop containers + free memory   |
| Port already in use | Auto-kill conflicting processes |
| Low memory (<1GB)   | Memory cleanup before build     |

### NPM/Build Errors

| Error                 | Auto-Fix                       |
| --------------------- | ------------------------------ |
| `npm ERR! ERESOLVE`   | Use --legacy-peer-deps         |
| `npm ERR! EINTEGRITY` | Clear cache + remove lock file |
| Module not found      | Clean reinstall node_modules   |
| JS heap out of memory | Increase Node.js heap size     |

### Database Errors

| Error                | Auto-Fix                 |
| -------------------- | ------------------------ |
| Prisma migrate error | Regenerate Prisma client |
| SQLite busy/locked   | Wait + retry             |

## 🔄 6-Level Progressive Recovery

Jika error spesifik tidak terdeteksi, sistem menerapkan recovery progresif:

```
Level 1: Check Docker health + Linux mode + port fix + light cleanup
Level 2: Memory cleanup + WSL restart + medium cleanup
Level 3: Docker engine health check + Container health fix + network reset
Level 4: Full Docker engine repair (handles 500 errors)
Level 5: Aggressive cleanup + Docker engine repair (nuclear option)
Level 6: Ultimate recovery - full Docker/WSL/Hyper-V reset
```

## 🏥 Docker Diagnostics Tool

Jika build terus gagal, jalankan diagnostics tool:

```powershell
cd scripts\build
.\docker-diagnose.ps1       # Check only
.\docker-diagnose.ps1 -Fix  # Check and auto-fix

# The tool checks:
# - Docker Desktop process
# - Docker daemon (500 errors, pipe errors)
# - Container mode (Linux/Windows)
# - WSL2 status (docker-desktop distros)
# - Hyper-V services (vmcompute, hns)
# - Port conflicts
```

## 📁 Struktur File

```
scripts/build/
├── build.ps1              # Main orchestrator
├── config.ps1             # Configuration (paths, timeouts)
├── logger.ps1             # Logging utilities
├── environment-checker.ps1 # Pre-build checks
├── docker-fixer.ps1       # Docker auto-fix (500 error, pipe, etc)
├── cleanup.ps1            # Cleanup utilities
├── nginx-validator.ps1    # Nginx config validation
├── retry-logic.ps1        # Retry with backoff
├── error-recovery.ps1     # Error detection & recovery (6 levels)
├── frontend-builder.ps1   # Frontend build
├── backend-builder.ps1    # Backend build
├── docker-operations.ps1  # Docker build/deploy
├── post-deploy.ps1        # Post-deploy tasks
├── docker-diagnose.ps1    # Docker diagnostics tool (new!)
└── README.md              # This file
```

## Mode Build

### 1. Quick Apply

- **Kapan**: Perubahan config, restart saja
- **Apa yang dilakukan**: Restart containers tanpa rebuild
- **Waktu**: ~10 detik

### 2. Incremental Build

- **Kapan**: Perubahan code (TypeScript, Vue)
- **Apa yang dilakukan**: Build frontend + Docker rebuild
- **Waktu**: ~2-3 menit

### 3. Full Rebuild

- **Kapan**: Perubahan dependencies (package.json)
- **Apa yang dilakukan**: Clean all + full npm install + Docker no-cache
- **Waktu**: ~5-10 menit

## Command Line Options

```bash
# With n8n
.\build-new.bat incremental -WithN8n

# Specific RAM spec
.\build-new.bat full -Spec lowspec
.\build-new.bat full -Spec standard
.\build-new.bat full -Spec highspec

# Combined
.\build-new.bat incremental -WithN8n -Spec standard
```

## Validation Only

```bash
.\build-new.bat validate
```

Checks:

- Docker running + Linux mode
- Disk space
- Port availability
- Nginx configuration

## Aggressive Cleanup

```bash
.\build-new.bat clean
```

Removes ALL:

- Docker images
- Docker volumes
- Docker cache
- Build artifacts

## Log Files

Logs are saved to: `logs/build_YYYY-MM-DD_HH-mm-ss.log`

## Recovery Levels

Jika build gagal, sistem akan mencoba fix secara progresif:

1. **Level 1**: Ensure Linux mode + light cleanup
2. **Level 2**: Restart WSL + medium cleanup
3. **Level 3**: Aggressive cleanup + full restart

## Customization

Edit `config.ps1` untuk mengubah:

```powershell
# Timeouts
$script:DockerStartTimeout = 60
$script:ContainerHealthTimeout = 30

# Retry settings
$script:MaxRetries = 3
$script:RetryDelayBase = 5

# Minimum disk space (GB)
$script:MinDiskSpaceGB = 5
```

## Troubleshooting

### Docker tetap error setelah auto-fix

1. Jalankan cleanup manual:

   ```bash
   .\build-new.bat clean
   ```

2. Restart Docker Desktop

3. Jalankan build lagi

### Port already in use

Check process yang menggunakan port:

```powershell
netstat -ano | findstr :80
netstat -ano | findstr :3000
```

### WSL terus-menerus error

Reset WSL:

```powershell
wsl --shutdown
wsl --unregister docker-desktop
```

Lalu restart Docker Desktop.
