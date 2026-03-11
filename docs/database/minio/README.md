# AdaptEd — Хранилище файлов MinIO (S3)

## Обзор

MinIO используется как S3-совместимое объектное хранилище для файлового контента платформы: объяснения учебного материала, LLM-переобъяснения, иконки, аватарки и логотипы.

Связь с PostgreSQL — через поля `storage_key` и `*_key` (TEXT), хранящие путь к объекту в MinIO.

## Бакет

| Параметр | Значение |
|----------|----------|
| Название | `adapt_ed` |
| Доступ | Private |
| Версионирование | Отключено (версионирование через PG) |

---

## Структура хранилища

```
adapt_ed/
│
├── explanations/                       Базовые объяснения (экспертные)
│     └── {base_explanation_id}.json    JSON с блоками контента
│
├── llm-explanations/                   LLM-переобъяснения
│     └── {llm_re_explanation_id}.json  JSON с блоками контента
│
├── assets/                             Статические ресурсы платформы
│     ├── schools/
│     │     └── {school_id}/
│     │           └── logo.png          Логотип школы
│     ├── users/
│     │     └── {user_id}/
│     │           └── avatar.png        Аватарка пользователя
│     ├── subjects/
│     │     └── {subject_id}/
│     │           └── icon.svg          Иконка предмета
│     ├── chapters/
│     │     └── {chapter_id}/
│     │           └── icon.svg          Иконка главы
│     ├── topics/
│     │     └── {topic_id}/
│     │           └── icon.svg          Иконка параграфа
│     ├── subtopics/
│     │     └── {subtopic_id}/
│     │           └── icon.svg          Иконка подтемы
│     └── interests/
│           └── {interest_id}/
│                 └── icon.svg          Иконка интереса
│
└── media/                              Медиафайлы (позже)
      └── {uuid}.png                    Картинки в объяснениях и вопросах
```

---

## Связь с PostgreSQL

| Таблица PG | Поле | Путь в MinIO |
|------------|------|-------------|
| `base_explanations` | `storage_key` | `explanations/{id}.json` |
| `llm_re_explanations` | `storage_key` | `llm-explanations/{id}.json` |
| `schools` | `logo_key` | `assets/schools/{id}/logo.png` |
| `users` | `avatar_key` | `assets/users/{id}/avatar.png` |
| `subjects` | `icon_key` | `assets/subjects/{id}/icon.svg` |
| `chapters` | `icon_key` | `assets/chapters/{id}/icon.svg` |
| `topics` | `icon_key` | `assets/topics/{id}/icon.svg` |
| `subtopics` | `icon_key` | `assets/subtopics/{id}/icon.svg` |
| `interests` | `icon_key` | `assets/interests/{id}/icon.svg` |

---

## Формат объяснений (JSON-блоки)

Файлы `explanations/*.json` и `llm-explanations/*.json` имеют единый формат — массив блоков контента:

```json
{
  "blocks": [
    {
      "type": "text",
      "content": "# Заголовок\nТекст объяснения в формате markdown..."
    },
    {
      "type": "code",
      "language": "python",
      "content": "x = 5\nprint(x)"
    },
    {
      "type": "diagram",
      "content": "flowchart TD\n  A[Начало] --> B{Условие}\n  B -- Да --> C[Действие]\n  B -- Нет --> D[Конец]"
    },
    {
      "type": "table",
      "headers": ["Десятичная", "Двоичная", "Восьмеричная"],
      "rows": [
        ["5", "101", "5"],
        ["10", "1010", "12"],
        ["15", "1111", "17"]
      ]
    }
  ]
}
```

### Типы блоков

| Тип | Описание | Статус |
|-----|----------|--------|
| `text` | Markdown-текст | MVP |
| `code` | Код с подсветкой синтаксиса (`language` обязателен) | MVP |
| `diagram` | Mermaid-диаграмма (рендерится на фронте) | MVP |
| `table` | Таблица (`headers` + `rows`) | MVP |
| `image` | Картинка из MinIO (`key`, `alt`) | Позже |
| `formula` | LaTeX-формула (физика, математика) | Позже |
| `video` | Ссылка на видео (`url`) | Позже |

### Будущие типы блоков (примеры)

```json
{
  "type": "image",
  "key": "media/uuid.png",
  "alt": "Схема архитектуры компьютера"
}
```

```json
{
  "type": "formula",
  "content": "E = mc^2"
}
```

```json
{
  "type": "video",
  "url": "https://youtube.com/watch?v=...",
  "title": "Объяснение алгоритмов"
}
```

---

## Конфигурация

Переменные окружения (`.env`):

| Переменная | Описание | По умолчанию |
|-----------|----------|-------------|
| `MINIO_HOST` | Хост MinIO | `localhost` |
| `MINIO_PORT_API` | API порт | `9000` |
| `MINIO_PORT_WEB` | Web-консоль порт | `9001` |
| `MINIO_USER` | Логин | `minio_root` |
| `MINIO_PASSWORD` | Пароль | — |
| `MINIO_BUCKET` | Бакет | `adapt_ed` |
