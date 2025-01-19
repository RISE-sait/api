-- name: CreateStaff :execrows
INSERT INTO staff (id, created_at, role,updated_at, is_active)
VALUES ((SELECT id FROM users WHERE email = $1), CURRENT_TIMESTAMP, $2, CURRENT_TIMESTAMP, $3);

-- name: GetStaffByEmail :one
SELECT oi.name, u.email, s.is_active, s.created_at, s.updated_at, s.role FROM staff s 
JOIN users u ON s.id = u.id 
JOIN user_optional_info oi ON oi.id = u.id WHERE u.email = $1;

-- name: GetAllStaff :many
SELECT * FROM staff;

-- name: UpdateStaff :execrows
UPDATE staff
SET 
    is_active = $1, 
    role = $2, 
    updated_at = CURRENT_TIMESTAMP
WHERE id = (SELECT id FROM users WHERE email = $3);


-- name: DeleteStaff :execrows
DELETE FROM staff WHERE id = (SELECT id FROM users WHERE email = $1);