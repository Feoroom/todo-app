
migration:
	migrate create -seq -ext .sql -dir ./migrations ${name}

up:
	migrate -path ./migrations -database ${TODO_DB_DSN} up

down:
	migrate -path ./migrations -database ${TODO_DB_DSN} down ${q}