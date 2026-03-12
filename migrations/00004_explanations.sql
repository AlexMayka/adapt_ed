-- +goose Up

-- ============================================================
-- Домен 3: Объяснения
-- Таблицы: base_explanations, llm_re_explanations
-- ============================================================

-- ---------- base_explanations ----------

CREATE TABLE base_explanations (
    id          uuid             PRIMARY KEY DEFAULT gen_random_uuid(),
    subtopic_id uuid             NOT NULL,
    level       difficulty_level NOT NULL,
    title       text             NOT NULL,
    storage_key text             NOT NULL,
    is_active   bool             NOT NULL DEFAULT true,
    version     serial           NOT NULL,
    created_at  timestamptz      NOT NULL DEFAULT now()
);

ALTER TABLE base_explanations
    ADD CONSTRAINT fk_base_explanations_subtopic
    FOREIGN KEY (subtopic_id) REFERENCES subtopics (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_base_explanations_subtopic_level_active
    ON base_explanations (subtopic_id, level, is_active);

COMMENT ON TABLE  base_explanations              IS 'Базовое экспертное объяснение подтемы на определённом уровне сложности. Контент хранится в MinIO как JSON с блоками (text, code, diagram, table). Версионируется.';
COMMENT ON COLUMN base_explanations.id           IS 'Идентификатор версии объяснения';
COMMENT ON COLUMN base_explanations.subtopic_id  IS 'Подтема';
COMMENT ON COLUMN base_explanations.level        IS 'Уровень сложности';
COMMENT ON COLUMN base_explanations.title        IS 'Заголовок объяснения';
COMMENT ON COLUMN base_explanations.storage_key  IS 'Путь к JSON-файлу с блоками контента в MinIO';
COMMENT ON COLUMN base_explanations.is_active    IS 'Текущая версия';
COMMENT ON COLUMN base_explanations.version      IS 'Номер версии';
COMMENT ON COLUMN base_explanations.created_at   IS 'Дата создания версии';

-- ---------- llm_re_explanations ----------

CREATE TABLE llm_re_explanations (
    id                 uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_profile_id uuid        NOT NULL,
    subtopic_id        uuid        NOT NULL,
    attempt_number     smallint    NOT NULL,
    prompt_used        text        NOT NULL,
    storage_key        text        NOT NULL,
    model_id           text,
    created_at         timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE llm_re_explanations
    ADD CONSTRAINT fk_llm_re_expl_profile
    FOREIGN KEY (student_profile_id) REFERENCES student_profiles (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE llm_re_explanations
    ADD CONSTRAINT fk_llm_re_expl_subtopic
    FOREIGN KEY (subtopic_id) REFERENCES subtopics (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_llm_re_expl_profile_subtopic
    ON llm_re_explanations (student_profile_id, subtopic_id);

CREATE UNIQUE INDEX uq_llm_re_expl_attempt
    ON llm_re_explanations (student_profile_id, subtopic_id, attempt_number);

COMMENT ON TABLE  llm_re_explanations                    IS 'LLM-переобъяснение, адаптированное под интересы ученика. Привязано к конкретной версии профиля для отслеживания эффективности. Контент в MinIO.';
COMMENT ON COLUMN llm_re_explanations.id                 IS 'Идентификатор переобъяснения';
COMMENT ON COLUMN llm_re_explanations.student_profile_id IS 'Версия профиля ученика на момент генерации';
COMMENT ON COLUMN llm_re_explanations.subtopic_id        IS 'Подтема';
COMMENT ON COLUMN llm_re_explanations.attempt_number     IS 'Номер попытки переобъяснения (1, 2, 3...)';
COMMENT ON COLUMN llm_re_explanations.prompt_used        IS 'Промпт, отправленный в LLM';
COMMENT ON COLUMN llm_re_explanations.storage_key        IS 'Путь к JSON-файлу с блоками контента в MinIO';
COMMENT ON COLUMN llm_re_explanations.model_id           IS 'Модель LLM (gpt-4o, claude-sonnet...)';
COMMENT ON COLUMN llm_re_explanations.created_at         IS 'Дата генерации';

-- +goose Down

DROP TABLE IF EXISTS llm_re_explanations  CASCADE;
DROP TABLE IF EXISTS base_explanations    CASCADE;
