.PHONY: test test-backend test-frontend test-lighthouse ci-local build

test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

test-lighthouse:
	cd backend && go test -v ./tests/e2e/lighthouse/

ci-local: test test-lighthouse

build:
	cd backend && go build ./... && cd ../frontend && npm run build

.PHONY: deploy-vps deploy-local deploy-dry-run setup-vps health backup

deploy-vps:
	bash scripts/deploy-vps.sh

deploy-local:
	bash scripts/deploy-local.sh

deploy-dry-run:
	bash scripts/deploy-vps.sh --dry-run

setup-vps:
	bash scripts/docker-host-setup.sh

health:
	bash scripts/health-check.sh --target $${TARGET:-local}

backup:
	bash scripts/backup.sh --target $${TARGET:-local}
