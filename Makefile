.PHONY: up down logs check migrate test fmt reset
up:
	docker compose up -d --build
down:
	docker compose down
logs:
	docker compose logs -f --tail=100
check:
	docker compose config --quiet
migrate:
	docker compose run --rm user-migrate up
	docker compose run --rm commerce-migrate up
	docker compose run --rm wallet-migrate up
test:
	go test ./...
	npm --prefix web/portal run typecheck
fmt:
	gofmt -w internal services
reset:
	powershell -ExecutionPolicy Bypass -File scripts/dev.ps1 reset
