-- +goose Up

CREATE TYPE user_role AS ENUM (
    'student',
    'teacher',
    'school_admin',
    'super_admin'
);

CREATE TYPE difficulty_level AS ENUM (
    'simple',
    'medium',
    'advanced'
);

CREATE TYPE topic_status AS ENUM (
    'not_started',
    'in_progress',
    'understood',
    'needs_review'
);

CREATE TYPE question_type AS ENUM (
    'single_choice',
    'multiple_choice',
    'exact_answer'
);

CREATE TYPE assessment_type AS ENUM (
    'mini_check',
    'homework',
    'checkpoint',
    'final_exam'
);

CREATE TYPE assessment_status AS ENUM (
    'draft',
    'assigned',
    'completed'
);

CREATE TYPE attempt_status AS ENUM (
    'in_progress',
    'completed',
    'abandoned'
);

CREATE TYPE event_kind AS ENUM (
    'explanation_viewed',
    'mini_check_started',
    'mini_check_passed',
    'mini_check_failed',
    'llm_re_explanation_viewed',
    'level_changed',
    'status_changed'
);

-- +goose Down

DROP TYPE IF EXISTS event_kind CASCADE;
DROP TYPE IF EXISTS attempt_status CASCADE;
DROP TYPE IF EXISTS assessment_status CASCADE;
DROP TYPE IF EXISTS assessment_type CASCADE;
DROP TYPE IF EXISTS question_type CASCADE;
DROP TYPE IF EXISTS topic_status CASCADE;
DROP TYPE IF EXISTS difficulty_level CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
