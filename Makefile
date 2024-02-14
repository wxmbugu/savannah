pwd := ${CURDIR}

postgres:
	docker run --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=secret -p 5432:5432 -v ~/postgres_data:/data/db -d postgres:14-alpine
createdb:
	docker exec -it postgres createdb --username=postgres --owner=postgres savannah
startdb:
	docker start postgres
stopdb:
	docker stop postgres
accessdb:
	docker exec -it postgres psql -U postgres savannah
dropdb:
	docker exec -it postgres dropdb savannah
migrate:
	docker pull migrate/migrate
migrateup:
	docker run -v "$(pwd)/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/savannah?sslmode=disable" -verbose up
migratedown:
	docker run -v "$(pwd)/db/migrations:/migrations" --network host migrate/migrate -path=/migrations/ -database "postgresql://postgres:secret@localhost:5432/savannah?sslmode=disable" -verbose down -all
test:
	go test -v -cover ./...
server:
	go run ./cmd/app
app:
	go build ./cmd/app

.PHONY: postgres test
