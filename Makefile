INFRA_DIR=infra
ENV_FILE=$(INFRA_DIR)/.env
INFRA_COMPOSE=docker compose --env-file $(ENV_FILE) -f $(INFRA_DIR)/docker-compose.yml
LOAD_ENV=@set -a; \
	[ -f $(ENV_FILE) ] && . ./$(ENV_FILE) || true; \
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
	goose -dir migrations postgres "host=$${PG_HOST} port=$${PG_PORT} user=$${PG_USER} password=$${PG_PASSWORD} dbname=$${PG_DB} sslmode=$${PG_SSL_MODE}" status

migrate-up:
	echo "Запуск миграций"
	$(LOAD_ENV) \
	goose -dir migrations postgres "host=$${PG_HOST} port=$${PG_PORT} user=$${PG_USER} password=$${PG_PASSWORD} dbname=$${PG_DB} sslmode=$${PG_SSL_MODE}" up


APP_COMPOSE=docker compose --env-file $(ENV_FILE)

app-up: infra-up
	$(APP_COMPOSE) up -d --build

app-down:
	$(APP_COMPOSE) down

app-rebuild: infra-up
	$(APP_COMPOSE) up -d --build --force-recreate

app-logs:
	$(APP_COMPOSE) logs -f --tail=200

copy-env:
	cp -n $(INFRA_DIR)/.env.example $(ENV_FILE) || true

infra-up: copy-env
	$(INFRA_COMPOSE) up -d

infra-down:
	$(INFRA_COMPOSE) down -v

infra-ps:
	$(INFRA_COMPOSE) ps

infra-logs:
	$(INFRA_COMPOSE) logs -f --tail=200

test:
	go test ./...

test-integration:
	go test -tags=integration -v -count=1 ./internal/storage/postgres/ ./internal/storage/minio/ ./internal/storage/redis/
