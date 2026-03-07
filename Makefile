run:
	echo "Запуск программы"
	@set -a; . ./.env; set +a; \
	go run ./cmd/main.go

build:
	echo "Сборка бинарника"
	go build -o bin/backend-sales-radar ./cmd


run-cmd-make:
	echo "Запуск бинарного файла"
	./bin/backend-sales-radar

create-migrations-goose-sql:
	echo "Создание миграции SQL-file"
	@read -p "Введите название миграции: " name; \
	goose -dir migrations create $$name sql


create-migrations-goose-go:
	echo "Создание миграции Go-file"
	@read -p "Введите название миграции: " name; \
	goose -dir migrations create $$name go

migrate-status:
	echo "Статус goose"
	@set -a; . ./.env; set +a; \
	goose -dir migrations postgres "host=$${SR_PG_HOST} port=$${SR_PG_PORT} user=$${SR_PG_USER} password=$${SR_PG_PASSWORD} dbname=$${SR_PG_DB} sslmode=disable" status

migrate-up:
	echo "Запуск миграций"
	@set -a; . ./.env; set +a; \
	goose -dir migrations postgres "host=$${SR_PG_HOST} port=$${SR_PG_PORT} user=$${SR_PG_USER} password=$${SR_PG_PASSWORD} dbname=$${SR_PG_DB} sslmode=disable" up

