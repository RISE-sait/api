-- name: CreateBarberService :execrows
INSERT INTO haircut.barber_services (barber_id, service_id)
VALUES ($1, $2);

-- name: GetBarberServices :many
SELECT bs.*, (u.first_name || ' ' || u.last_name)::text as barber_name, hs.name as haircut_name
FROM haircut.barber_services bs
         JOIN users.users u ON u.id = bs.barber_id
         JOIN haircut.haircut_services hs ON hs.id = bs.service_id;

-- name: DeleteBarberService :execrows
DELETE
FROM haircut.barber_services
WHERE id = $1;