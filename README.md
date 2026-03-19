# AdaptEd — Backend

Адаптивная образовательная платформа, которая объясняет учебный материал с учётом уровня ученика и интересов, проверяет понимание темы и при необходимости перестраивает подачу материала для лучшего усвоения.

## Цель

Не просто «дать ответ», а довести тему до понимания — подбирая объяснение, проверяя усвоение и перестраивая подачу при ошибках.

## Что делает система

1. Даёт ученику объяснение темы, адаптированное под его уровень
2. Задаёт короткую проверку после объяснения
3. Определяет, есть ли непонимание
4. Перестраивает объяснение, если тема не усвоена
5. Сохраняет прогресс и формирует картину знаний

## Для кого

- **Ученик** — инструмент понимания темы
- **Учитель** — сопровождение и аналитика
- **Школа** — цифровая среда для выявления пробелов в знаниях

## Tech stack

| Слой | Технологии |
|---|---|
| App | Go, Gin, JWT |
| Data | PostgreSQL, Redis |
| Storage | MinIO (S3) |
| Observability | Prometheus, Grafana, Loki, Promtail, Node Exporter |
| Docs | Swagger |
| Deploy | Docker, Docker Compose |

## Quick start

```bash
# 1. Скопировать env-файл
cp infra/.env.example infra/.env

# 2. Поднять инфраструктуру + приложение
make app-up
```

## Makefile commands

### App

| Команда | Описание |
|---|---|
| `make run` | Запуск Go-приложения локально (без Docker) |
| `make build` | Сборка бинарника |
| `make app-up` | Поднять infra + собрать и запустить backend в Docker |
| `make app-down` | Остановить backend-контейнер |
| `make app-rebuild` | Пересобрать и перезапустить backend |
| `make app-logs` | Логи backend-контейнера |

### Infra

| Команда | Описание |
|---|---|
| `make infra-up` | Поднять инфраструктуру (PG, Redis, Prometheus, Grafana, Loki) |
| `make infra-down` | Остановить инфраструктуру |
| `make infra-ps` | Статус контейнеров |
| `make infra-logs` | Логи инфраструктуры |

### Migrations

| Команда | Описание |
|---|---|
| `make migrate-up` | Применить миграции |
| `make migrate-status` | Статус миграций |
| `make create-migrations-goose-sql` | Создать SQL-миграцию |

## Env

Единый env-файл `infra/.env` — все переменные: PostgreSQL (`PG_*`), Redis (`REDIS_*`), MinIO (`MINIO_*`), приложение (`APP_*`), observability.

## Endpoints

| Сервис | Адрес |
|---|---|
| App | `localhost:8000` |
| PostgreSQL | `localhost:5433` |
| Redis | `localhost:6380` |
| MinIO API | `localhost:9000` |
| MinIO Console | `http://localhost:9001` |
| Prometheus | `http://localhost:9091` |
| Grafana | `http://localhost:3001` |
| Loki | `http://localhost:3101` |
| Node Exporter | `http://localhost:9101/metrics` |

Grafana и MinIO credentials — в `infra/.env`.
