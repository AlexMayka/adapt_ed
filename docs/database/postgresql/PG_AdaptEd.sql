CREATE TYPE "user_role" AS ENUM (
  'student',
  'teacher',
  'school_admin'
);

CREATE TYPE "difficulty_level" AS ENUM (
  'simple',
  'medium',
  'advanced'
);

CREATE TYPE "topic_status" AS ENUM (
  'not_started',
  'in_progress',
  'understood',
  'needs_review'
);

CREATE TYPE "question_type" AS ENUM (
  'single_choice',
  'multiple_choice',
  'exact_answer'
);

CREATE TYPE "assessment_type" AS ENUM (
  'mini_check',
  'homework',
  'checkpoint',
  'final_exam'
);

CREATE TYPE "assessment_status" AS ENUM (
  'draft',
  'assigned',
  'completed'
);

CREATE TYPE "attempt_status" AS ENUM (
  'in_progress',
  'completed',
  'abandoned'
);

CREATE TYPE "event_kind" AS ENUM (
  'explanation_viewed',
  'mini_check_started',
  'mini_check_passed',
  'mini_check_failed',
  'llm_re_explanation_viewed',
  'level_changed',
  'status_changed'
);

CREATE TABLE "schools" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" text NOT NULL,
  "city" text NOT NULL,
  "logo_key" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  "deleted_at" timestamptz
);

CREATE TABLE "classes" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "school_id" uuid NOT NULL,
  "number_of_class" smallint NOT NULL,
  "suffixes_of_class" text NOT NULL,
  "academic_year_start" date NOT NULL DEFAULT '
        CASE
            WHEN CURRENT_DATE >= make_date(extract(year from CURRENT_DATE)::int, 9, 1)
                THEN make_date(extract(year from CURRENT_DATE)::int, 9, 1)
            ELSE
                make_date(extract(year from CURRENT_DATE)::int - 1, 9, 1)
        END
    ',
  "academic_year_finish" date NOT NULL DEFAULT '
    CASE
      WHEN CURRENT_DATE > make_date(extract(year from CURRENT_DATE)::int - 1, 8, 31)
          THEN make_date(extract(year from CURRENT_DATE)::int, 8, 31)
      ELSE
          make_date(extract(year from CURRENT_DATE)::int, 8, 31)
    END
  ',
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  "deleted_at" timestamptz
);

CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "role" user_role NOT NULL DEFAULT 'student',
  "class_id" uuid,
  "school_id" uuid,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "last_name" text NOT NULL,
  "first_name" text NOT NULL,
  "middle_name" text,
  "avatar_key" text,
  "is_active" bool NOT NULL DEFAULT true,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now()),
  "deleted_at" timestamptz
);

CREATE TABLE "interests" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" text UNIQUE NOT NULL,
  "icon_key" text,
  "is_verified" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "student_profiles" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "user_id" uuid NOT NULL,
  "default_level" difficulty_level NOT NULL DEFAULT 'simple',
  "interests" uuid[] NOT NULL DEFAULT '{}',
  "is_active" bool NOT NULL DEFAULT true,
  "version" serial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "teacher_classes" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "teacher_id" uuid NOT NULL,
  "class_id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "refresh_tokens" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "user_id" uuid NOT NULL,
  "token_hash" text NOT NULL,
  "device_info" text,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "subjects" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" text NOT NULL,
  "slug" text UNIQUE NOT NULL,
  "icon_key" text,
  "color" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "grades" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "subject_id" uuid NOT NULL,
  "grade_number" smallint NOT NULL,
  "textbook" text,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "chapters" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "grade_id" uuid NOT NULL,
  "title" text NOT NULL,
  "sort_order" smallint NOT NULL,
  "icon_key" text,
  "is_active" bool NOT NULL DEFAULT true,
  "version" serial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "topics" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "chapter_id" uuid NOT NULL,
  "title" text NOT NULL,
  "sort_order" smallint NOT NULL,
  "icon_key" text,
  "is_active" bool NOT NULL DEFAULT true,
  "version" serial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "subtopics" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "topic_id" uuid NOT NULL,
  "title" text NOT NULL,
  "sort_order" smallint NOT NULL,
  "icon_key" text,
  "is_active" bool NOT NULL DEFAULT true,
  "version" serial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "base_explanations" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "subtopic_id" uuid NOT NULL,
  "level" difficulty_level NOT NULL,
  "title" text NOT NULL,
  "storage_key" text NOT NULL,
  "is_active" bool NOT NULL DEFAULT true,
  "version" serial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "llm_re_explanations" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "student_profile_id" uuid NOT NULL,
  "subtopic_id" uuid NOT NULL,
  "attempt_number" smallint NOT NULL,
  "prompt_used" text NOT NULL,
  "storage_key" text NOT NULL,
  "model_id" text,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "questions" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "subtopic_id" uuid NOT NULL,
  "type" question_type NOT NULL,
  "difficulty" difficulty_level NOT NULL DEFAULT 'medium',
  "text" text NOT NULL,
  "exact_answer" text,
  "explanation" text,
  "is_active" bool NOT NULL DEFAULT true,
  "version" serial NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "question_options" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "question_id" uuid NOT NULL,
  "text" text NOT NULL,
  "is_correct" bool NOT NULL DEFAULT false,
  "sort_order" smallint NOT NULL DEFAULT 0
);

CREATE TABLE "assessments" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "type" assessment_type NOT NULL,
  "status" assessment_status NOT NULL DEFAULT 'draft',
  "title" text NOT NULL,
  "topic_id" uuid,
  "chapter_id" uuid,
  "grade_id" uuid,
  "class_id" uuid,
  "assigned_by" uuid,
  "due_at" timestamptz,
  "time_limit_sec" int,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "assessment_questions" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "assessment_id" uuid NOT NULL,
  "question_id" uuid NOT NULL,
  "sort_order" smallint NOT NULL DEFAULT 0
);

CREATE TABLE "student_assessment_attempts" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "student_id" uuid NOT NULL,
  "assessment_id" uuid NOT NULL,
  "status" attempt_status NOT NULL DEFAULT 'in_progress',
  "score" numeric(5,2),
  "total_questions" smallint NOT NULL DEFAULT 0,
  "correct_answers" smallint NOT NULL DEFAULT 0,
  "started_at" timestamptz NOT NULL DEFAULT (now()),
  "completed_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "student_answers" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "attempt_id" uuid NOT NULL,
  "question_id" uuid NOT NULL,
  "selected_option_ids" uuid[] DEFAULT '{}',
  "text_answer" text,
  "is_correct" bool NOT NULL,
  "answered_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "student_topic_progress" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "student_id" uuid NOT NULL,
  "subtopic_id" uuid NOT NULL,
  "status" topic_status NOT NULL DEFAULT 'not_started',
  "current_level" difficulty_level NOT NULL DEFAULT 'simple',
  "explanation_attempts" smallint NOT NULL DEFAULT 0,
  "mini_check_attempts" smallint NOT NULL DEFAULT 0,
  "last_attempt_at" timestamptz,
  "understood_at" timestamptz,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "student_topic_events" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "student_id" uuid NOT NULL,
  "subtopic_id" uuid NOT NULL,
  "kind" event_kind NOT NULL,
  "payload" jsonb NOT NULL DEFAULT '{}',
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE UNIQUE INDEX ON "classes" ("school_id", "number_of_class", "suffixes_of_class");

CREATE UNIQUE INDEX "uq_users_email" ON "users" ("email");

CREATE INDEX "idx_users_fullname" ON "users" ("last_name", "first_name", "middle_name", "school_id");

CREATE INDEX "idx_student_profiles_user_id_is_active" ON "student_profiles" ("user_id", "is_active");

CREATE UNIQUE INDEX "uq_teacher_classes" ON "teacher_classes" ("teacher_id", "class_id");

CREATE INDEX "idx_teacher_classes_class_id" ON "teacher_classes" ("class_id");

CREATE UNIQUE INDEX "uq_refresh_tokens_hash" ON "refresh_tokens" ("token_hash");

CREATE INDEX "idx_refresh_tokens_user_id" ON "refresh_tokens" ("user_id");

CREATE INDEX "idx_refresh_tokens_user_active" ON "refresh_tokens" ("user_id", "revoked_at");

CREATE INDEX "idx_refresh_tokens_expires" ON "refresh_tokens" ("expires_at");

CREATE UNIQUE INDEX ON "grades" ("subject_id", "grade_number");

CREATE INDEX ON "chapters" ("grade_id", "sort_order");

CREATE INDEX ON "topics" ("chapter_id", "sort_order");

CREATE INDEX ON "subtopics" ("topic_id", "sort_order");

CREATE INDEX "idx_subtopics_topic_id" ON "subtopics" ("topic_id");

CREATE INDEX ON "base_explanations" ("subtopic_id", "level", "is_active");

CREATE INDEX "idx_llm_re_expl_profile_subtopic" ON "llm_re_explanations" ("student_profile_id", "subtopic_id");

CREATE UNIQUE INDEX "uq_llm_re_expl_attempt" ON "llm_re_explanations" ("student_profile_id", "subtopic_id", "attempt_number");

CREATE INDEX "idx_questions_subtopic_diff_active" ON "questions" ("subtopic_id", "difficulty", "is_active");

CREATE INDEX "idx_question_options_question_id" ON "question_options" ("question_id");

CREATE INDEX "idx_assessments_type" ON "assessments" ("type");

CREATE INDEX "idx_assessments_class_id" ON "assessments" ("class_id");

CREATE INDEX "idx_assessments_topic_id" ON "assessments" ("topic_id");

CREATE INDEX "idx_assessments_chapter_id" ON "assessments" ("chapter_id");

CREATE UNIQUE INDEX "uq_assessment_questions" ON "assessment_questions" ("assessment_id", "question_id");

CREATE INDEX "idx_assessment_questions_assessment" ON "assessment_questions" ("assessment_id");

CREATE INDEX "idx_attempts_student_assessment" ON "student_assessment_attempts" ("student_id", "assessment_id");

CREATE INDEX "idx_attempts_student_id" ON "student_assessment_attempts" ("student_id");

CREATE INDEX "idx_attempts_assessment_id" ON "student_assessment_attempts" ("assessment_id");

CREATE UNIQUE INDEX "uq_student_answers_attempt_question" ON "student_answers" ("attempt_id", "question_id");

CREATE INDEX "idx_student_answers_attempt_id" ON "student_answers" ("attempt_id");

CREATE INDEX "idx_student_answers_question_correct" ON "student_answers" ("question_id", "is_correct");

CREATE UNIQUE INDEX "uq_student_topic_progress" ON "student_topic_progress" ("student_id", "subtopic_id");

CREATE INDEX "idx_stp_student_id" ON "student_topic_progress" ("student_id");

CREATE INDEX "idx_stp_student_status" ON "student_topic_progress" ("student_id", "status");

CREATE INDEX "idx_ste_student_subtopic" ON "student_topic_events" ("student_id", "subtopic_id");

CREATE INDEX "idx_ste_kind" ON "student_topic_events" ("kind");

CREATE INDEX "idx_ste_created_at" ON "student_topic_events" ("created_at");

COMMENT ON TABLE "schools" IS 'Учебное заведение. Верхний уровень организационной структуры. Ученик может быть без школы (индивидуальный).';

COMMENT ON COLUMN "schools"."id" IS 'Идентификатор учебного заведения';

COMMENT ON COLUMN "schools"."name" IS 'Название школы';

COMMENT ON COLUMN "schools"."city" IS 'Город школы';

COMMENT ON COLUMN "schools"."logo_key" IS 'Путь к логотипу в MinIO';

COMMENT ON COLUMN "schools"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "schools"."updated_at" IS 'Дата обновления записи';

COMMENT ON COLUMN "schools"."deleted_at" IS 'Дата удаления записи, soft delete';

COMMENT ON TABLE "classes" IS 'Класс внутри школы (7А, 7Б). Содержит точные даты учебного года для аналитики и автоматической смены класса.';

COMMENT ON COLUMN "classes"."id" IS 'Идентификатор класса';

COMMENT ON COLUMN "classes"."school_id" IS 'Школа класса';

COMMENT ON COLUMN "classes"."number_of_class" IS 'Порядковый номер класса (5, 6, 7, 8...)';

COMMENT ON COLUMN "classes"."suffixes_of_class" IS 'Суффикс класса (А, Б, В...)';

COMMENT ON COLUMN "classes"."academic_year_start" IS 'Дата начала учебного года';

COMMENT ON COLUMN "classes"."academic_year_finish" IS 'Дата окончания учебного года';

COMMENT ON COLUMN "classes"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "classes"."updated_at" IS 'Дата обновления записи';

COMMENT ON COLUMN "classes"."deleted_at" IS 'Дата удаления записи, soft delete';

COMMENT ON TABLE "users" IS 'Все пользователи платформы. Роль определяет возможности. school_id и class_id nullable — индивидуальные ученики без школы.';

COMMENT ON COLUMN "users"."id" IS 'Идентификатор пользователя';

COMMENT ON COLUMN "users"."role" IS 'Роль пользователя';

COMMENT ON COLUMN "users"."class_id" IS 'Класс пользователя';

COMMENT ON COLUMN "users"."school_id" IS 'Школа пользователя';

COMMENT ON COLUMN "users"."email" IS 'Email пользователя';

COMMENT ON COLUMN "users"."password_hash" IS 'Хэш пароля пользователя';

COMMENT ON COLUMN "users"."last_name" IS 'Фамилия пользователя';

COMMENT ON COLUMN "users"."first_name" IS 'Имя пользователя';

COMMENT ON COLUMN "users"."middle_name" IS 'Отчество пользователя';

COMMENT ON COLUMN "users"."avatar_key" IS 'Путь к аватарке в MinIO';

COMMENT ON COLUMN "users"."is_active" IS 'Активен ли пользователь';

COMMENT ON COLUMN "users"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "users"."updated_at" IS 'Дата обновления записи';

COMMENT ON COLUMN "users"."deleted_at" IS 'Дата удаления записи, soft delete';

COMMENT ON TABLE "interests" IS 'Справочник интересов учеников. Используются LLM для генерации аналогий. Новые интересы проходят модерацию (is_verified).';

COMMENT ON COLUMN "interests"."id" IS 'Идентификатор интереса';

COMMENT ON COLUMN "interests"."name" IS 'Название интереса (футбол, Minecraft, робототехника...)';

COMMENT ON COLUMN "interests"."icon_key" IS 'Путь к иконке интереса в MinIO';

COMMENT ON COLUMN "interests"."is_verified" IS 'Прошёл проверку (LLM-модерация / ручная)';

COMMENT ON COLUMN "interests"."created_at" IS 'Дата создания записи';

COMMENT ON TABLE "student_profiles" IS 'Профиль ученика с версионированием. Хранит уровень сложности и интересы для LLM. Новая версия создаётся при изменении уровня или интересов — для отслеживания эффективности промптов.';

COMMENT ON COLUMN "student_profiles"."id" IS 'Идентификатор профиля';

COMMENT ON COLUMN "student_profiles"."user_id" IS 'Пользователь профиля';

COMMENT ON COLUMN "student_profiles"."default_level" IS 'Уровень пользователя (может редактироваться системой)';

COMMENT ON COLUMN "student_profiles"."interests" IS 'Массив ID интересов из таблицы interests';

COMMENT ON COLUMN "student_profiles"."is_active" IS 'Активна ли запись (текущая версия)';

COMMENT ON COLUMN "student_profiles"."version" IS 'Номер версии профиля';

COMMENT ON COLUMN "student_profiles"."created_at" IS 'Дата создания версии';

COMMENT ON TABLE "teacher_classes" IS 'Связь M:N между учителями и классами. Учитель может вести несколько классов, класс может иметь нескольких учителей.';

COMMENT ON COLUMN "teacher_classes"."id" IS 'Идентификатор связи учитель-класс';

COMMENT ON COLUMN "teacher_classes"."teacher_id" IS 'Идентификатор учителя';

COMMENT ON COLUMN "teacher_classes"."class_id" IS 'Идентификатор класса';

COMMENT ON COLUMN "teacher_classes"."created_at" IS 'Дата создания записи';

COMMENT ON TABLE "refresh_tokens" IS 'JWT refresh токены. Хранится хэш, не сам токен. device_info для отображения активных сессий. Отзыв через revoked_at без удаления.';

COMMENT ON COLUMN "refresh_tokens"."id" IS 'Идентификатор токена';

COMMENT ON COLUMN "refresh_tokens"."user_id" IS 'Пользователь';

COMMENT ON COLUMN "refresh_tokens"."token_hash" IS 'Хэш refresh токена';

COMMENT ON COLUMN "refresh_tokens"."device_info" IS 'Информация об устройстве (User-Agent, IP)';

COMMENT ON COLUMN "refresh_tokens"."expires_at" IS 'Срок действия токена';

COMMENT ON COLUMN "refresh_tokens"."revoked_at" IS 'Дата отзыва (NULL = активен)';

COMMENT ON COLUMN "refresh_tokens"."created_at" IS 'Дата создания';

COMMENT ON TABLE "subjects" IS 'Учебный предмет (Информатика, Физика, Математика). Стабильный справочник без версионирования.';

COMMENT ON COLUMN "subjects"."id" IS 'Идентификатор предмета';

COMMENT ON COLUMN "subjects"."name" IS 'Название предмета (Информатика)';

COMMENT ON COLUMN "subjects"."slug" IS 'Slug предмета (informatics)';

COMMENT ON COLUMN "subjects"."icon_key" IS 'Путь к иконке предмета в MinIO';

COMMENT ON COLUMN "subjects"."color" IS 'Цвет предмета в UI (#HEX)';

COMMENT ON COLUMN "subjects"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "subjects"."updated_at" IS 'Дата обновления записи';

COMMENT ON TABLE "grades" IS 'Связка предмет + номер класса. Содержит ссылку на учебник. Стабильный справочник без версионирования.';

COMMENT ON COLUMN "grades"."id" IS 'Идентификатор класса-предмета';

COMMENT ON COLUMN "grades"."subject_id" IS 'Предмет';

COMMENT ON COLUMN "grades"."grade_number" IS 'Номер класса (7)';

COMMENT ON COLUMN "grades"."textbook" IS 'Учебник (Босова Л.Л., ФГОС)';

COMMENT ON COLUMN "grades"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "grades"."updated_at" IS 'Дата обновления записи';

COMMENT ON TABLE "chapters" IS 'Глава учебной программы. Версионируется — при изменении создаётся новая версия.';

COMMENT ON COLUMN "chapters"."id" IS 'Идентификатор версии главы';

COMMENT ON COLUMN "chapters"."grade_id" IS 'Класс-предмет';

COMMENT ON COLUMN "chapters"."title" IS 'Название главы';

COMMENT ON COLUMN "chapters"."sort_order" IS 'Порядок главы (1, 2, 3...)';

COMMENT ON COLUMN "chapters"."icon_key" IS 'Путь к иконке главы в MinIO';

COMMENT ON COLUMN "chapters"."is_active" IS 'Текущая версия';

COMMENT ON COLUMN "chapters"."version" IS 'Номер версии';

COMMENT ON COLUMN "chapters"."created_at" IS 'Дата создания версии';

COMMENT ON TABLE "topics" IS 'Параграф внутри главы. Группирует подтемы. Версионируется.';

COMMENT ON COLUMN "topics"."id" IS 'Идентификатор версии параграфа';

COMMENT ON COLUMN "topics"."chapter_id" IS 'Глава';

COMMENT ON COLUMN "topics"."title" IS 'Название параграфа';

COMMENT ON COLUMN "topics"."sort_order" IS 'Порядок параграфа внутри главы';

COMMENT ON COLUMN "topics"."icon_key" IS 'Путь к иконке параграфа в MinIO';

COMMENT ON COLUMN "topics"."is_active" IS 'Текущая версия';

COMMENT ON COLUMN "topics"."version" IS 'Номер версии';

COMMENT ON COLUMN "topics"."created_at" IS 'Дата создания версии';

COMMENT ON TABLE "subtopics" IS 'Подтема — атомарная единица обучения. К ней привязаны объяснения, вопросы и прогресс ученика. Версионируется.';

COMMENT ON COLUMN "subtopics"."id" IS 'Идентификатор версии подтемы';

COMMENT ON COLUMN "subtopics"."topic_id" IS 'Родительский параграф';

COMMENT ON COLUMN "subtopics"."title" IS 'Название подтемы';

COMMENT ON COLUMN "subtopics"."sort_order" IS 'Порядок подтемы внутри параграфа';

COMMENT ON COLUMN "subtopics"."icon_key" IS 'Путь к иконке подтемы в MinIO';

COMMENT ON COLUMN "subtopics"."is_active" IS 'Текущая версия';

COMMENT ON COLUMN "subtopics"."version" IS 'Номер версии';

COMMENT ON COLUMN "subtopics"."created_at" IS 'Дата создания версии';

COMMENT ON TABLE "base_explanations" IS 'Базовое экспертное объяснение подтемы на определённом уровне сложности. Контент хранится в MinIO как JSON с блоками (text, code, diagram, table). Версионируется.';

COMMENT ON COLUMN "base_explanations"."id" IS 'Идентификатор версии объяснения';

COMMENT ON COLUMN "base_explanations"."subtopic_id" IS 'Подтема';

COMMENT ON COLUMN "base_explanations"."level" IS 'Уровень сложности';

COMMENT ON COLUMN "base_explanations"."title" IS 'Заголовок объяснения';

COMMENT ON COLUMN "base_explanations"."storage_key" IS 'Путь к JSON-файлу с блоками контента в MinIO';

COMMENT ON COLUMN "base_explanations"."is_active" IS 'Текущая версия';

COMMENT ON COLUMN "base_explanations"."version" IS 'Номер версии';

COMMENT ON COLUMN "base_explanations"."created_at" IS 'Дата создания версии';

COMMENT ON TABLE "llm_re_explanations" IS 'LLM-переобъяснение, адаптированное под интересы ученика. Привязано к конкретной версии профиля для отслеживания эффективности. Контент в MinIO.';

COMMENT ON COLUMN "llm_re_explanations"."id" IS 'Идентификатор переобъяснения';

COMMENT ON COLUMN "llm_re_explanations"."student_profile_id" IS 'Версия профиля ученика на момент генерации';

COMMENT ON COLUMN "llm_re_explanations"."subtopic_id" IS 'Подтема';

COMMENT ON COLUMN "llm_re_explanations"."attempt_number" IS 'Номер попытки переобъяснения (1, 2, 3...)';

COMMENT ON COLUMN "llm_re_explanations"."prompt_used" IS 'Промпт, отправленный в LLM';

COMMENT ON COLUMN "llm_re_explanations"."storage_key" IS 'Путь к JSON-файлу с блоками контента в MinIO';

COMMENT ON COLUMN "llm_re_explanations"."model_id" IS 'Модель LLM (gpt-4o, claude-sonnet...)';

COMMENT ON COLUMN "llm_re_explanations"."created_at" IS 'Дата генерации';

COMMENT ON TABLE "questions" IS 'Банк вопросов. Привязан к подтеме. Версионируется — старые версии сохраняются для истории ответов учеников.';

COMMENT ON COLUMN "questions"."id" IS 'Идентификатор версии вопроса';

COMMENT ON COLUMN "questions"."subtopic_id" IS 'Подтема';

COMMENT ON COLUMN "questions"."type" IS 'Тип вопроса';

COMMENT ON COLUMN "questions"."difficulty" IS 'Сложность вопроса';

COMMENT ON COLUMN "questions"."text" IS 'Текст вопроса (markdown)';

COMMENT ON COLUMN "questions"."exact_answer" IS 'Ожидаемый ответ (для exact_answer)';

COMMENT ON COLUMN "questions"."explanation" IS 'Пояснение после ответа';

COMMENT ON COLUMN "questions"."is_active" IS 'Текущая версия';

COMMENT ON COLUMN "questions"."version" IS 'Номер версии';

COMMENT ON COLUMN "questions"."created_at" IS 'Дата создания версии';

COMMENT ON TABLE "question_options" IS 'Варианты ответов для single_choice и multiple_choice вопросов. Привязаны к конкретной версии вопроса.';

COMMENT ON COLUMN "question_options"."id" IS 'Идентификатор варианта ответа';

COMMENT ON COLUMN "question_options"."question_id" IS 'Вопрос (конкретная версия)';

COMMENT ON COLUMN "question_options"."text" IS 'Текст варианта';

COMMENT ON COLUMN "question_options"."is_correct" IS 'Правильный ли вариант';

COMMENT ON COLUMN "question_options"."sort_order" IS 'Порядок отображения';

COMMENT ON TABLE "assessments" IS 'Проверка знаний: мини-проверка (после подтемы), ДЗ (по параграфу/главе), КТ (после главы, адаптивный), итоговый экзамен (весь курс, адаптивный). Может быть назначена учителем или сгенерирована автоматически.';

COMMENT ON COLUMN "assessments"."id" IS 'Идентификатор проверки';

COMMENT ON COLUMN "assessments"."type" IS 'Тип проверки';

COMMENT ON COLUMN "assessments"."status" IS 'Статус проверки';

COMMENT ON COLUMN "assessments"."title" IS 'Название проверки';

COMMENT ON COLUMN "assessments"."topic_id" IS 'Параграф (для mini_check, ДЗ по параграфу)';

COMMENT ON COLUMN "assessments"."chapter_id" IS 'Глава (для КТ, ДЗ по главе)';

COMMENT ON COLUMN "assessments"."grade_id" IS 'Класс-предмет (для итогового экзамена)';

COMMENT ON COLUMN "assessments"."class_id" IS 'Класс (кому назначено)';

COMMENT ON COLUMN "assessments"."assigned_by" IS 'Учитель (NULL = автогенерация)';

COMMENT ON COLUMN "assessments"."due_at" IS 'Дедлайн (для ДЗ)';

COMMENT ON COLUMN "assessments"."time_limit_sec" IS 'Ограничение по времени в секундах';

COMMENT ON COLUMN "assessments"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "assessments"."updated_at" IS 'Дата обновления записи';

COMMENT ON TABLE "assessment_questions" IS 'Связь M:N между проверкой и вопросами. Определяет состав и порядок вопросов в конкретной проверке.';

COMMENT ON COLUMN "assessment_questions"."id" IS 'Идентификатор связи';

COMMENT ON COLUMN "assessment_questions"."assessment_id" IS 'Проверка';

COMMENT ON COLUMN "assessment_questions"."question_id" IS 'Вопрос (конкретная версия)';

COMMENT ON COLUMN "assessment_questions"."sort_order" IS 'Порядок вопроса в проверке';

COMMENT ON TABLE "student_assessment_attempts" IS 'Попытка ученика пройти проверку. Трекает результат, время начала и завершения. Ученик может иметь несколько попыток на одну проверку.';

COMMENT ON COLUMN "student_assessment_attempts"."id" IS 'Идентификатор попытки';

COMMENT ON COLUMN "student_assessment_attempts"."student_id" IS 'Ученик';

COMMENT ON COLUMN "student_assessment_attempts"."assessment_id" IS 'Проверка';

COMMENT ON COLUMN "student_assessment_attempts"."status" IS 'Статус попытки';

COMMENT ON COLUMN "student_assessment_attempts"."score" IS 'Процент 0.00-100.00 (NULL пока in_progress)';

COMMENT ON COLUMN "student_assessment_attempts"."total_questions" IS 'Всего вопросов';

COMMENT ON COLUMN "student_assessment_attempts"."correct_answers" IS 'Правильных ответов';

COMMENT ON COLUMN "student_assessment_attempts"."started_at" IS 'Время начала';

COMMENT ON COLUMN "student_assessment_attempts"."completed_at" IS 'Время завершения';

COMMENT ON COLUMN "student_assessment_attempts"."created_at" IS 'Дата создания записи';

COMMENT ON TABLE "student_answers" IS 'Ответ ученика на конкретный вопрос в рамках попытки. Для choice — массив UUID выбранных вариантов, для exact_answer — текст.';

COMMENT ON COLUMN "student_answers"."id" IS 'Идентификатор ответа';

COMMENT ON COLUMN "student_answers"."attempt_id" IS 'Попытка';

COMMENT ON COLUMN "student_answers"."question_id" IS 'Вопрос (конкретная версия)';

COMMENT ON COLUMN "student_answers"."selected_option_ids" IS 'Выбранные варианты (для choice вопросов)';

COMMENT ON COLUMN "student_answers"."text_answer" IS 'Текстовый ответ (для exact_answer)';

COMMENT ON COLUMN "student_answers"."is_correct" IS 'Правильный ли ответ';

COMMENT ON COLUMN "student_answers"."answered_at" IS 'Время ответа';

COMMENT ON TABLE "student_topic_progress" IS 'Прогресс ученика по подтеме. Мутабельное состояние — обновляется при каждом взаимодействии. Ядро адаптивного цикла: статус, уровень, счётчики попыток.';

COMMENT ON COLUMN "student_topic_progress"."id" IS 'Идентификатор прогресса';

COMMENT ON COLUMN "student_topic_progress"."student_id" IS 'Ученик';

COMMENT ON COLUMN "student_topic_progress"."subtopic_id" IS 'Подтема';

COMMENT ON COLUMN "student_topic_progress"."status" IS 'Статус изучения';

COMMENT ON COLUMN "student_topic_progress"."current_level" IS 'Текущий уровень объяснения';

COMMENT ON COLUMN "student_topic_progress"."explanation_attempts" IS 'Сколько раз смотрел объяснение';

COMMENT ON COLUMN "student_topic_progress"."mini_check_attempts" IS 'Сколько мини-проверок прошёл';

COMMENT ON COLUMN "student_topic_progress"."last_attempt_at" IS 'Дата последней активности';

COMMENT ON COLUMN "student_topic_progress"."understood_at" IS 'Когда подтема была понята';

COMMENT ON COLUMN "student_topic_progress"."created_at" IS 'Дата создания записи';

COMMENT ON COLUMN "student_topic_progress"."updated_at" IS 'Дата обновления записи';

COMMENT ON TABLE "student_topic_events" IS 'Append-only лог событий ученика по подтемам. Не удаляется и не обновляется. Питает аналитику, историю ошибок и карту знаний. Payload — гибкий JSON под каждый тип события.';

COMMENT ON COLUMN "student_topic_events"."id" IS 'Идентификатор события';

COMMENT ON COLUMN "student_topic_events"."student_id" IS 'Ученик';

COMMENT ON COLUMN "student_topic_events"."subtopic_id" IS 'Подтема';

COMMENT ON COLUMN "student_topic_events"."kind" IS 'Тип события';

COMMENT ON COLUMN "student_topic_events"."payload" IS 'Данные события (JSON)';

COMMENT ON COLUMN "student_topic_events"."created_at" IS 'Дата события';

ALTER TABLE "classes" ADD FOREIGN KEY ("school_id") REFERENCES "schools" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "users" ADD FOREIGN KEY ("class_id") REFERENCES "classes" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "users" ADD FOREIGN KEY ("school_id") REFERENCES "schools" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_profiles" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "teacher_classes" ADD FOREIGN KEY ("teacher_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "teacher_classes" ADD FOREIGN KEY ("class_id") REFERENCES "classes" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "refresh_tokens" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "grades" ADD FOREIGN KEY ("subject_id") REFERENCES "subjects" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "chapters" ADD FOREIGN KEY ("grade_id") REFERENCES "grades" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "topics" ADD FOREIGN KEY ("chapter_id") REFERENCES "chapters" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "subtopics" ADD FOREIGN KEY ("topic_id") REFERENCES "topics" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "base_explanations" ADD FOREIGN KEY ("subtopic_id") REFERENCES "subtopics" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "llm_re_explanations" ADD FOREIGN KEY ("student_profile_id") REFERENCES "student_profiles" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "llm_re_explanations" ADD FOREIGN KEY ("subtopic_id") REFERENCES "subtopics" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "questions" ADD FOREIGN KEY ("subtopic_id") REFERENCES "subtopics" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "question_options" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessments" ADD FOREIGN KEY ("topic_id") REFERENCES "topics" ("id") ON DELETE SET NULL ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessments" ADD FOREIGN KEY ("chapter_id") REFERENCES "chapters" ("id") ON DELETE SET NULL ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessments" ADD FOREIGN KEY ("grade_id") REFERENCES "grades" ("id") ON DELETE SET NULL ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessments" ADD FOREIGN KEY ("class_id") REFERENCES "classes" ("id") ON DELETE SET NULL ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessments" ADD FOREIGN KEY ("assigned_by") REFERENCES "users" ("id") ON DELETE SET NULL ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessment_questions" ADD FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "assessment_questions" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_assessment_attempts" ADD FOREIGN KEY ("student_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_assessment_attempts" ADD FOREIGN KEY ("assessment_id") REFERENCES "assessments" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_answers" ADD FOREIGN KEY ("attempt_id") REFERENCES "student_assessment_attempts" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_answers" ADD FOREIGN KEY ("question_id") REFERENCES "questions" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_topic_progress" ADD FOREIGN KEY ("student_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_topic_progress" ADD FOREIGN KEY ("subtopic_id") REFERENCES "subtopics" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_topic_events" ADD FOREIGN KEY ("student_id") REFERENCES "users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "student_topic_events" ADD FOREIGN KEY ("subtopic_id") REFERENCES "subtopics" ("id") ON DELETE CASCADE ON UPDATE NO ACTION DEFERRABLE INITIALLY IMMEDIATE;
