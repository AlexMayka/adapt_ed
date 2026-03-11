INFRA_DIR=infra
INFRA_COMPOSE=docker compose --env-file $(INFRA_DIR)/.env -f $(INFRA_DIR)/docker-compose.yml
LOAD_ENV=@set -a; \
	[ -f $(INFRA_DIR)/.env ] && . ./$(INFRA_DIR)/.env || true; \
	[ -f ./.env ] && . ./.env || true; \
	set +a;

run:
	echo "Запуск программы"
	$(LOAD_ENV) \
	go run ./cmd/main.go

build:
	echo "Сборка бинарника"
	go build -o bin/adapt-ed-backend ./cmd


run-cmd-make:
	echo "Запуск бинарного файла"
	./bin/adapt-ed-backend

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
	$(LOAD_ENV) \
	goose -dir migrations postgres "host=$${PG_HOST} port=$${PG_PORT} user=$${PG_USER} password=$${PG_PASSWORD} dbname=$${PG_DB} sslmode=disable" status

migrate-up:
	echo "Запуск миграций"
	$(LOAD_ENV) \
	goose -dir migrations postgres "host=$${PG_HOST} port=$${PG_PORT} user=$${PG_USER} password=$${PG_PASSWORD} dbname=$${PG_DB} sslmode=disable" up


APP_COMPOSE=docker compose --env-file $(INFRA_DIR)/.env --env-file .env

app-up: infra-up
	$(APP_COMPOSE) up -d --build

app-down:
	$(APP_COMPOSE) down

app-rebuild: infra-up
	$(APP_COMPOSE) up -d --build --force-recreate

app-logs:
	$(APP_COMPOSE) logs -f --tail=200

infra-copy-env:
	cp -n $(INFRA_DIR)/.env.example $(INFRA_DIR)/.env || true

infra-up: infra-copy-env
	$(INFRA_COMPOSE) up -d

infra-down:
	$(INFRA_COMPOSE) down -v

infra-ps:
	$(INFRA_COMPOSE) ps

infra-logs:
	$(INFRA_COMPOSE) logs -f --tail=200
