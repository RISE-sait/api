-- +goose Up
-- +goose StatementBegin
CREATE TYPE practice_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');

CREATE TABLE practices
(
    id                                UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL UNIQUE,
    description TEXT                     NOT NULL,
    level                             practice_level           NOT NULL DEFAULT 'all',
    capacity                          INT                      NOT NULL,
    created_at                        TIMESTAMP WITH TIME ZONE  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS practices;
DROP TYPE IF EXISTS practice_level;
-- +goose StatementEnd
