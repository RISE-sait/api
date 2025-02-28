-- +goose Up
-- +goose StatementBegin
CREATE TYPE practice_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');

CREATE TABLE practices
(
    id                                UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name                              VARCHAR(50)              NOT NULL UNIQUE,
    description                       TEXT,
    level                             practice_level           NOT NULL DEFAULT 'all',
    should_email_booking_notification BOOLEAN                           DEFAULT True,
    capacity                          INT                      NOT NULL,
    start_date                        TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date                          TIMESTAMP WITH TIME ZONE,
    created_at                        TIMESTAMP WITH TIME ZONE  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                        TIMESTAMP WITH TIME ZONE  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT end_after_start CHECK (end_date > start_date)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS practices;
DROP TYPE IF EXISTS practice_level;
-- +goose StatementEnd
