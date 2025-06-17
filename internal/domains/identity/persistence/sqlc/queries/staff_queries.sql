-- name: GetStaffById :one
SELECT s.*, sr.role_name, u.hubspot_id
FROM staff.staff s
         JOIN users.users u ON s.id = u.id
         JOIN staff.staff_roles sr ON s.role_id = sr.id
WHERE u.id = $1;

-- name: GetStaffRoles :many
SELECT *
FROM staff.staff_roles;

-- name: ApproveStaff :one
WITH approved_staff as (SELECT *
                                                FROM staff.pending_staff ps
                                                WHERE ps.id = $1),
         u AS (
             INSERT INTO users.users (country_alpha2_code, gender, first_name, last_name, dob,
                                                                  parent_id, phone, email, has_sms_consent, has_marketing_email_consent)
                         SELECT 
                                         aps.country_alpha2_code,
                                         aps.gender,
                                         aps.first_name,
                                         aps.last_name,
                                         aps.dob,
                                         NULL,
                                         aps.phone,
                                         aps.email,
                                         false,
                                         false
                                                
                         FROM approved_staff aps
                         RETURNING id
                                ),
         s AS (
                INSERT INTO staff.staff (id, role_id, is_active)
                VALUES (
                                (SELECT u.id FROM u),
                                (SELECT aps2.role_id from approved_staff aps2),
                                true
                )
                RETURNING *
         ),
         d AS (
                DELETE FROM staff.pending_staff
                WHERE id = $1
         )
SELECT * FROM s;

-- name: CreatePendingStaff :one
INSERT INTO staff.pending_staff(first_name, last_name, email, gender, dob, phone, country_alpha2_code, role_id)
VALUES ($1, $2, $3, $4, $5, $6, $7,
        (SELECT id FROM staff.staff_roles WHERE role_name = $8))
RETURNING *;

-- name: GetPendingStaffs :many
SELECT id, first_name, last_name, email, gender, phone, country_alpha2_code, role_id, created_at, updated_at, dob
FROM staff.pending_staff;