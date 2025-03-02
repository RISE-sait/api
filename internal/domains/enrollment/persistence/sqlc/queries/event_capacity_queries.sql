-- name: GetEventIsFull :one
SELECT
    COUNT(ce.customer_id) >= COALESCE(p.capacity, c.capacity) AS is_full
FROM events e
         LEFT JOIN customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN practices p ON e.practice_id = p.id
         LEFT JOIN course.courses c ON e.course_id = c.id
WHERE e.id = @event_id
GROUP BY e.id, e.practice_id, e.course_id, p.capacity, c.capacity;
