-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events.customer_enrollment
(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NUll,
    event_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checked_in_at TIMESTAMPTZ,
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES events.events (id) ON DELETE CASCADE,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE CASCADE,
    CONSTRAINT unique_customer_event UNIQUE (customer_id, event_id),
    CONSTRAINT chk_cancelled_checked_in_null CHECK (
        NOT (is_cancelled AND checked_in_at IS NOT NULL)
    )
);

CREATE TABLE IF NOT EXISTS events.staff
(
    event_id UUID NOT NULL,
    staff_id UUID NOT NULL,
    PRIMARY KEY (event_id, staff_id),
    CONSTRAINT fk_event FOREIGN KEY (event_id) REFERENCES events.events (id) ON DELETE CASCADE,
    CONSTRAINT fk_staff FOREIGN KEY (staff_id) REFERENCES staff.staff (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events.staff;
DROP TABLE IF EXISTS events.customer_enrollment;
-- +goose StatementEnd
