-- name: CreateHaircutService :execrows
INSERT INTO haircut.haircut_services (name, description, price, duration_in_min)
VALUES ($1, $2, $3, $4);

-- name: GetHaircutServices :many
SELECT *
FROM haircut.haircut_services;

-- name: UpdateHaircutService :execrows
UPDATE haircut.haircut_services
SET name            = $1,
    description     = $2,
    duration_in_min = $3,
    price           = $4,
    updated_at      = current_timestamp
WHERE id = $5;

-- name: DeleteHaircutService :execrows
DELETE
FROM haircut.haircut_services
WHERE id = $1;