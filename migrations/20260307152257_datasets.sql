-- +goose Up

CREATE TYPE dataset_status AS ENUM (
    'uploaded',
    'processing',
    'processed',
    'failed'
);

COMMENT ON TYPE dataset_status IS 'Статусы обработки загруженного файла';

CREATE TYPE dataset_file_type AS ENUM (
    'csv',
    'xlsx',
    'xls'
);

COMMENT ON TYPE dataset_file_type IS 'Тип загруженного файла';

CREATE TYPE signal_severity AS ENUM (
    'low',
    'medium',
    'high',
    'critical'
);

COMMENT ON TYPE signal_severity IS 'Уровень критичности сигнала';

CREATE TABLE datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,

    original_filename VARCHAR(255) NOT NULL,
    s3_key TEXT NOT NULL,
    file_type dataset_file_type NOT NULL,

    rows_count INTEGER,

    status dataset_status NOT NULL DEFAULT 'uploaded',

    total_leads INTEGER,
    won INTEGER,
    lost INTEGER,
    open INTEGER,
    conversion NUMERIC,
    estimated_loss NUMERIC,

    result_json JSONB,

    pdf_s3_key TEXT,
    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_datasets_user_id
      FOREIGN KEY (user_id)
          REFERENCES users(id)
          ON DELETE CASCADE
);

CREATE INDEX idx_datasets_user_id ON datasets(user_id);
CREATE INDEX idx_datasets_status ON datasets(status);
CREATE INDEX idx_datasets_created_at ON datasets(created_at);
CREATE INDEX idx_datasets_status_created_at ON datasets(status, created_at);

COMMENT ON TABLE datasets IS 'Загруженные пользователем файлы и результаты их анализа';

COMMENT ON COLUMN datasets.id IS 'Уникальный идентификатор набора данных';
COMMENT ON COLUMN datasets.user_id IS 'Пользователь, загрузивший файл';
COMMENT ON COLUMN datasets.original_filename IS 'Оригинальное имя файла';
COMMENT ON COLUMN datasets.s3_key IS 'Ключ объекта в S3 хранилище';
COMMENT ON COLUMN datasets.file_type IS 'Тип загруженного файла';
COMMENT ON COLUMN datasets.rows_count IS 'Количество строк после чтения файла';
COMMENT ON COLUMN datasets.status IS 'Текущий статус обработки файла';

COMMENT ON COLUMN datasets.total_leads IS 'Общее количество лидов';
COMMENT ON COLUMN datasets.won IS 'Количество успешных сделок';
COMMENT ON COLUMN datasets.lost IS 'Количество потерянных сделок';
COMMENT ON COLUMN datasets.open IS 'Количество открытых лидов';
COMMENT ON COLUMN datasets.conversion IS 'Конверсия продаж';
COMMENT ON COLUMN datasets.estimated_loss IS 'Оценка потенциальных потерь';

COMMENT ON COLUMN datasets.result_json IS 'JSON с детализированными результатами анализа';
COMMENT ON COLUMN datasets.pdf_s3_key IS 'Ключ PDF отчета в S3';
COMMENT ON COLUMN datasets.error_message IS 'Ошибка обработки';
COMMENT ON COLUMN datasets.created_at IS 'Дата создания записи';
COMMENT ON COLUMN datasets.updated_at IS 'Дата последнего обновления записи';

-- +goose Down

DROP TABLE IF EXISTS datasets;

DROP TYPE IF EXISTS signal_severity;
DROP TYPE IF EXISTS dataset_file_type;
DROP TYPE IF EXISTS dataset_status;