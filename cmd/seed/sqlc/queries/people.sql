-- name: InsertUsers :many
WITH prepared_data AS (SELECT unnest(@country_alpha2_code_array::text[])            AS country_alpha2_code,
                              unnest(@first_name_array::text[])                     AS first_name,
                              unnest(@last_name_array::text[])                      AS last_name,
                              unnest(@age_array::int[])                             AS age,
                              unnest(@parent_id_array::uuid[]) AS parent_id,
                              unnest(@gender_array::char[])    AS gender,
                              unnest(@phone_array::text[])                          AS phone,
                              unnest(@email_array::text[])                          AS email,
                              unnest(@has_marketing_email_consent_array::boolean[]) AS has_marketing_email_consent,
                              unnest(@has_sms_consent_array::boolean[])             AS has_sms_consent)
INSERT
INTO users.users (country_alpha2_code,
                  first_name,
                  last_name,
                  age,
                  gender,
                  parent_id,
                  phone,
                  email,
                  has_marketing_email_consent,
                  has_sms_consent)
SELECT country_alpha2_code,
       first_name,
       last_name,
       age,
       NULLIF(gender, 'N')                                       AS gender,    -- Replace 'N' with NULL
       NULLIF(parent_id, '00000000-0000-0000-0000-000000000000') AS parent_id, -- Replace default UUID with NULL
       phone,
       email,
       has_marketing_email_consent,
       has_sms_consent
FROM prepared_data
ON CONFLICT DO NOTHING
RETURNING id;


-- name: InsertAthletes :many
INSERT
INTO athletic.athletes (id)
VALUES (unnest(@id_array::uuid[]))
RETURNING id;

-- name: InsertStaffRoles :exec
INSERT INTO staff.staff_roles (role_name)
VALUES ('admin'),
       ('superadmin'),
       ('coach'),
       ('instructor'),
       ('receptionist'),
       ('barber');

-- name: InsertStaff :many
WITH staff_data AS (SELECT e.email,
                           ia.is_active,
                           rn.role_name
                    FROM unnest(@emails::text[]) WITH ORDINALITY AS e(email, idx)
                             JOIN
                         unnest(@is_active_array::bool[]) WITH ORDINALITY AS ia(is_active, idx)
                         ON e.idx = ia.idx
                             JOIN
                         unnest(@role_name_array::text[]) WITH ORDINALITY AS rn(role_name, idx)
                         ON e.idx = rn.idx)
INSERT
INTO staff.staff (id, is_active, role_id)
SELECT u.id,
       sd.is_active,
       sr.id
FROM staff_data sd
         JOIN
     users.users u ON u.email = sd.email
         JOIN
     staff.staff_roles sr ON sr.role_name = sd.role_name
RETURNING id;

-- name: UpdateParents :execrows
UPDATE users.users
SET parent_id = (SELECT id from users.users WHERE email = 'parent@gmail.com')
WHERE email IN ('klintlee1@gmail.com', 'sukhdeepboparai2005@gmail.com');