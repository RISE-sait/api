-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS haircut;

CREATE TABLE haircut.haircut_services
(
    id              uuid PRIMARY KEY       DEFAULT gen_random_uuid(),
    name            varchar(50)   NOT NULL UNIQUE,
    description     varchar(255),
    price           decimal(5, 2) NOT NULL,
    duration_in_min int           NOT NULL CHECK (duration_in_min > 0),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE haircut.barber_services
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    barber_id  uuid        NOT NULL,
    service_id uuid        NOT NULL,
    CONSTRAINT fk_barber
        FOREIGN KEY (barber_id)
            REFERENCES users.staff (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_service
        FOREIGN KEY (service_id)
            REFERENCES haircut.haircut_services (id)
            ON DELETE CASCADE,
    CONSTRAINT unique_barber_service UNIQUE (barber_id, service_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE barber.barber_events
    ADD COLUMN service_type_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

ALTER TABLE barber.barber_events
    ADD CONSTRAINT fk_service_type
        FOREIGN KEY (service_type_id)
            REFERENCES haircut.haircut_services (id)
            ON DELETE SET NULL;

ALTER TABLE barber.barber_events
    RENAME TO events;

ALTER TABLE barber.events
    SET SCHEMA haircut;

DROP SCHEMA IF EXISTS barber;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS barber;

ALTER TABLE haircut.events
    SET SCHEMA barber;

ALTER TABLE barber.events
    rename to barber_events;

ALTER TABLE barber.barber_events
    DROP CONSTRAINT IF EXISTS fk_service_type;

ALTER TABLE barber.barber_events
    DROP COLUMN IF EXISTS service_type_id;

DROP TABLE IF EXISTS haircut.barber_services;
DROP TABLE IF EXISTS haircut.haircut_services;

DROP SCHEMA IF EXISTS haircut;
-- +goose StatementEnd
