-- +goose Up
-- +goose StatementBegin
CREATE TABLE users.athletes
(
    id              UUID PRIMARY KEY REFERENCES users.users (id),
    profile_pic_url TEXT,
    wins            INT         NOT NULL DEFAULT 0,                 -- Number of games won
    losses          INT         NOT NULL DEFAULT 0,                 -- Number of games lost
    points          INT         NOT NULL DEFAULT 0,                 -- Total points scored
    steals          INT         NOT NULL DEFAULT 0,                 -- Total steals
    assists         INT         NOT NULL DEFAULT 0,                 -- Total assists
    rebounds        INT         NOT NULL DEFAULT 0,                 -- Total rebounds
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Track last update time
);

ALTER TABLE users.users
    DROP COLUMN profile_pic_url,
    DROP COLUMN wins,
    DROP COLUMN losses,
    DROP COLUMN points,
    DROP COLUMN steals,
    DROP COLUMN assists,
    DROP COLUMN rebounds;

ALTER TABLE users.pending_users
    ADD COLUMN is_parent bool not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.pending_users
    DROP COLUMN is_parent;

ALTER TABLE users.users
    ADD COLUMN profile_pic_url TEXT,
    ADD COLUMN wins            INT NOT NULL DEFAULT 0,
    ADD COLUMN losses          INT NOT NULL DEFAULT 0,
    ADD COLUMN points          INT NOT NULL DEFAULT 0,
    ADD COLUMN steals          INT NOT NULL DEFAULT 0,
    ADD COLUMN assists         INT NOT NULL DEFAULT 0,
    ADD COLUMN rebounds        INT NOT NULL DEFAULT 0;

DROP TABLE users.athletes;
-- +goose StatementEnd
