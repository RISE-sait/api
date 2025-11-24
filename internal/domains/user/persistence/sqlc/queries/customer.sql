-- name: UpdateAthleteStats :execrows
UPDATE athletic.athletes
SET wins       = COALESCE(sqlc.narg('wins'), wins),
    losses     = COALESCE(sqlc.narg('losses'), losses),
    points     = COALESCE(sqlc.narg('points'), points),
    steals     = COALESCE(sqlc.narg('steals'), steals),
    assists    = COALESCE(sqlc.narg('assists'), assists),
    rebounds   = COALESCE(sqlc.narg('rebounds'), rebounds),
    updated_at = current_timestamp
WHERE id = sqlc.arg('id');

-- name: UpdateAthleteTeam :execrows
UPDATE athletic.athletes
SET team_id = $1
WHERE id = sqlc.arg('athlete_id');

-- name: UpdateAthleteProfile :execrows
UPDATE athletic.athletes
SET photo_url  = $2,
    updated_at = current_timestamp
WHERE id = $1;

-- name: GetCustomers :many
SELECT u.*,
       m.name           AS membership_name,
       mp.id            AS membership_plan_id,
       mp.name          AS membership_plan_name,
       cmp.start_date   AS membership_start_date,
       cmp.renewal_date AS membership_plan_renewal_date,
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals,
       a.photo_url
FROM users.users u
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE u.is_archived = FALSE
  AND (u.parent_id = $1 OR $1 IS NULL)
  AND (sqlc.narg('search')::varchar IS NULL
  OR u.first_name ILIKE sqlc.narg('search') || '%'
  OR u.last_name ILIKE sqlc.narg('search') || '%'
  OR u.email ILIKE sqlc.narg('search') || '%'
  OR u.phone ILIKE sqlc.narg('search') || '%'
  OR u.notes ILIKE sqlc.narg('search') || '%')
  AND NOT EXISTS (SELECT 1
                  FROM staff.staff s
                  WHERE s.id = u.id)
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetCustomer :one
SELECT u.*,
       m.name           AS membership_name,
       mp.id            AS membership_plan_id,
       mp.name          AS membership_plan_name,
       cmp.start_date   AS membership_start_date,
       cmp.renewal_date AS membership_plan_renewal_date,
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals,
       a.photo_url
FROM users.users u
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE u.is_archived = FALSE
  AND (u.id = sqlc.narg('id') OR sqlc.narg('id') IS NULL)
  AND (u.email = sqlc.narg('email') OR sqlc.narg('email') IS NULL)
  AND NOT EXISTS (SELECT 1
                  FROM staff.staff s
                  WHERE s.id = u.id);

-- name: CreateAthleteInfo :execrows
INSERT INTO athletic.athletes (id, rebounds, assists, losses, wins, points)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAthletes :many
SELECT
  u.id,
  u.first_name,
  u.last_name,
  a.points,
  a.wins,
  a.losses,
  a.assists,
  a.rebounds,
  a.steals,
  a.photo_url,
  a.team_id
FROM athletic.athletes a
JOIN users.users u ON u.id = a.id
ORDER BY a.points DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountCustomers :one
SELECT COUNT(*)
FROM users.users u
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE u.is_archived = FALSE
  AND (u.parent_id = $1 OR $1 IS NULL)
  AND (sqlc.narg('search')::varchar IS NULL
  OR u.first_name ILIKE sqlc.narg('search') || '%'
  OR u.last_name ILIKE sqlc.narg('search') || '%'
  OR u.email ILIKE sqlc.narg('search') || '%'
  OR u.phone ILIKE sqlc.narg('search') || '%'
  OR u.notes ILIKE sqlc.narg('search') || '%')
  AND NOT EXISTS (SELECT 1 FROM staff.staff s WHERE s.id = u.id);

-- name: GetActiveMembershipInfo :one
SELECT
    cmp.id,
    cmp.customer_id,
    cmp.start_date,
    cmp.renewal_date,
    cmp.status,
    cmp.created_at,
    cmp.updated_at,
    cmp.photo_url,
    mp.membership_id,
    cmp.membership_plan_id,
    m.name AS membership_name,
    mp.name AS membership_plan_name
FROM users.customer_membership_plans cmp
    JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
    JOIN membership.memberships m ON m.id = mp.membership_id
WHERE cmp.customer_id = $1
  AND cmp.status = 'active'
ORDER BY cmp.start_date DESC
LIMIT 1;

-- name: ListMembershipHistory :many
SELECT
    cmp.id,
    cmp.customer_id,
    cmp.start_date,
    cmp.renewal_date,
    cmp.status,
    cmp.created_at,
    cmp.updated_at,
    mp.membership_id,
    m.name AS membership_name,
    m.description AS membership_description,
    m.benefits AS membership_benefits,
    mp.id AS membership_plan_id,
    mp.name AS membership_plan_name,
    mp.unit_amount,
    mp.currency,
    mp.interval,
    mp.stripe_price_id
FROM users.customer_membership_plans cmp
    JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
    JOIN membership.memberships m ON m.id = mp.membership_id
WHERE cmp.customer_id = $1
ORDER BY cmp.start_date DESC;

-- name: ArchiveCustomer :execrows
UPDATE users.users
SET is_archived = TRUE,
    updated_at = current_timestamp
WHERE id = sqlc.arg('id');

-- name: UnarchiveCustomer :execrows
UPDATE users.users
SET is_archived = FALSE,
    updated_at = current_timestamp
WHERE id = sqlc.arg('id');

-- name: ListArchivedCustomers :many
SELECT u.*
FROM users.users u
WHERE u.is_archived = TRUE
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: DeleteCustomerAccount :execrows
DELETE FROM users.users WHERE id = $1;

-- name: DeleteCustomerMemberships :execrows
DELETE FROM users.customer_membership_plans WHERE customer_id = $1;

-- name: DeleteCustomerEnrollments :execrows
DELETE FROM program.customer_enrollment WHERE customer_id = $1;

-- name: DeleteCustomerEventEnrollments :execrows  
DELETE FROM events.customer_enrollment WHERE customer_id = $1;

-- Note: credit_transactions has ON DELETE CASCADE, so it will be cleaned automatically

-- name: DeleteAthleteData :execrows
DELETE FROM athletic.athletes WHERE id = $1;

-- name: UpdateCustomerNotes :execrows
UPDATE users.users
SET notes = sqlc.arg('notes'),
    updated_at = current_timestamp
WHERE id = sqlc.arg('customer_id');

-- name: SuspendUser :execrows
UPDATE users.users
SET suspended_at = sqlc.arg('suspended_at'),
    suspension_reason = sqlc.arg('suspension_reason'),
    suspended_by = sqlc.arg('suspended_by'),
    suspension_expires_at = sqlc.narg('suspension_expires_at'),
    updated_at = current_timestamp
WHERE id = sqlc.arg('user_id');

-- name: UnsuspendUser :execrows
UPDATE users.users
SET suspended_at = NULL,
    suspension_reason = NULL,
    suspended_by = NULL,
    suspension_expires_at = NULL,
    updated_at = current_timestamp
WHERE id = sqlc.arg('user_id');

-- name: SuspendUserMemberships :execrows
UPDATE users.customer_membership_plans
SET suspended_at = sqlc.arg('suspended_at'),
    suspension_billing_paused = TRUE,
    updated_at = current_timestamp
WHERE customer_id = sqlc.arg('user_id')
  AND status = 'active';

-- name: UnsuspendUserMemberships :execrows
UPDATE users.customer_membership_plans
SET suspended_at = NULL,
    suspension_billing_paused = FALSE,
    updated_at = current_timestamp
WHERE customer_id = sqlc.arg('user_id')
  AND suspended_at IS NOT NULL;

-- name: ExtendMembershipRenewalDate :execrows
UPDATE users.customer_membership_plans
SET renewal_date = sqlc.arg('new_renewal_date'),
    updated_at = current_timestamp
WHERE customer_id = sqlc.arg('user_id')
  AND suspended_at IS NOT NULL;

-- name: GetUserActiveMemberships :many
SELECT *
FROM users.customer_membership_plans
WHERE customer_id = sqlc.arg('user_id')
  AND status = 'active';

-- name: GetUserSuspendedMemberships :many
SELECT *
FROM users.customer_membership_plans
WHERE customer_id = sqlc.arg('user_id')
  AND suspended_at IS NOT NULL;

-- name: GetSuspensionInfo :one
SELECT
    u.suspended_at,
    u.suspension_reason,
    suspender.first_name || ' ' || suspender.last_name as suspended_by,
    u.suspension_expires_at
FROM users.users u
LEFT JOIN users.users suspender ON suspender.id = u.suspended_by
WHERE u.id = sqlc.arg('user_id');

-- name: GetMembershipByStripeSubscriptionID :one
SELECT *
FROM users.customer_membership_plans
WHERE square_subscription_id = sqlc.arg('subscription_id')
LIMIT 1;