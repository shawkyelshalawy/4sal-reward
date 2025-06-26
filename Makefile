MIGRATION_PATH=internal/infrastructure/db/migrations
DB_URL=postgres://admin:secret@localhost:5432/rewarddb?sslmode=disable


migrate_up:
    @migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up

migrate_down:
    @migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down

migrate_create:
    @echo Enter migration name: && set /p NAME= && migrate create -ext sql -dir $(MIGRATION_PATH) -seq %NAME%

server:
    @go run ./cmd/main.go

test:
    @go test -v -cover ./...

.PHONY: migrate_up migrate_down migrate_create server test