-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS course_membership
(
    course_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    price_per_booking DECIMAL(4, 2),
    is_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (course_id, membership_id),
    CONSTRAINT fk_course
        FOREIGN KEY (course_id) 
        REFERENCES course.courses (id),
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id) 
        REFERENCES membership.memberships (id),
    CONSTRAINT chk_price_if_eligible
       CHECK (
        (is_eligible = TRUE AND price_per_booking IS NOT NULL) OR
        (is_eligible = FALSE)
    )
);

CREATE TABLE IF NOT EXISTS practice_membership
(
    practice_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    price_per_booking DECIMAL(4, 2) NULL,
    is_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (practice_id, membership_id),
    CONSTRAINT fk_practice
        FOREIGN KEY (practice_id)
        REFERENCES practices (id),
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id) 
        REFERENCES membership.memberships (id),
    CONSTRAINT chk_price_if_eligible
       CHECK (
        (is_eligible = TRUE AND price_per_booking IS NOT NULL) OR
        (is_eligible = FALSE)
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS course_membership;
DROP TABLE IF EXISTS practice_membership;
-- +goose StatementEnd