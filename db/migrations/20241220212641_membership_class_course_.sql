-- +goose Up
-- +goose StatementBegin
CREATE TABLE course_membership (
    course_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    price_per_booking DECIMAL(4, 2) NULL,
    is_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (course_id, membership_id),
    CONSTRAINT fk_course
        FOREIGN KEY (course_id) 
        REFERENCES courses (id),
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id) 
        REFERENCES memberships (id),
    CONSTRAINT chk_price_if_eligible
       CHECK (
        (is_eligible = TRUE AND price_per_booking IS NOT NULL) OR
        (is_eligible = FALSE)
    )
);

CREATE TABLE class_membership (
    class_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    price_per_booking DECIMAL(4, 2) NULL,
    is_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (class_id, membership_id),
    CONSTRAINT fk_class
        FOREIGN KEY (class_id) 
        REFERENCES classes (id),
    CONSTRAINT fk_membership
        FOREIGN KEY (membership_id) 
        REFERENCES memberships (id),
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
DROP TABLE IF EXISTS class_membership;
-- +goose StatementEnd