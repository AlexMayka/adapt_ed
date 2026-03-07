-- +goose Up
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT fk_refresh_tokens_users_id
        FOREIGN KEY (user_id)
            REFERENCES users(id)
            ON DELETE CASCADE
);

COMMENT ON TABLE refresh_tokens IS 'Токены пользователей системы';
COMMENT ON COLUMN refresh_tokens.id IS 'Уникальный идентификатор токена';
COMMENT ON COLUMN refresh_tokens.user_id IS 'ID пользователя';
COMMENT ON COLUMN refresh_tokens.token_hash IS 'Токен refresh пользователя';
COMMENT ON COLUMN refresh_tokens.expires_at IS 'Срок жизни пользователя';
COMMENT ON COLUMN refresh_tokens.created_at IS 'Дата создания токена';

CREATE INDEX idx_token_user_id ON refresh_tokens(user_id);
CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
