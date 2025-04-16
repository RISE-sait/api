-- name: UpdateUserInfo :execrows
UPDATE users.users
SET first_name                  = sqlc.arg('first_name'),
    last_name                   = sqlc.arg('last_name'),
    email                       = sqlc.arg('email'),
    phone                       = sqlc.arg('phone'),
    dob                         = sqlc.arg('dob'),
    country_alpha2_code         = sqlc.arg('country_alpha2_code'),
    has_marketing_email_consent = sqlc.arg('has_marketing_email_consent'),
    has_sms_consent             = sqlc.arg('has_sms_consent'),
    gender                      = sqlc.arg('gender'),
    parent_id                   = $1,
    updated_at                  = current_timestamp
WHERE id = sqlc.arg('id');