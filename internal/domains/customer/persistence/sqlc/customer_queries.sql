-- name: GetCustomersForEvent :many
SELECT cu.user_id as customer_id, 
       oi.name,
       u.email,
       m.name as membership_name,
       cm.renewal_date as membership_renewal_date, 
       COUNT(*) FILTER (WHERE ce.checkedinat IS NOT NULL) AS attendance
FROM customers cu
JOIN users u ON cu.user_id = u.id
JOIN user_optional_info oi ON u.id = oi.id
JOIN customer_events ce ON cu.user_id = ce.customer_id
JOIN customer_memberships cm 
    ON cu.user_id = cm.customer_id
    AND cm.renewal_date = (
        SELECT MAX(cm2.renewal_date) 
        FROM customer_memberships cm2
        WHERE cm2.customer_id = cu.user_id
    ) 
JOIN memberships m ON cm.membership_id = m.id
WHERE (ce.event_id = $1 OR $1 = '00000000-0000-0000-0000-000000000000')
GROUP BY cu.user_id, oi.name, u.email, m.name, cm.renewal_date;