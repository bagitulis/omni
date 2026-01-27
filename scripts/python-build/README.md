# Omni Build System

Modern build automation for Omni multi-platform e-commerce system.

## Quick Start

### Windows

```cmd
.\build.bat smart
```

### Linux/Mac

```bash
chmod +x build.sh
./build.sh smart
```

## Commands

### `smart` (Recommended)

Smart build with dependency caching.

- Only reinstalls npm dependencies if `package-lock.json` changed
- Uses Docker layer cache
- **Time:** 2-3 minutes

```bash
# Default (standard spec)
.\build.bat smart

# With specific spec
.\build.bat smart --spec=lowspec
.\build.bat smart --spec=highspec

# Skip frontend build
.\build.bat smart --skip-frontend
```

### `quick`

Quick restart - containers only (no rebuild).

- **Time:** 10-20 seconds

```bash
.\build.bat quick
```

### `full`

Full rebuild from scratch (no cache).

- Clears all caches
- Reinstalls all dependencies
- **Time:** 5-10 minutes

```bash
.\build.bat full
```

### `validate`

Check environment only (no build).

- Verifies Docker is running
- Checks Linux containers mode
- **Time:** 5 seconds

```bash
.\build.bat validate
```

### `clean`

Stop containers and cleanup.

- Stops all containers
- Removes unused Docker resources
- **Time:** 1-2 minutes

```bash
.\build.bat clean
```

### `status`

Show container and service health status.

```bash
.\build.bat status
```

## Specification Levels

- **`lowspec`** - 2GB RAM (VPS kecil)
- **`standard`** - 4GB RAM (Production - RECOMMENDED)
- **`highspec`** - 8GB+ RAM (High traffic)

## Architecture

```
omni_build/
├── models.py           # Data models (Pydantic)
├── config.py           # Configuration management
├── logger.py           # Rich console logging
├── error_handler.py    # 22+ error patterns with auto-fix
├── docker_manager.py   # Docker operations
├── frontend_builder.py # NPM build automation
├── health_checker.py   # Health checks
└── cli.py              # Command-line interface
```

## Error Recovery

The build system automatically detects and fixes 22+ error types:

**Docker Infrastructure (10 patterns)**

- Docker engine errors (500)
- Named pipe errors
- WSL2 mount cache corruption
- Hyper-V issues
- BuildKit errors
- Container name conflicts
- Dependency failures
- PostgreSQL data corruption

**Network/DNS (4 patterns)**

- DNS resolution failures
- Alpine repository errors
- Docker registry connectivity
- General network errors

**Resources (3 patterns)**

- Disk space exhaustion
- Out of memory (OOM)
- Port conflicts

**Build Errors (4 patterns)**

- NPM integrity errors
- TypeScript errors (requires manual fix)
- Import resolution errors (requires manual fix)
- Syntax errors (requires manual fix)

## Progressive Fix Levels

If auto-fix fails, the system escalates through 6 fix levels:

1. **Level 1:** Light cleanup + port conflicts
2. **Level 2:** Network flush + DNS repair
3. **Level 3:** Full DNS + network reset
4. **Level 4:** WSL restart + network reset
5. **Level 5:** Docker engine full restart
6. **Level 6:** Aggressive cleanup + system reset

## Development

### Setup Virtual Environment

```bash
cd scripts/python-build
python setup.py
```

### Run Tests

```bash
cd scripts/python-build
.venv\Scripts\activate  # Windows
source .venv/bin/activate  # Linux/Mac

pytest tests/ -v --cov=omni_build
```

### Dependencies

- Python 3.11+
- Docker Desktop (Linux mode)
- Node.js 18+ (for frontend)
- 5GB+ free disk space

## Success Criteria (AGENTS.md)

Every build must provide **2 proofs of success**:

1. **Proof 1:** Docker logs showing valid timestamps
2. **Proof 2:** Command output with explicit success confirmation

## Environment Variables

Override defaults via `.env` file:

```env
DOCKER_BUILD_TIMEOUT=1200
DOCKER_DEPLOY_TIMEOUT=600
MAX_BUILD_RETRIES=7
BACKEND_HEALTH_URL=http://localhost:3000/api/health
```

## Troubleshooting

### Docker Not Ready

```bash
# Check Docker status
.\build.bat validate

# If failed, restart Docker Desktop manually
```

### Build Fails Repeatedly

```bash
# Run full rebuild
.\build.bat full

# Check logs
docker logs omni-backend
docker logs omni-postgres
```

### Frontend Build Errors

```bash
# Force reinstall dependencies
.\build.bat smart --spec=standard

# Or manually:
cd frontend
npm ci --legacy-peer-deps
npm run build
```

## CI/CD Integration

See `.github/workflows/build-test.yml` for GitHub Actions setup.

## License

Proprietary - Omni Team

---

**Questions?** Check `AGENTS.md` for development guidelines.
