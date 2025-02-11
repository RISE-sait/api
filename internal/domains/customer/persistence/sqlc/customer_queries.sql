-- name: GetCustomers :many
SELECT
    cu.user_id as customer_id,
    oi.first_name,
    oi.last_name,
    oi.phone,

    u.email,
    ce.is_cancelled as is_event_booking_cancelled,
    ce.checked_in_at
FROM
    customers cu
    JOIN users u ON cu.user_id = u.id
    JOIN user_optional_info oi ON u.id = oi.id
    JOIN customer_events ce ON cu.user_id = ce.customer_id
WHERE (
        ce.event_id = sqlc.narg ('event_id')
        OR sqlc.narg ('event_id') IS NULL
    );