-- +goose Up

-- ============================================================
-- Домен 2: Учебная программа
-- Таблицы: subjects, grades, chapters, topics, subtopics
-- ============================================================

-- ---------- subjects ----------

CREATE TABLE subjects (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text        NOT NULL,
    slug       text        UNIQUE NOT NULL,
    icon_key   text,
    color      text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE  subjects              IS 'Учебный предмет (Информатика, Физика, Математика). Стабильный справочник без версионирования.';
COMMENT ON COLUMN subjects.id           IS 'Идентификатор предмета';
COMMENT ON COLUMN subjects.name         IS 'Название предмета (Информатика)';
COMMENT ON COLUMN subjects.slug         IS 'Slug предмета (informatics)';
COMMENT ON COLUMN subjects.icon_key     IS 'Путь к иконке предмета в MinIO';
COMMENT ON COLUMN subjects.color        IS 'Цвет предмета в UI (#HEX)';
COMMENT ON COLUMN subjects.created_at   IS 'Дата создания записи';
COMMENT ON COLUMN subjects.updated_at   IS 'Дата обновления записи';

-- ---------- grades ----------

CREATE TABLE grades (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id   uuid        NOT NULL,
    grade_number smallint    NOT NULL,
    textbook     text,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE grades
    ADD CONSTRAINT fk_grades_subject
    FOREIGN KEY (subject_id) REFERENCES subjects (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_grades_subject_number
    ON grades (subject_id, grade_number);

COMMENT ON TABLE  grades                IS 'Связка предмет + номер класса. Содержит ссылку на учебник. Стабильный справочник без версионирования.';
COMMENT ON COLUMN grades.id             IS 'Идентификатор класса-предмета';
COMMENT ON COLUMN grades.subject_id     IS 'Предмет';
COMMENT ON COLUMN grades.grade_number   IS 'Номер класса (7)';
COMMENT ON COLUMN grades.textbook       IS 'Учебник (Босова Л.Л., ФГОС)';
COMMENT ON COLUMN grades.created_at     IS 'Дата создания записи';
COMMENT ON COLUMN grades.updated_at     IS 'Дата обновления записи';

-- ---------- chapters ----------

CREATE TABLE chapters (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    grade_id   uuid        NOT NULL,
    title      text        NOT NULL,
    sort_order smallint    NOT NULL,
    icon_key   text,
    is_active  bool        NOT NULL DEFAULT true,
    version    serial      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE chapters
    ADD CONSTRAINT fk_chapters_grade
    FOREIGN KEY (grade_id) REFERENCES grades (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_chapters_grade_sort
    ON chapters (grade_id, sort_order);

COMMENT ON TABLE  chapters              IS 'Глава учебной программы. Версионируется — при изменении создаётся новая версия.';
COMMENT ON COLUMN chapters.id           IS 'Идентификатор версии главы';
COMMENT ON COLUMN chapters.grade_id     IS 'Класс-предмет';
COMMENT ON COLUMN chapters.title        IS 'Название главы';
COMMENT ON COLUMN chapters.sort_order   IS 'Порядок главы (1, 2, 3...)';
COMMENT ON COLUMN chapters.icon_key     IS 'Путь к иконке главы в MinIO';
COMMENT ON COLUMN chapters.is_active    IS 'Текущая версия';
COMMENT ON COLUMN chapters.version      IS 'Номер версии';
COMMENT ON COLUMN chapters.created_at   IS 'Дата создания версии';

-- ---------- topics ----------

CREATE TABLE topics (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id uuid        NOT NULL,
    title      text        NOT NULL,
    sort_order smallint    NOT NULL,
    icon_key   text,
    is_active  bool        NOT NULL DEFAULT true,
    version    serial      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE topics
    ADD CONSTRAINT fk_topics_chapter
    FOREIGN KEY (chapter_id) REFERENCES chapters (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_topics_chapter_sort
    ON topics (chapter_id, sort_order);

COMMENT ON TABLE  topics              IS 'Параграф внутри главы. Группирует подтемы. Версионируется.';
COMMENT ON COLUMN topics.id           IS 'Идентификатор версии параграфа';
COMMENT ON COLUMN topics.chapter_id   IS 'Глава';
COMMENT ON COLUMN topics.title        IS 'Название параграфа';
COMMENT ON COLUMN topics.sort_order   IS 'Порядок параграфа внутри главы';
COMMENT ON COLUMN topics.icon_key     IS 'Путь к иконке параграфа в MinIO';
COMMENT ON COLUMN topics.is_active    IS 'Текущая версия';
COMMENT ON COLUMN topics.version      IS 'Номер версии';
COMMENT ON COLUMN topics.created_at   IS 'Дата создания версии';

-- ---------- subtopics ----------

CREATE TABLE subtopics (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id   uuid        NOT NULL,
    title      text        NOT NULL,
    sort_order smallint    NOT NULL,
    icon_key   text,
    is_active  bool        NOT NULL DEFAULT true,
    version    serial      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE subtopics
    ADD CONSTRAINT fk_subtopics_topic
    FOREIGN KEY (topic_id) REFERENCES topics (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_subtopics_topic_sort
    ON subtopics (topic_id, sort_order);

CREATE INDEX idx_subtopics_topic_id
    ON subtopics (topic_id);

COMMENT ON TABLE  subtopics              IS 'Подтема — атомарная единица обучения. К ней привязаны объяснения, вопросы и прогресс ученика. Версионируется.';
COMMENT ON COLUMN subtopics.id           IS 'Идентификатор версии подтемы';
COMMENT ON COLUMN subtopics.topic_id     IS 'Родительский параграф';
COMMENT ON COLUMN subtopics.title        IS 'Название подтемы';
COMMENT ON COLUMN subtopics.sort_order   IS 'Порядок подтемы внутри параграфа';
COMMENT ON COLUMN subtopics.icon_key     IS 'Путь к иконке подтемы в MinIO';
COMMENT ON COLUMN subtopics.is_active    IS 'Текущая версия';
COMMENT ON COLUMN subtopics.version      IS 'Номер версии';
COMMENT ON COLUMN subtopics.created_at   IS 'Дата создания версии';

-- +goose Down

DROP TABLE IF EXISTS subtopics CASCADE;
DROP TABLE IF EXISTS topics    CASCADE;
DROP TABLE IF EXISTS chapters  CASCADE;
DROP TABLE IF EXISTS grades    CASCADE;
DROP TABLE IF EXISTS subjects  CASCADE;
