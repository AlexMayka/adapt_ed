-- +goose Up

-- ============================================================
-- Домен 5: Прогресс и аналитика
-- Таблицы: student_topic_progress, student_topic_events
-- ============================================================

-- ---------- student_topic_progress ----------

CREATE TABLE student_topic_progress (
    id                   uuid             PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id           uuid             NOT NULL,
    subtopic_id          uuid             NOT NULL,
    status               topic_status     NOT NULL DEFAULT 'not_started',
    current_level        difficulty_level NOT NULL DEFAULT 'simple',
    explanation_attempts smallint         NOT NULL DEFAULT 0,
    mini_check_attempts  smallint         NOT NULL DEFAULT 0,
    last_attempt_at      timestamptz,
    understood_at        timestamptz,
    created_at           timestamptz      NOT NULL DEFAULT now(),
    updated_at           timestamptz      NOT NULL DEFAULT now()
);

ALTER TABLE student_topic_progress
    ADD CONSTRAINT fk_stp_student
    FOREIGN KEY (student_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE student_topic_progress
    ADD CONSTRAINT fk_stp_subtopic
    FOREIGN KEY (subtopic_id) REFERENCES subtopics (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_student_topic_progress
    ON student_topic_progress (student_id, subtopic_id);

CREATE INDEX idx_stp_student_id
    ON student_topic_progress (student_id);

CREATE INDEX idx_stp_student_status
    ON student_topic_progress (student_id, status);

COMMENT ON TABLE  student_topic_progress                      IS 'Прогресс ученика по подтеме. Мутабельное состояние — обновляется при каждом взаимодействии. Ядро адаптивного цикла: статус, уровень, счётчики попыток.';
COMMENT ON COLUMN student_topic_progress.id                   IS 'Идентификатор прогресса';
COMMENT ON COLUMN student_topic_progress.student_id           IS 'Ученик';
COMMENT ON COLUMN student_topic_progress.subtopic_id          IS 'Подтема';
COMMENT ON COLUMN student_topic_progress.status               IS 'Статус изучения';
COMMENT ON COLUMN student_topic_progress.current_level        IS 'Текущий уровень объяснения';
COMMENT ON COLUMN student_topic_progress.explanation_attempts IS 'Сколько раз смотрел объяснение';
COMMENT ON COLUMN student_topic_progress.mini_check_attempts  IS 'Сколько мини-проверок прошёл';
COMMENT ON COLUMN student_topic_progress.last_attempt_at      IS 'Дата последней активности';
COMMENT ON COLUMN student_topic_progress.understood_at        IS 'Когда подтема была понята';
COMMENT ON COLUMN student_topic_progress.created_at           IS 'Дата создания записи';
COMMENT ON COLUMN student_topic_progress.updated_at           IS 'Дата обновления записи';

-- ---------- student_topic_events ----------

CREATE TABLE student_topic_events (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  uuid        NOT NULL,
    subtopic_id uuid        NOT NULL,
    kind        event_kind  NOT NULL,
    payload     jsonb       NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE student_topic_events
    ADD CONSTRAINT fk_ste_student
    FOREIGN KEY (student_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE student_topic_events
    ADD CONSTRAINT fk_ste_subtopic
    FOREIGN KEY (subtopic_id) REFERENCES subtopics (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_ste_student_subtopic
    ON student_topic_events (student_id, subtopic_id);

CREATE INDEX idx_ste_kind
    ON student_topic_events (kind);

CREATE INDEX idx_ste_created_at
    ON student_topic_events (created_at);

COMMENT ON TABLE  student_topic_events              IS 'Append-only лог событий ученика по подтемам. Не удаляется и не обновляется. Питает аналитику, историю ошибок и карту знаний. Payload — гибкий JSON под каждый тип события.';
COMMENT ON COLUMN student_topic_events.id           IS 'Идентификатор события';
COMMENT ON COLUMN student_topic_events.student_id   IS 'Ученик';
COMMENT ON COLUMN student_topic_events.subtopic_id  IS 'Подтема';
COMMENT ON COLUMN student_topic_events.kind         IS 'Тип события';
COMMENT ON COLUMN student_topic_events.payload      IS 'Данные события (JSON)';
COMMENT ON COLUMN student_topic_events.created_at   IS 'Дата события';

-- +goose Down

DROP TABLE IF EXISTS student_topic_events   CASCADE;
DROP TABLE IF EXISTS student_topic_progress CASCADE;
