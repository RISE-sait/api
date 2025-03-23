-- +goose Up
-- +goose StatementBegin
CREATE schema if not exists events;

ALTER TABLE public.customer_enrollment
    SET SCHEMA events;

ALTER TABLE public.staff
    SET SCHEMA events;

ALTER TABlE public.events
    ADD COLUMN capacity int;

ALTER TABLE public.practices
    DROP COLUMN capacity;

ALTER TABLE course.courses
    DROP COLUMN capacity;

ALTER TABLE course.courses
    set schema public;

DROP schema course;

ALTER TABlE public.events
    set schema events;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE events.events
    SET SCHEMA public;

CREATE schema if not exists course;

ALTER TABLE public.courses
    set schema course;

ALTER TABlE public.events
    DROP COLUMN capacity;

ALTER TABLE public.practices
    ADD COLUMN capacity int;

ALTER TABLE course.courses
    ADD COLUMN capacity int;

-- Revert schema changes
ALTER TABLE events.customer_enrollment
    SET SCHEMA public;

ALTER TABLE events.staff
    SET SCHEMA public;

DROP SCHEMA IF EXISTS events;

-- +goose StatementEnd