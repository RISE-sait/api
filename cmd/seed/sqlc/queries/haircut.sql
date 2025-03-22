-- name: InsertHaircutEvents :exec
WITH prepared_data AS (SELECT unnest(@begin_date_time_array::timestamptz[]) AS begin_date_time,
                              unnest(@end_date_time_array::timestamptz[])   AS end_date_time,
                              unnest(@customer_id_array::uuid[])            AS customer_id,
                              unnest(@barber_email_array::text[])           AS barber_email,
                              unnest(@haircut_name_array::text[])           AS haircut_name),
     haircut_data AS (SELECT pd.begin_date_time,
                             pd.end_date_time,
                             pd.customer_id,
                             ub.id AS barber_id,
                             h.id  as haircut_id
                      FROM prepared_data pd
                               LEFT JOIN
                           users.users ub ON ub.email = pd.barber_email
                               JOIN haircut.haircut_services h ON h.name = pd.haircut_name)
INSERT
INTO haircut.events (begin_date_time, end_date_time, customer_id, barber_id, service_type_id)
SELECT begin_date_time,
       end_date_time,
       customer_id,
       barber_id,
       haircut_id
FROM haircut_data
WHERE customer_id IS NOT NULL
  AND barber_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- name: InsertHaircutServices :exec
WITH prepared_data AS (SELECT unnest(@name_array::text[])           AS name,
                              unnest(@description_array::text[])    AS description,
                              unnest(@price_array::numeric[])       AS price,
                              unnest(@duration_in_min_array::int[]) AS duration_in_min)
INSERT
INTO haircut.haircut_services (name, description, price, duration_in_min)
SELECT name,
       NULLIF(description, ''),
       price,
       duration_in_min
FROM prepared_data;

-- name: InsertBarberServices :exec
WITH prepared_data AS (SELECT unnest(@barber_email_array::text[]) AS barber_email,
                              unnest(@service_name_array::text[]) AS service_name)
INSERT
INTO haircut.barber_services(barber_id, service_id)
SELECT u.id, h.id
FROM prepared_data p
         JOIN users.users u ON p.barber_email = u.email
         JOIN haircut.haircut_services h ON p.service_name = h.name;