# Contributing to OMNI

## Development Setup

### Prerequisites

- Go 1.24+
- Node.js 18+ with npm
- Python 3.11+
- Docker Desktop (Linux containers mode)
- 5GB+ free disk space

### Getting Started

1. Clone the repository
2. Copy environment config: `cp .env.example .env`
3. Start the build system: `python build.py`
4. Select "Smart build" from the menu

### Development Workflow

1. Create a feature branch from `main`
2. Make changes following the architecture patterns
3. Run tests: `go test ./...` (backend) and `npm run build` (frontend)
4. Run backup if schema changed: `python build.py backup`
5. Commit with conventional format: `type(scope): description`

## Architecture

```
Request -> Handler -> Service -> Repository -> Database
```

- **Handler**: Parse request, validate, format response
- **Service**: Business logic, orchestration
- **Repository**: Database access ONLY

## Code Standards

- Go: Follow `backend/AGENTS.md` conventions
- React: Follow `frontend/AGENTS.md` conventions
- JSON fields: Always `snake_case`
- File limit: ~300 lines per code file (quality signal)

## Commit Message Format

```
type(scope): short description

Types: feat, fix, refactor, docs, test, chore
Scope: backend, frontend, build, db, sdk
```

## Database Changes

When modifying database schema:

1. Apply changes via migration or code
2. Run `python build.py smart` to deploy
3. Run `python build.py backup` to capture new schema
4. Commit the updated `backups/smart/manifest.json` and `backups/smart/_schema/*.sql`

## Testing

### Backend
```bash
cd backend
go build ./...
go test ./...
```

### Frontend
```bash
cd frontend
npm run build
npm run test
```

### Build System
```bash
cd scripts/python-build
python -m pytest tests/ -v
```

## Backup/Restore

- Binary backup data is NOT tracked in git (see .gitignore)
- Only `manifest.json` and `_schema/*.sql` are tracked
- For PC-to-PC sync, copy the `backups/smart/` folder directly

## License

Proprietary - All rights reserved.
