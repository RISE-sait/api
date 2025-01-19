build:
	@go build cmd/server/main.go

run:
	@go run cmd/server/main.go

docker-up:
	@docker-compose up -d

migrate:
	@go run cmd/migrate/main.go $(cmd)