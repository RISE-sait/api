-- name: CreateWaiverSignedStatus :execrows
WITH prepared_data as (SELECT unnest(@user_id_array::uuid[])    as user_id,
                              unnest(@waiver_url_array::text[]) as waiver_url,
                              unnest(@is_signed_array::bool[])  as is_signed)
INSERT INTO waiver.waiver_signing (user_id, waiver_id, is_signed)
SELECT p.user_id, w.id, p.is_signed
FROM prepared_data p
         LEFT JOIN waiver.waiver w ON w.waiver_url = p.waiver_url;

-- name: GetRequiredWaivers :many
SELECT *
FROM waiver.waiver;