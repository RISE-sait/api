-- +goose Up
-- +goose StatementBegin

CREATE schema if not exists program;

CREATE TYPE program.program_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');
CREATE TYPE program.program_type AS ENUM ('practice', 'course','other');

CREATE TABLE IF NOT EXISTS program.programs
(
    id          UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name        VARCHAR(150)             NOT NULL UNIQUE,
    description TEXT                     NOT NULL,
    level       program.program_level    NOT NULL DEFAULT 'all',
    type        program.program_type     NOT NULL,
    capacity int,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS program.games
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    win_team   uuid not null,
    lose_team  uuid not null,
    win_score  int  not null    default 0,
    lose_score int  not null    default 0,
    CONSTRAINT fk_program_id FOREIGN KEY (id) REFERENCES program.programs (id) ON DELETE cascade,
    CONSTRAINT fk_win_team FOREIGN KEY (win_team) REFERENCES athletic.teams (id) ON DELETE cascade,
    CONSTRAINT fk_lose_team FOREIGN KEY (lose_team) REFERENCES athletic.teams (id) ON DELETE cascade
);

CREATE TABLE IF NOT EXISTS program.program_membership
(
    program_id              UUID                     NOT NULL,
    membership_id           UUID,
    stripe_program_price_id VARCHAR(50)              NOT NULL, -- Stripe Price ID for this program-membership combo
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_program
        FOREIGN KEY (program_id)
            REFERENCES program.programs (id) ON DELETE CASCADE,
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id)
            REFERENCES membership.memberships (id) ON DELETE CASCADE,
    CONSTRAINT unique_stripe_program_price_id
        UNIQUE (stripe_program_price_id)
);

CREATE TABLE IF NOT EXISTS program.customer_enrollment
(
    id           UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    customer_id  UUID                     NOT NUll,
    program_id   UUID                     NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_cancelled bool                     NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_program FOREIGN KEY (program_id) REFERENCES program.programs (id) ON DELETE CASCADE,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE CASCADE,
    CONSTRAINT unique_customer_program UNIQUE (customer_id, program_id)
);

CREATE UNIQUE INDEX unique_program_membership_not_null
    ON program.program_membership (program_id, membership_id)
    WHERE membership_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS program cascade;
-- +goose StatementEnd