-- +goose Up

-- ============================================================
-- Домен 4: Проверки и вопросы
-- Таблицы: questions, question_options, assessments,
--           assessment_questions, student_assessment_attempts,
--           student_answers
-- ============================================================

-- ---------- questions ----------

CREATE TABLE questions (
    id          uuid             PRIMARY KEY DEFAULT gen_random_uuid(),
    subtopic_id uuid             NOT NULL,
    type        question_type    NOT NULL,
    difficulty  difficulty_level NOT NULL DEFAULT 'medium',
    text        text             NOT NULL,
    exact_answer text,
    explanation text,
    is_active   bool             NOT NULL DEFAULT true,
    version     serial           NOT NULL,
    created_at  timestamptz      NOT NULL DEFAULT now()
);

ALTER TABLE questions
    ADD CONSTRAINT fk_questions_subtopic
    FOREIGN KEY (subtopic_id) REFERENCES subtopics (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_questions_subtopic_diff_active
    ON questions (subtopic_id, difficulty, is_active);

COMMENT ON TABLE  questions               IS 'Банк вопросов. Привязан к подтеме. Версионируется — старые версии сохраняются для истории ответов учеников.';
COMMENT ON COLUMN questions.id            IS 'Идентификатор версии вопроса';
COMMENT ON COLUMN questions.subtopic_id   IS 'Подтема';
COMMENT ON COLUMN questions.type          IS 'Тип вопроса';
COMMENT ON COLUMN questions.difficulty    IS 'Сложность вопроса';
COMMENT ON COLUMN questions.text          IS 'Текст вопроса (markdown)';
COMMENT ON COLUMN questions.exact_answer  IS 'Ожидаемый ответ (для exact_answer)';
COMMENT ON COLUMN questions.explanation   IS 'Пояснение после ответа';
COMMENT ON COLUMN questions.is_active     IS 'Текущая версия';
COMMENT ON COLUMN questions.version       IS 'Номер версии';
COMMENT ON COLUMN questions.created_at    IS 'Дата создания версии';

-- ---------- question_options ----------

CREATE TABLE question_options (
    id          uuid     PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id uuid     NOT NULL,
    text        text     NOT NULL,
    is_correct  bool     NOT NULL DEFAULT false,
    sort_order  smallint NOT NULL DEFAULT 0
);

ALTER TABLE question_options
    ADD CONSTRAINT fk_question_options_question
    FOREIGN KEY (question_id) REFERENCES questions (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_question_options_question_id
    ON question_options (question_id);

COMMENT ON TABLE  question_options               IS 'Варианты ответов для single_choice и multiple_choice вопросов. Привязаны к конкретной версии вопроса.';
COMMENT ON COLUMN question_options.id            IS 'Идентификатор варианта ответа';
COMMENT ON COLUMN question_options.question_id   IS 'Вопрос (конкретная версия)';
COMMENT ON COLUMN question_options.text          IS 'Текст варианта';
COMMENT ON COLUMN question_options.is_correct    IS 'Правильный ли вариант';
COMMENT ON COLUMN question_options.sort_order    IS 'Порядок отображения';

-- ---------- assessments ----------

CREATE TABLE assessments (
    id             uuid              PRIMARY KEY DEFAULT gen_random_uuid(),
    type           assessment_type   NOT NULL,
    status         assessment_status NOT NULL DEFAULT 'draft',
    title          text              NOT NULL,
    topic_id       uuid,
    chapter_id     uuid,
    program_id     uuid,
    class_id       uuid,
    assigned_by    uuid,
    due_at         timestamptz,
    time_limit_sec int,
    created_at     timestamptz       NOT NULL DEFAULT now(),
    updated_at     timestamptz       NOT NULL DEFAULT now()
);

ALTER TABLE assessments
    ADD CONSTRAINT fk_assessments_topic
    FOREIGN KEY (topic_id) REFERENCES topics (id)
    ON DELETE SET NULL ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE assessments
    ADD CONSTRAINT fk_assessments_chapter
    FOREIGN KEY (chapter_id) REFERENCES chapters (id)
    ON DELETE SET NULL ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE assessments
    ADD CONSTRAINT fk_assessments_program
    FOREIGN KEY (program_id) REFERENCES programs (id)
    ON DELETE SET NULL ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE assessments
    ADD CONSTRAINT fk_assessments_class
    FOREIGN KEY (class_id) REFERENCES classes (id)
    ON DELETE SET NULL ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE assessments
    ADD CONSTRAINT fk_assessments_assigned_by
    FOREIGN KEY (assigned_by) REFERENCES users (id)
    ON DELETE SET NULL ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_assessments_type
    ON assessments (type);

CREATE INDEX idx_assessments_class_id
    ON assessments (class_id);

CREATE INDEX idx_assessments_topic_id
    ON assessments (topic_id);

CREATE INDEX idx_assessments_chapter_id
    ON assessments (chapter_id);

CREATE INDEX idx_assessments_program_id
    ON assessments (program_id);

COMMENT ON TABLE  assessments                IS 'Проверка знаний: мини-проверка (после подтемы), ДЗ (по параграфу/главе), КТ (после главы, адаптивный), итоговый экзамен (весь курс, адаптивный). Может быть назначена учителем или сгенерирована автоматически.';
COMMENT ON COLUMN assessments.id             IS 'Идентификатор проверки';
COMMENT ON COLUMN assessments.type           IS 'Тип проверки';
COMMENT ON COLUMN assessments.status         IS 'Статус проверки';
COMMENT ON COLUMN assessments.title          IS 'Название проверки';
COMMENT ON COLUMN assessments.topic_id       IS 'Параграф (для mini_check, ДЗ по параграфу)';
COMMENT ON COLUMN assessments.chapter_id     IS 'Глава (для КТ, ДЗ по главе)';
COMMENT ON COLUMN assessments.program_id     IS 'Программа (для итогового экзамена)';
COMMENT ON COLUMN assessments.class_id       IS 'Класс (кому назначено)';
COMMENT ON COLUMN assessments.assigned_by    IS 'Учитель (NULL = автогенерация)';
COMMENT ON COLUMN assessments.due_at         IS 'Дедлайн (для ДЗ)';
COMMENT ON COLUMN assessments.time_limit_sec IS 'Ограничение по времени в секундах';
COMMENT ON COLUMN assessments.created_at     IS 'Дата создания записи';
COMMENT ON COLUMN assessments.updated_at     IS 'Дата обновления записи';

-- ---------- assessment_questions ----------

CREATE TABLE assessment_questions (
    id            uuid     PRIMARY KEY DEFAULT gen_random_uuid(),
    assessment_id uuid     NOT NULL,
    question_id   uuid     NOT NULL,
    sort_order    smallint NOT NULL DEFAULT 0
);

ALTER TABLE assessment_questions
    ADD CONSTRAINT fk_aq_assessment
    FOREIGN KEY (assessment_id) REFERENCES assessments (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE assessment_questions
    ADD CONSTRAINT fk_aq_question
    FOREIGN KEY (question_id) REFERENCES questions (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_assessment_questions
    ON assessment_questions (assessment_id, question_id);

CREATE INDEX idx_assessment_questions_assessment
    ON assessment_questions (assessment_id);

COMMENT ON TABLE  assessment_questions                IS 'Связь M:N между проверкой и вопросами. Определяет состав и порядок вопросов в конкретной проверке.';
COMMENT ON COLUMN assessment_questions.id             IS 'Идентификатор связи';
COMMENT ON COLUMN assessment_questions.assessment_id  IS 'Проверка';
COMMENT ON COLUMN assessment_questions.question_id    IS 'Вопрос (конкретная версия)';
COMMENT ON COLUMN assessment_questions.sort_order     IS 'Порядок вопроса в проверке';

-- ---------- student_assessment_attempts ----------

CREATE TABLE student_assessment_attempts (
    id              uuid           PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      uuid           NOT NULL,
    assessment_id   uuid           NOT NULL,
    status          attempt_status NOT NULL DEFAULT 'in_progress',
    score           numeric(5,2),
    total_questions smallint       NOT NULL DEFAULT 0,
    correct_answers smallint       NOT NULL DEFAULT 0,
    started_at      timestamptz    NOT NULL DEFAULT now(),
    completed_at    timestamptz,
    created_at      timestamptz    NOT NULL DEFAULT now()
);

ALTER TABLE student_assessment_attempts
    ADD CONSTRAINT fk_attempts_student
    FOREIGN KEY (student_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE student_assessment_attempts
    ADD CONSTRAINT fk_attempts_assessment
    FOREIGN KEY (assessment_id) REFERENCES assessments (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE INDEX idx_attempts_student_assessment
    ON student_assessment_attempts (student_id, assessment_id);

CREATE INDEX idx_attempts_student_id
    ON student_assessment_attempts (student_id);

CREATE INDEX idx_attempts_assessment_id
    ON student_assessment_attempts (assessment_id);

COMMENT ON TABLE  student_assessment_attempts                  IS 'Попытка ученика пройти проверку. Трекает результат, время начала и завершения. Ученик может иметь несколько попыток на одну проверку.';
COMMENT ON COLUMN student_assessment_attempts.id               IS 'Идентификатор попытки';
COMMENT ON COLUMN student_assessment_attempts.student_id       IS 'Ученик';
COMMENT ON COLUMN student_assessment_attempts.assessment_id    IS 'Проверка';
COMMENT ON COLUMN student_assessment_attempts.status           IS 'Статус попытки';
COMMENT ON COLUMN student_assessment_attempts.score            IS 'Процент 0.00-100.00 (NULL пока in_progress)';
COMMENT ON COLUMN student_assessment_attempts.total_questions  IS 'Всего вопросов';
COMMENT ON COLUMN student_assessment_attempts.correct_answers  IS 'Правильных ответов';
COMMENT ON COLUMN student_assessment_attempts.started_at       IS 'Время начала';
COMMENT ON COLUMN student_assessment_attempts.completed_at     IS 'Время завершения';
COMMENT ON COLUMN student_assessment_attempts.created_at       IS 'Дата создания записи';

-- ---------- student_answers ----------

CREATE TABLE student_answers (
    id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id          uuid        NOT NULL,
    question_id         uuid        NOT NULL,
    selected_option_ids uuid[]      DEFAULT '{}',
    text_answer         text,
    is_correct          bool        NOT NULL,
    answered_at         timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE student_answers
    ADD CONSTRAINT fk_answers_attempt
    FOREIGN KEY (attempt_id) REFERENCES student_assessment_attempts (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE student_answers
    ADD CONSTRAINT fk_answers_question
    FOREIGN KEY (question_id) REFERENCES questions (id)
    ON DELETE CASCADE ON UPDATE NO ACTION
    DEFERRABLE INITIALLY IMMEDIATE;

CREATE UNIQUE INDEX uq_student_answers_attempt_question
    ON student_answers (attempt_id, question_id);

CREATE INDEX idx_student_answers_attempt_id
    ON student_answers (attempt_id);

CREATE INDEX idx_student_answers_question_correct
    ON student_answers (question_id, is_correct);

COMMENT ON TABLE  student_answers                      IS 'Ответ ученика на конкретный вопрос в рамках попытки. Для choice — массив UUID выбранных вариантов, для exact_answer — текст.';
COMMENT ON COLUMN student_answers.id                   IS 'Идентификатор ответа';
COMMENT ON COLUMN student_answers.attempt_id           IS 'Попытка';
COMMENT ON COLUMN student_answers.question_id          IS 'Вопрос (конкретная версия)';
COMMENT ON COLUMN student_answers.selected_option_ids  IS 'Выбранные варианты (для choice вопросов)';
COMMENT ON COLUMN student_answers.text_answer          IS 'Текстовый ответ (для exact_answer)';
COMMENT ON COLUMN student_answers.is_correct           IS 'Правильный ли ответ';
COMMENT ON COLUMN student_answers.answered_at          IS 'Время ответа';

-- +goose Down

DROP TABLE IF EXISTS student_answers              CASCADE;
DROP TABLE IF EXISTS student_assessment_attempts  CASCADE;
DROP TABLE IF EXISTS assessment_questions         CASCADE;
DROP TABLE IF EXISTS assessments                  CASCADE;
DROP TABLE IF EXISTS question_options             CASCADE;
DROP TABLE IF EXISTS questions                    CASCADE;
