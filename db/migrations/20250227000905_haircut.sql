-- +goose Up

-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

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
            REFERENCES staff.staff (id)
            ON DELETE CASCADE,
    CONSTRAINT fk_service
        FOREIGN KEY (service_id)
            REFERENCES haircut.haircut_services (id)
            ON DELETE CASCADE,
    CONSTRAINT unique_barber_service UNIQUE (barber_id, service_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp with time zone
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS haircut.events
(
    id              UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    begin_date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date_time   TIMESTAMP WITH TIME ZONE NOT NULL,
    customer_id     UUID                     NOT NULL,
    barber_id       UUID                     NOT NULL,
    service_type_id uuid                     NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    created_at      TIMESTAMPTZ              NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ              NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE cascade,
    CONSTRAINT fk_barber FOREIGN KEY (barber_id) REFERENCES staff.staff (id) ON DELETE cascade,
    CONSTRAINT check_end_time CHECK (end_date_time > begin_date_time), -- Prevent invalid schedules
    CONSTRAINT unique_schedule
        EXCLUDE USING GIST (
        barber_id WITH =,
        tstzrange(begin_date_time, end_date_time, '[]') WITH &&
        ),
    CONSTRAINT fk_service_type
        FOREIGN KEY (service_type_id)
            REFERENCES haircut.haircut_services (id)
            ON DELETE SET NULL
);

-- +goose StatementEnd

-- +goose Down
DROP SCHEMA IF EXISTS haircut cascade;