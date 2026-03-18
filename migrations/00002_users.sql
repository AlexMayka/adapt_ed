-- +goose Up

-- ============================================================
-- Домен 1: Школы и пользователи
-- Таблицы: schools, classes, users, interests,
--           student_profiles, teacher_classes, refresh_tokens
-- ============================================================

-- ---------- schools ----------

CREATE TABLE schools (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name            text        NOT NULL,
    city            text        NOT NULL,
    logo_key        text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

COMMENT ON TABLE  schools              IS 'Учебное заведение. Верхний уровень организационной структуры. Ученик может быть без школы (индивидуальный).';
COMMENT ON COLUMN schools.id           IS 'Идентификатор учебного заведения';
COMMENT ON COLUMN schools.name         IS 'Название школы';
COMMENT ON COLUMN schools.city         IS 'Город школы';
COMMENT ON COLUMN schools.logo_key     IS 'Путь к логотипу в MinIO';
COMMENT ON COLUMN schools.created_at   IS 'Дата создания записи';
COMMENT ON COLUMN schools.updated_at   IS 'Дата обновления записи';
COMMENT ON COLUMN schools.deleted_at   IS 'Дата удаления записи, soft delete';

-- ---------- classes ----------

CREATE TABLE classes (
    id                   uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id            uuid        NOT NULL,
    number_of_class      smallint    NOT NULL,
    suffixes_of_class    text        NOT NULL,
    academic_year_start  date        NOT NULL DEFAULT (CASE
                             WHEN CURRENT_DATE >= make_date(extract(year FROM CURRENT_DATE)::int, 9, 1)
                             THEN make_date(extract(year FROM CURRENT_DATE)::int, 9, 1)
                             ELSE make_date(extract(year FROM CURRENT_DATE)::int - 1, 9, 1)
                         END),
    academic_year_finish date        NOT NULL DEFAULT (CASE
                             WHEN CURRENT_DATE > make_date(extract(year FROM CURRENT_DATE)::int - 1, 8, 31)
                             THEN make_date(extract(year FROM CURRENT_DATE)::int, 8, 31)
                             ELSE make_date(extract(year FROM CURRENT_DATE)::int, 8, 31)
                         END),
    created_at           timestamptz NOT NULL DEFAULT now(),
    updated_at           timestamptz NOT NULL DEFAULT now(),
    deleted_at           timestamptz
);

ALTER TABLE classes
    ADD CONSTRAINT fk_classes_school
    FOREIGN KEY (school_id) REFERENCES schools (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_classes_school_number_suffix
    ON classes (school_id, number_of_class, suffixes_of_class);

COMMENT ON TABLE  classes                       IS 'Класс внутри школы (7А, 7Б). Содержит точные даты учебного года для аналитики и автоматической смены класса.';
COMMENT ON COLUMN classes.id                    IS 'Идентификатор класса';
COMMENT ON COLUMN classes.school_id             IS 'Школа класса';
COMMENT ON COLUMN classes.number_of_class       IS 'Порядковый номер класса (5, 6, 7, 8...)';
COMMENT ON COLUMN classes.suffixes_of_class     IS 'Суффикс класса (А, Б, В...)';
COMMENT ON COLUMN classes.academic_year_start   IS 'Дата начала учебного года';
COMMENT ON COLUMN classes.academic_year_finish  IS 'Дата окончания учебного года';
COMMENT ON COLUMN classes.created_at            IS 'Дата создания записи';
COMMENT ON COLUMN classes.updated_at            IS 'Дата обновления записи';
COMMENT ON COLUMN classes.deleted_at            IS 'Дата удаления записи, soft delete';

-- ---------- users ----------

CREATE TABLE users (
    id              uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    role            user_role   NOT NULL DEFAULT 'student',
    class_id        uuid,
    school_id       uuid,
    email           text        NOT NULL,
    password_hash   text        NOT NULL,
    last_name       text        NOT NULL,
    first_name      text        NOT NULL,
    middle_name     text,
    avatar_key      text,
    session_version int         NOT NULL DEFAULT 1,
    is_active       bool        NOT NULL DEFAULT true,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

ALTER TABLE users
    ADD CONSTRAINT fk_users_class
    FOREIGN KEY (class_id) REFERENCES classes (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE users
    ADD CONSTRAINT fk_users_school
    FOREIGN KEY (school_id) REFERENCES schools (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_users_email
    ON users (email);

CREATE INDEX idx_users_fullname
    ON users (last_name, first_name, middle_name, school_id);

COMMENT ON TABLE  users                IS 'Все пользователи платформы. Роль определяет возможности. school_id и class_id nullable — индивидуальные ученики без школы.';
COMMENT ON COLUMN users.id             IS 'Идентификатор пользователя';
COMMENT ON COLUMN users.role           IS 'Роль пользователя';
COMMENT ON COLUMN users.class_id       IS 'Класс пользователя';
COMMENT ON COLUMN users.school_id      IS 'Школа пользователя';
COMMENT ON COLUMN users.email          IS 'Email пользователя';
COMMENT ON COLUMN users.password_hash  IS 'Хэш пароля пользователя';
COMMENT ON COLUMN users.last_name      IS 'Фамилия пользователя';
COMMENT ON COLUMN users.first_name     IS 'Имя пользователя';
COMMENT ON COLUMN users.middle_name    IS 'Отчество пользователя';
COMMENT ON COLUMN users.avatar_key     IS 'Путь к аватарке в MinIO';
COMMENT ON COLUMN users.is_active      IS 'Активен ли пользователь';
COMMENT ON COLUMN users.created_at     IS 'Дата создания записи';
COMMENT ON COLUMN users.updated_at     IS 'Дата обновления записи';
COMMENT ON COLUMN users.deleted_at     IS 'Дата удаления записи, soft delete';

-- ---------- interests ----------

CREATE TABLE interests (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        text        UNIQUE NOT NULL,
    icon_key    text,
    is_verified bool        NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE  interests              IS 'Справочник интересов учеников. Используются LLM для генерации аналогий. Новые интересы проходят модерацию (is_verified).';
COMMENT ON COLUMN interests.id           IS 'Идентификатор интереса';
COMMENT ON COLUMN interests.name         IS 'Название интереса (футбол, Minecraft, робототехника...)';
COMMENT ON COLUMN interests.icon_key     IS 'Путь к иконке интереса в MinIO';
COMMENT ON COLUMN interests.is_verified  IS 'Прошёл проверку (LLM-модерация / ручная)';
COMMENT ON COLUMN interests.created_at   IS 'Дата создания записи';

-- ---------- student_profiles ----------

CREATE TABLE student_profiles (
    id            uuid             PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid             NOT NULL,
    default_level difficulty_level NOT NULL DEFAULT 'simple',
    interests     uuid[]           NOT NULL DEFAULT '{}',
    is_active     bool             NOT NULL DEFAULT true,
    version       serial           NOT NULL,
    created_at    timestamptz      NOT NULL DEFAULT now()
);

ALTER TABLE student_profiles
    ADD CONSTRAINT fk_student_profiles_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_student_profiles_user_id_is_active
    ON student_profiles (user_id, is_active);

COMMENT ON TABLE  student_profiles                IS 'Профиль ученика с версионированием. Хранит уровень сложности и интересы для LLM. Новая версия создаётся при изменении уровня или интересов — для отслеживания эффективности промптов.';
COMMENT ON COLUMN student_profiles.id             IS 'Идентификатор профиля';
COMMENT ON COLUMN student_profiles.user_id        IS 'Пользователь профиля';
COMMENT ON COLUMN student_profiles.default_level  IS 'Уровень пользователя (может редактироваться системой)';
COMMENT ON COLUMN student_profiles.interests      IS 'Массив ID интересов из таблицы interests';
COMMENT ON COLUMN student_profiles.is_active      IS 'Активна ли запись (текущая версия)';
COMMENT ON COLUMN student_profiles.version        IS 'Номер версии профиля';
COMMENT ON COLUMN student_profiles.created_at     IS 'Дата создания версии';

-- ---------- teacher_classes ----------

CREATE TABLE teacher_classes (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id  uuid        NOT NULL,
    class_id    uuid        NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE teacher_classes
    ADD CONSTRAINT fk_teacher_classes_teacher
    FOREIGN KEY (teacher_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE teacher_classes
    ADD CONSTRAINT fk_teacher_classes_class
    FOREIGN KEY (class_id) REFERENCES classes (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_teacher_classes
    ON teacher_classes (teacher_id, class_id);

CREATE INDEX idx_teacher_classes_class_id
    ON teacher_classes (class_id);

COMMENT ON TABLE  teacher_classes              IS 'Связь M:N между учителями и классами. Учитель может вести несколько классов, класс может иметь нескольких учителей.';
COMMENT ON COLUMN teacher_classes.id           IS 'Идентификатор связи учитель-класс';
COMMENT ON COLUMN teacher_classes.teacher_id   IS 'Идентификатор учителя';
COMMENT ON COLUMN teacher_classes.class_id     IS 'Идентификатор класса';
COMMENT ON COLUMN teacher_classes.created_at   IS 'Дата создания записи';

-- ---------- refresh_tokens ----------

CREATE TABLE refresh_tokens (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL,
    token_hash  text        NOT NULL,
    device_info text,
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE refresh_tokens
    ADD CONSTRAINT fk_refresh_tokens_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_refresh_tokens_hash
    ON refresh_tokens (token_hash);

CREATE INDEX idx_refresh_tokens_user_id
    ON refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_user_active
    ON refresh_tokens (user_id, revoked_at);

CREATE INDEX idx_refresh_tokens_expires
    ON refresh_tokens (expires_at);

COMMENT ON TABLE  refresh_tokens              IS 'JWT refresh токены. Хранится хэш, не сам токен. device_info для отображения активных сессий. Отзыв через revoked_at без удаления.';
COMMENT ON COLUMN refresh_tokens.id           IS 'Идентификатор токена';
COMMENT ON COLUMN refresh_tokens.user_id      IS 'Пользователь';
COMMENT ON COLUMN refresh_tokens.token_hash   IS 'Хэш refresh токена';
COMMENT ON COLUMN refresh_tokens.device_info  IS 'Информация об устройстве (User-Agent, IP)';
COMMENT ON COLUMN refresh_tokens.expires_at   IS 'Срок действия токена';
COMMENT ON COLUMN refresh_tokens.revoked_at   IS 'Дата отзыва (NULL = активен)';
COMMENT ON COLUMN refresh_tokens.created_at   IS 'Дата создания';

-- +goose Down

DROP TABLE IF EXISTS refresh_tokens  CASCADE;
DROP TABLE IF EXISTS teacher_classes  CASCADE;
DROP TABLE IF EXISTS student_profiles CASCADE;
DROP TABLE IF EXISTS interests        CASCADE;
DROP TABLE IF EXISTS users            CASCADE;
DROP TABLE IF EXISTS classes          CASCADE;
DROP TABLE IF EXISTS schools          CASCADE;
