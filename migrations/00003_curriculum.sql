-- +goose Up

-- ============================================================
-- Домен 2: Учебная программа
-- Таблицы: subjects, programs, chapters, topics, subtopics,
--           school_programs, student_programs
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

COMMENT ON TABLE  subjects              IS 'Учебный предмет (Физика, Математика, Информатика). Стабильный справочник без версионирования.';
COMMENT ON COLUMN subjects.id           IS 'Идентификатор предмета';
COMMENT ON COLUMN subjects.name         IS 'Название предмета (Физика)';
COMMENT ON COLUMN subjects.slug         IS 'Slug предмета (physics)';
COMMENT ON COLUMN subjects.icon_key     IS 'Путь к иконке предмета в MinIO';
COMMENT ON COLUMN subjects.color        IS 'Цвет предмета в UI (#HEX)';
COMMENT ON COLUMN subjects.created_at   IS 'Дата создания записи';
COMMENT ON COLUMN subjects.updated_at   IS 'Дата обновления записи';

-- ---------- programs ----------

CREATE TABLE programs (
    id           uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id   uuid        NOT NULL,
    grade_number smallint    NOT NULL,
    slug         text        UNIQUE NOT NULL,
    title        text        NOT NULL,
    author       text,
    textbook     text,
    description  text,
    is_active    bool        NOT NULL DEFAULT true,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE programs
    ADD CONSTRAINT fk_programs_subject
    FOREIGN KEY (subject_id) REFERENCES subjects (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_programs_subject_grade
    ON programs (subject_id, grade_number);

COMMENT ON TABLE  programs                IS 'Конкретная учебная программа (курс). Один предмет + класс может иметь несколько программ (разные авторы, учебники, уровни). Школы и индивидуалы покупают/подключают программы.';
COMMENT ON COLUMN programs.id             IS 'Идентификатор программы';
COMMENT ON COLUMN programs.subject_id     IS 'Предмет';
COMMENT ON COLUMN programs.grade_number   IS 'Номер класса (7, 8, 9...)';
COMMENT ON COLUMN programs.slug           IS 'Уникальный slug программы (physics-7-peryshkin)';
COMMENT ON COLUMN programs.title          IS 'Название программы (Физика 7 класс, базовый уровень)';
COMMENT ON COLUMN programs.author         IS 'Автор/авторы (Пёрышкин И.М., Иванов А.И.)';
COMMENT ON COLUMN programs.textbook       IS 'Учебник (ФГОС 2024)';
COMMENT ON COLUMN programs.description    IS 'Описание программы';
COMMENT ON COLUMN programs.is_active      IS 'Активна ли программа (доступна для покупки/подключения)';
COMMENT ON COLUMN programs.created_at     IS 'Дата создания записи';
COMMENT ON COLUMN programs.updated_at     IS 'Дата обновления записи';

-- ---------- chapters ----------

CREATE TABLE chapters (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    program_id uuid        NOT NULL,
    title      text        NOT NULL,
    sort_order smallint    NOT NULL,
    icon_key   text,
    is_active  bool        NOT NULL DEFAULT true,
    version    serial      NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE chapters
    ADD CONSTRAINT fk_chapters_program
    FOREIGN KEY (program_id) REFERENCES programs (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_chapters_program_sort
    ON chapters (program_id, sort_order);

COMMENT ON TABLE  chapters              IS 'Глава учебной программы. Версионируется — при изменении создаётся новая версия.';
COMMENT ON COLUMN chapters.id           IS 'Идентификатор версии главы';
COMMENT ON COLUMN chapters.program_id   IS 'Программа';
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

-- ---------- school_programs ----------

CREATE TABLE school_programs (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  uuid        NOT NULL,
    program_id uuid        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE school_programs
    ADD CONSTRAINT fk_school_programs_school
    FOREIGN KEY (school_id) REFERENCES schools (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE school_programs
    ADD CONSTRAINT fk_school_programs_program
    FOREIGN KEY (program_id) REFERENCES programs (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_school_programs
    ON school_programs (school_id, program_id);

CREATE INDEX idx_school_programs_school_id
    ON school_programs (school_id);

CREATE INDEX idx_school_programs_program_id
    ON school_programs (program_id);

COMMENT ON TABLE  school_programs              IS 'Связь M:N школа ↔ программа. Школа покупает пакет — все ученики школы получают доступ к программе.';
COMMENT ON COLUMN school_programs.id           IS 'Идентификатор связи';
COMMENT ON COLUMN school_programs.school_id    IS 'Школа';
COMMENT ON COLUMN school_programs.program_id   IS 'Программа';
COMMENT ON COLUMN school_programs.created_at   IS 'Дата подключения программы';

-- ---------- student_programs ----------

CREATE TABLE student_programs (
    id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL,
    program_id uuid        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE student_programs
    ADD CONSTRAINT fk_student_programs_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE student_programs
    ADD CONSTRAINT fk_student_programs_program
    FOREIGN KEY (program_id) REFERENCES programs (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_student_programs
    ON student_programs (user_id, program_id);

CREATE INDEX idx_student_programs_user_id
    ON student_programs (user_id);

CREATE INDEX idx_student_programs_program_id
    ON student_programs (program_id);

COMMENT ON TABLE  student_programs              IS 'Связь M:N ученик-индивидуал ↔ программа. Ученик без школы покупает программы самостоятельно.';
COMMENT ON COLUMN student_programs.id           IS 'Идентификатор связи';
COMMENT ON COLUMN student_programs.user_id      IS 'Ученик';
COMMENT ON COLUMN student_programs.program_id   IS 'Программа';
COMMENT ON COLUMN student_programs.created_at   IS 'Дата покупки программы';

-- +goose Down

DROP TABLE IF EXISTS student_programs CASCADE;
DROP TABLE IF EXISTS school_programs  CASCADE;
DROP TABLE IF EXISTS subtopics        CASCADE;
DROP TABLE IF EXISTS topics           CASCADE;
DROP TABLE IF EXISTS chapters         CASCADE;
DROP TABLE IF EXISTS programs         CASCADE;
DROP TABLE IF EXISTS subjects         CASCADE;
