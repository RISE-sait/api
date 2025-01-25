-- name: CreateStaff :execrows
INSERT INTO staff (role, is_active) VALUES ($1, $2);