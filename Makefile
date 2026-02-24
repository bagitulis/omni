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
