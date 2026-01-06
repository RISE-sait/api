-- name: GetEventEnrolledCustomers :many
-- Get all enrolled customers for an event with their email and push token status
SELECT
    u.id,
    u.first_name,
    u.last_name,
    u.email,
    EXISTS(SELECT 1 FROM notifications.push_tokens pt WHERE pt.user_id = u.id) AS has_push_token
FROM events.customer_enrollment ce
JOIN users.users u ON ce.customer_id = u.id
WHERE ce.event_id = $1
  AND ce.is_cancelled = false
  AND ce.payment_status = 'paid'
ORDER BY u.last_name, u.first_name;

-- name: GetEventEnrolledCustomerEmails :many
-- Get emails of all enrolled customers for an event (for email sending)
SELECT
    u.id,
    u.first_name,
    u.email
FROM events.customer_enrollment ce
JOIN users.users u ON ce.customer_id = u.id
WHERE ce.event_id = $1
  AND ce.is_cancelled = false
  AND ce.payment_status = 'paid'
  AND u.email IS NOT NULL
  AND u.email != '';

-- name: GetEventEnrolledCustomerPushTokens :many
-- Get push tokens of all enrolled customers for an event (for push notifications)
SELECT
    u.id AS user_id,
    u.first_name,
    pt.expo_push_token,
    pt.device_type
FROM events.customer_enrollment ce
JOIN users.users u ON ce.customer_id = u.id
JOIN notifications.push_tokens pt ON pt.user_id = u.id
WHERE ce.event_id = $1
  AND ce.is_cancelled = false
  AND ce.payment_status = 'paid';

-- name: GetEventEnrolledCustomersByIDs :many
-- Get specific enrolled customers by their IDs
SELECT
    u.id,
    u.first_name,
    u.last_name,
    u.email,
    EXISTS(SELECT 1 FROM notifications.push_tokens pt WHERE pt.user_id = u.id) AS has_push_token
FROM events.customer_enrollment ce
JOIN users.users u ON ce.customer_id = u.id
WHERE ce.event_id = $1
  AND ce.is_cancelled = false
  AND ce.payment_status = 'paid'
  AND u.id = ANY($2::uuid[])
ORDER BY u.last_name, u.first_name;

-- name: CreateNotificationHistory :one
-- Record a notification in the history
INSERT INTO events.notification_history (
    event_id,
    sent_by,
    channel,
    subject,
    message,
    include_event_details,
    recipient_count,
    email_success_count,
    email_failure_count,
    push_success_count,
    push_failure_count
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: GetNotificationHistoryByEvent :many
-- Get notification history for an event
SELECT
    nh.id,
    nh.event_id,
    nh.sent_by,
    u.first_name || ' ' || u.last_name AS sent_by_name,
    nh.channel,
    nh.subject,
    nh.message,
    nh.include_event_details,
    nh.recipient_count,
    nh.email_success_count,
    nh.email_failure_count,
    nh.push_success_count,
    nh.push_failure_count,
    nh.created_at
FROM events.notification_history nh
JOIN users.users u ON nh.sent_by = u.id
WHERE nh.event_id = $1
ORDER BY nh.created_at DESC;

-- name: GetEventEnrollmentCount :one
-- Get the count of enrolled customers for an event
SELECT COUNT(*) AS count
FROM events.customer_enrollment ce
WHERE ce.event_id = $1
  AND ce.is_cancelled = false
  AND ce.payment_status = 'paid';

-- name: CheckCoachHasAccessToEvent :one
-- Check if a coach (staff) has access to an event (via team or direct assignment)
SELECT EXISTS(
    SELECT 1 FROM events.events e
    LEFT JOIN athletic.teams t ON e.team_id = t.id
    LEFT JOIN events.staff es ON es.event_id = e.id
    WHERE e.id = $1
      AND (
        t.coach_id = $2  -- Coach of the team associated with event
        OR es.staff_id = $2  -- Directly assigned to event
      )
) AS has_access;
