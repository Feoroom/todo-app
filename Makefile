help:
	@ echo "Usage: "
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' \ sed -e 's/^/ /'
## create a new db migration
.PHONY: migrate
migrate:
	migrate create -seq -ext .sql -dir ./migrations ${name}
## apply all new migrations
.PHONY: up
up:
	migrate -path ./migrations -database ${TODO_DB_DSN} up
## delete migrations
.PHONY: down
down:
	migrate -path ./migrations -database ${TODO_DB_DSN} down ${q}
##
.PHONY: audit
audit:
	go mod tidy
	go mod verify
	go fmt ./...
	go vet ./...
	#staticcheck ./...
	go test -race -vet=off ./...

.PHONY: vendor
vendor: audit
	go mod vendor


current_time = $(shell date --iso-8601=seconds)
version=$(shell git describe --tags --dirty)

.PHONY: build
build: audit
	@echo ${current_time}
	@echo ${version}
	go build -ldflags='-s -X main.BuildTime=${current_time} -X main.Version=${version}' -o ./bin/api ./cmd/api