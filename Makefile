postgres-up:
	docker run --name postgres-chat -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=root -e POSTGRES_DB=financial-chat -d postgres:latest

postgres-down:
	docker stop postgres-chat

postgres-run:
	docker start postgres-chat

postgres-rm:
	docker container rm -f postgres-chat

postgres-createdb:
	docker exec -it postgres-chat createdb --username=root --owner=root financial-chat

postgres-dropdb:
	docker exec -it postgres-recipes dropdb financial-chat

postgres-migrateup:
	migrate -path internal/services/datastore/postgresql/users/migration/ -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose up

postgres-migrateup1:
	migrate -path internal/services/datastore/postgresql/users/migration/ -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose up 1

postgres-migratedown:
	migrate -path internal/services/datastore/postgresql/users/migration/ -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose down

postgres-migratedown1:
	migrate -path internal/services/datastore/postgresql/users/migration/ -database "postgresql://root:root@localhost:5432/financial-chat?sslmode=disable" -verbose down 1

rabbitmq-up:
	docker run -d --name rabbitmq-chat -p 5672:5672 -p 15672:15672 rabbitmq:3.9-management

rabbitmq-down:
	docker stop rabbitmq-chat

rabbitmq-run:
	docker start rabbitmq-chat

rabbitmq-rm:
	docker container rm -f rabbitmq-chat

clean-all:
	docker container rm -f rabbitmq-chat
	docker container rm -f postgres-chat

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run cmd/chat/main.go

mock:
	mockgen -package mockeduserstore -destination internal/services/datastore/mocks/postgresql/users/mockedStore.go github.com/gmaschi/jobsity-go-financial-chat/internal/services/datastore/postgresql/users Store

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server mock