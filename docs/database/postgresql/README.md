# AdaptEd — Схема базы данных PostgreSQL

## Обзор

Схема БД для адаптивной образовательной платформы AdaptEd.
Покрывает полный цикл адаптивного обучения: от регистрации ученика до аналитики прогресса.

**Онлайн-диаграмма:** [dbdiagram.io](https://dbdiagram.io/d/PG_AdaptEd-69b1d31877d079431b62b990)

## Файлы

| Файл | Описание |
|------|----------|
| `PG_AdaptEd.txt` | Исходный код схемы (DBML) |
| `PG_AdaptEd.sql` | SQL-скрипт (сгенерирован из DBML) |
| `PG_AdaptEd.png` | Диаграмма связей (растр) |
| `PG_AdaptEd.svg` | Диаграмма связей (вектор, масштабируется) |
| `PG_AdaptEd.pdf` | Диаграмма связей (PDF) |

## Статистика

- **22 таблицы** (5 доменов)
- **8 ENUM-типов**
- **PostgreSQL 16**

---

## Домены

### Домен 1: Школы и пользователи

Управление пользователями, школами, классами. Поддерживает индивидуальных учеников без школы.

| Таблица | Назначение |
|---------|-----------|
| `schools` | Учебное заведение (soft delete) |
| `classes` | Класс в школе (7А, 7Б) с датами учебного года |
| `users` | Все пользователи: `student` / `teacher` / `school_admin` |
| `interests` | Справочник интересов для LLM-персонализации (с модерацией) |
| `student_profiles` | Профиль ученика: уровень + интересы (версионируется) |
| `teacher_classes` | Связь M:N учитель <-> класс |
| `refresh_tokens` | JWT refresh токены (хэш, устройство, отзыв) |

**Особенности:**
- `school_id` и `class_id` в `users` — **nullable** (индивидуальные ученики)
- `student_profiles` версионируется: новая запись при изменении уровня/интересов — для отслеживания эффективности LLM-промптов
- `interests` проходят модерацию (`is_verified`) перед использованием в промптах
- Refresh токены хранят хэш, не сам токен; `revoked_at` для отзыва без удаления

### Домен 2: Учебная программа

Иерархия контента: предмет -> класс -> глава -> параграф -> подтема.

| Таблица | Назначение |
|---------|-----------|
| `subjects` | Предмет (Информатика, Физика...) |
| `grades` | Предмет + номер класса + учебник |
| `chapters` | Глава учебной программы (версионируется) |
| `topics` | Параграф внутри главы (версионируется) |
| `subtopics` | Подтема — атомарная единица обучения (версионируется) |

**Иерархия (пример MVP):**
```
Информатика (subject)
  └── 7 класс, Босова (grade)
        └── Гл. 1: Информация и информационные процессы (chapter)
              └── §1: Информация и её свойства (topic)
                    ├── Что такое информация (subtopic)
                    ├── Свойства информации (subtopic)
                    └── Виды информации (subtopic)
```

**Особенности:**
- `subjects` и `grades` — стабильные справочники, без версионирования
- `chapters`, `topics`, `subtopics` — версионируются (`is_active` + `version`)
- `sort_order` определяет порядок отображения
- Добавление нового предмета/класса = новые строки, без изменения схемы

### Домен 3: Объяснения

Контент для обучения: базовые экспертные объяснения и LLM-переобъяснения.

| Таблица | Назначение |
|---------|-----------|
| `base_explanations` | Базовое объяснение подтемы по уровню сложности (версионируется) |
| `llm_re_explanations` | LLM-переобъяснение, адаптированное под интересы ученика |

**Хранение контента:**
- Текст объяснений хранится в **MinIO** как JSON с блоками
- В PG — только метаданные и `storage_key` (путь к файлу)
- MVP типы блоков: `text`, `code`, `diagram`, `table`
- Позже: `image`, `formula`, `video`

**Формат JSON-блоков (MinIO):**
```json
{
  "blocks": [
    {"type": "text", "content": "# Алгоритм\nАлгоритм — это..."},
    {"type": "diagram", "content": "flowchart TD\n  A-->B"},
    {"type": "code", "language": "python", "content": "print('hello')"},
    {"type": "table", "headers": ["Десятичная","Двоичная"], "rows": [["5","101"]]}
  ]
}
```

**LLM-переобъяснения:**
- Привязаны к конкретной версии `student_profiles` (не к пользователю напрямую)
- Хранят `prompt_used` и `model_id` для аудита и итерации промптов
- `attempt_number` — порядковый номер переобъяснения

### Домен 4: Проверки и вопросы

Четыре типа проверок и банк вопросов для адаптивного подбора.

| Таблица | Назначение |
|---------|-----------|
| `questions` | Банк вопросов по подтемам (версионируется) |
| `question_options` | Варианты ответов для choice-вопросов |
| `assessments` | Экземпляр проверки |
| `assessment_questions` | Состав и порядок вопросов в проверке |
| `student_assessment_attempts` | Попытка ученика пройти проверку |
| `student_answers` | Ответы ученика на каждый вопрос |

**Типы проверок:**

| Тип | Когда | Охват | Подбор вопросов |
|-----|-------|-------|-----------------|
| `mini_check` | После подтемы | 2-3 вопроса | Автоматический |
| `homework` | Назначает учитель | Параграф / глава | Учитель или авто |
| `checkpoint` | После главы | Все подтемы главы | Адаптивный (слабые подтемы) |
| `final_exam` | Конец курса | Весь курс | Полностью адаптивный |

**Типы вопросов:**
- `single_choice` — один правильный вариант
- `multiple_choice` — несколько правильных вариантов
- `exact_answer` — точный текстовый/числовой ответ

### Домен 5: Прогресс и аналитика

Отслеживание прогресса и лог событий.

| Таблица | Назначение |
|---------|-----------|
| `student_topic_progress` | Текущее состояние изучения подтемы (мутабельное) |
| `student_topic_events` | Лог событий (append-only) |

**Статусы подтемы:**
```
not_started → in_progress → understood
                   ↓
              needs_review → in_progress → understood
```

**Типы событий (`event_kind`):**

| Событие | Описание | Пример payload |
|---------|---------|---------|
| `explanation_viewed` | Просмотр объяснения | `{"level": "simple"}` |
| `mini_check_started` | Начало мини-проверки | `{"assessment_id": "..."}` |
| `mini_check_passed` | Прошёл мини-проверку | `{"score": 100.0, "attempt_id": "..."}` |
| `mini_check_failed` | Не прошёл мини-проверку | `{"score": 33.3, "attempt_id": "..."}` |
| `llm_re_explanation_viewed` | Просмотр LLM-переобъяснения | `{"re_explanation_id": "..."}` |
| `level_changed` | Смена уровня сложности | `{"from": "simple", "to": "medium"}` |
| `status_changed` | Смена статуса подтемы | `{"from": "in_progress", "to": "understood"}` |

---

## ENUM-типы

| ENUM | Значения | Таблицы |
|------|----------|---------|
| `user_role` | student, teacher, school_admin | `users` |
| `difficulty_level` | simple, medium, advanced | `student_profiles`, `base_explanations`, `questions`, `student_topic_progress` |
| `topic_status` | not_started, in_progress, understood, needs_review | `student_topic_progress` |
| `question_type` | single_choice, multiple_choice, exact_answer | `questions` |
| `assessment_type` | mini_check, homework, checkpoint, final_exam | `assessments` |
| `assessment_status` | draft, assigned, completed | `assessments` |
| `attempt_status` | in_progress, completed, abandoned | `student_assessment_attempts` |
| `event_kind` | 7 типов событий | `student_topic_events` |

---

## Версионирование

Таблицы с версионированием используют паттерн:

| Поле | Тип | Описание |
|------|-----|----------|
| `is_active` | bool | `true` = текущая версия, `false` = архив |
| `version` | serial | Автоинкрементный номер версии |
| `created_at` | timestamptz | Дата создания версии (без `updated_at`) |

При изменении записи создаётся **новая строка**, старая помечается `is_active = false`.

**Версионируемые таблицы:** `student_profiles`, `chapters`, `topics`, `subtopics`, `base_explanations`, `questions`

---