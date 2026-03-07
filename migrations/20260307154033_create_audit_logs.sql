-- +goose Up

CREATE TYPE audit_action AS ENUM (
    'register',
    'login',
    'logout',
    'refresh_token',
    'upload_dataset',
    'update_dataset',
    'delete_dataset',
    'process_dataset',
    'download_report',
    'download_source_file'
);

COMMENT ON TYPE audit_action IS 'Тип действия, зафиксированного в журнале аудита';

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    action audit_action NOT NULL,
    entity_type VARCHAR(100),
    entity_id UUID,
    details_json JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_audit_logs_user_id
        FOREIGN KEY (user_id)
            REFERENCES users(id)
            ON DELETE SET NULL
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_entity_type_entity_id ON audit_logs(entity_type, entity_id);

COMMENT ON TABLE audit_logs IS 'Журнал аудита действий пользователей и системных событий';
COMMENT ON COLUMN audit_logs.id IS 'Уникальный идентификатор записи аудита';
COMMENT ON COLUMN audit_logs.user_id IS 'Пользователь, выполнивший действие, если он определён';
COMMENT ON COLUMN audit_logs.action IS 'Тип действия';
COMMENT ON COLUMN audit_logs.entity_type IS 'Тип сущности, к которой относится действие, например dataset';
COMMENT ON COLUMN audit_logs.entity_id IS 'Идентификатор сущности, к которой относится действие';
COMMENT ON COLUMN audit_logs.details_json IS 'Дополнительные детали действия в формате JSON';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP-адрес, с которого было выполнено действие';
COMMENT ON COLUMN audit_logs.user_agent IS 'User-Agent клиента';
COMMENT ON COLUMN audit_logs.created_at IS 'Дата и время фиксации действия';

-- +goose Down
DROP TABLE IF EXISTS audit_logs;
DROP TYPE IF EXISTS audit_action;