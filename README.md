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
| Docs | Swagger (swaggo) |
| Deploy | Docker, Docker Compose |

## Архитектура

```
cmd/main.go                     — точка входа, graceful shutdown
internal/
  config/                       — конфигурация из env
  logger/                       — структурированный slog-логгер
  auth/                         — JWT access/refresh токены
  routers/
    routers.go                  — Gin engine, DI, маршруты
    middleware/                  — recovery, logging, prometheus, cors, auth, roles
    handlers/
      auth/                     — регистрация, логин, refresh, logout, getMe
      school/                   — CRUD школ
      class/                    — CRUD классов
      user/                     — CRUD пользователей
      interest/                 — CRUD интересов
      profile/                  — профиль ученика
  services/                     — бизнес-логика
  repositories/                 — слой данных (PG, Redis)
  dto/                          — доменные структуры
  errors/                       — AppError, коды ошибок
  storage/                      — PG, Redis, MinIO клиенты
  utils/                        — хеширование, UUID, пароли
migrations/                     — goose SQL-миграции
docs/                           — swagger (авто-генерация)
infra/                          — docker-compose инфраструктуры
```

## Реализованные модули

### Авторизация
- JWT access + refresh токены (SHA-256 хеш, множественные сессии)
- Cache-aside: Redis как горячий кеш, PostgreSQL как источник истины
- Регистрация (самостоятельная + админом с генерацией пароля)
- Логин, Refresh, Logout, LogoutAll, GetMe
- Ролевая модель: `student`, `teacher`, `school_admin`, `super_admin`
- Swagger BearerAuth

### Школы
- CRUD с soft delete и восстановлением
- Фильтрация по имени/городу, пагинация

### Классы
- CRUD вложенный в школу (`/schools/:id/classes`)
- Автоматический расчёт учебного года
- school_admin работает только со своей школой

### Пользователи
- CRUD с фильтрами (школа, класс, роль, ФИО)
- Обновление профиля, смена пароля
- Активация/деактивация, soft delete/restore
- Валидация: teacher и school_admin требуют school_id

### Интересы
- Справочник интересов для LLM-адаптации
- Массовая верификация по списку ID
- 25 дефолтных интересов в seed-миграции

### Профиль ученика
- Автоматическое создание при регистрации студента
- Версионирование (новая запись при изменении)
- Выбор интересов и уровня сложности

### Учебная программа (схема БД)
- `subjects` — справочник предметов
- `programs` — конкретный курс (предмет + класс + автор/учебник)
- `chapters → topics → subtopics` — иерархия с версионированием
- `school_programs` — школа покупает программу (все ученики получают доступ)
- `student_programs` — индивидуал покупает программу сам

### Observability
- Prometheus метрики (`http_requests_total`, `http_request_duration_seconds`)
- Grafana дашборды
- Loki + Promtail для логов
- Структурированное логирование (JSON slog)

## Quick start

```bash
# 1. Скопировать env-файл
cp infra/.env.example infra/.env

# 2. Поднять инфраструктуру + приложение
make app-up

# 3. Применить миграции
make migrate-up
```

Swagger UI: `http://localhost:8000/swagger/index.html`

## Makefile commands

### App

| Команда | Описание |
|---|---|
| `make run` | Запуск Go-приложения локально (swag init + go run) |
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
| `make migrate-down` | Откатить последнюю миграцию |
| `make migrate-reset` | Сбросить все миграции |
| `make migrate-redo` | Сбросить и применить заново |
| `make migrate-status` | Статус миграций |
| `make create-migrations-goose-sql` | Создать SQL-миграцию |

### Tests

| Команда | Описание |
|---|---|
| `make test` | Юнит-тесты |
| `make test-integration` | Интеграционные тесты (PG, Redis, MinIO) |

## API Endpoints

### Infra

| Метод | Путь | Описание |
|---|---|---|
| GET | `/health` | Liveness-проба |
| GET | `/ready` | Readiness-проба (PG + Redis) |
| GET | `/metrics` | Prometheus метрики |
| GET | `/swagger/*` | Swagger UI |

### Auth

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| POST | `/auth/registration` | публичный | Самостоятельная регистрация (student) |
| POST | `/auth/registration/admin` | super_admin, school_admin | Создание пользователя с генерацией пароля |
| POST | `/auth/login` | публичный | Логин по email + пароль |
| GET | `/auth/me` | авторизованный | Данные текущего пользователя |
| POST | `/auth/refresh` | публичный | Обновление пары токенов |
| POST | `/auth/logout` | авторизованный | Отзыв одного refresh-токена |
| POST | `/auth/logout-all` | авторизованный | Отзыв всех токенов + инкремент session_version |

### Users

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/users/{id}` | super_admin, school_admin | Получение пользователя по ID |
| GET | `/users` | super_admin, school_admin | Список пользователей (фильтры: school, class, role, ФИО) |
| PATCH | `/users/me` | авторизованный | Обновление своего профиля |
| PATCH | `/users/{id}` | super_admin, school_admin | Обновление чужого профиля |
| POST | `/users/me/password` | авторизованный | Смена пароля |
| PATCH | `/users/{id}/active` | super_admin, school_admin | Активация/деактивация |
| DELETE | `/users/{id}` | super_admin | Soft delete |
| POST | `/users/{id}/restore` | super_admin | Восстановление |

### Profile (ученик)

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/users/me/profile` | student | Получение профиля |
| PATCH | `/users/me/profile` | student | Обновление интересов и уровня |

### Schools

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/schools/{id}` | авторизованный | Получение школы |
| GET | `/schools` | авторизованный | Список школ (фильтры: name, city) |
| POST | `/schools` | super_admin | Создание школы |
| PATCH | `/schools/{id}` | super_admin, school_admin (своя) | Обновление школы |
| DELETE | `/schools/{id}` | super_admin | Soft delete |
| POST | `/schools/{id}/restore` | super_admin | Восстановление |

### Classes

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/schools/{school_id}/classes/{id}` | авторизованный | Получение класса |
| GET | `/schools/{school_id}/classes` | авторизованный | Список классов школы |
| POST | `/schools/{school_id}/classes` | super_admin, school_admin (своя) | Создание класса |
| PATCH | `/schools/{school_id}/classes/{id}` | super_admin, school_admin (своя) | Обновление класса |
| DELETE | `/schools/{school_id}/classes/{id}` | super_admin, school_admin (своя) | Soft delete |
| POST | `/schools/{school_id}/classes/{id}/restore` | super_admin, school_admin (своя) | Восстановление |

### Interests

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| GET | `/interests` | авторизованный | Список интересов (поиск по name) |
| POST | `/interests` | super_admin | Создание интереса |
| PATCH | `/interests/{id}` | super_admin | Обновление интереса |
| DELETE | `/interests/{id}` | super_admin | Удаление интереса |
| POST | `/interests/verify` | super_admin | Массовая верификация по списку ID |

## Env

Единый env-файл `infra/.env` — все переменные: PostgreSQL (`PG_*`), Redis (`REDIS_*`), MinIO (`MINIO_*`), приложение (`APP_*`), observability.

## Endpoints

| Сервис | Адрес |
|---|---|
| App | `localhost:8000` |
| Swagger | `http://localhost:8000/swagger/index.html` |
| PostgreSQL | `localhost:5433` |
| Redis | `localhost:6380` |
| MinIO API | `localhost:9000` |
| MinIO Console | `http://localhost:9001` |
| Prometheus | `http://localhost:9091` |
| Grafana | `http://localhost:3001` |
| Loki | `http://localhost:3101` |
| Node Exporter | `http://localhost:9101/metrics` |

Grafana и MinIO credentials — в `infra/.env`.
