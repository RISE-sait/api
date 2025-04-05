-- name: InsertMemberships :many
INSERT INTO membership.memberships (name, description, benefits)
VALUES (unnest(@name_array::text[]), unnest(@description_array::text[]), unnest(@benefits_array::text[]))
RETURNING id;

-- name: InsertMembershipPlans :exec
INSERT INTO membership.membership_plans (name, stripe_joining_fee_id, stripe_price_id, membership_id, amt_periods)
SELECT name,
       stripe_joining_fee_id,
       stripe_price_id,
       (SELECT id FROM membership.memberships m WHERE m.name = membership_name),
       CASE WHEN ap.amt_periods = 0 THEN NULL ELSE ap.amt_periods END
FROM unnest(@name_array::text[]) WITH ORDINALITY AS n(name, ord)
         JOIN
     unnest(@stripe_joining_fee_id_array::varchar[]) WITH ORDINALITY AS f(stripe_joining_fee_id, ord) ON n.ord = f.ord
         JOIN
     unnest(@stripe_price_id_array::varchar[]) WITH ORDINALITY AS p(stripe_price_id, ord) ON n.ord = p.ord
         JOIN
     unnest(@membership_name_array::text[]) WITH ORDINALITY AS m(membership_name, ord) ON n.ord = m.ord
         JOIN
     unnest(@amt_periods_array::int[]) WITH ORDINALITY AS ap(amt_periods, ord) ON n.ord = ap.ord
RETURNING id;

-- name: InsertClientsMembershipPlans :exec
WITH prepared_data as (SELECT unnest(@customer_email_array::text[])    as customer_email,
                              unnest(@membership_plan_name::text[])    as membership_plan_name,
                              unnest(@start_date_array::timestamptz[]) as start_date,
                              unnest(@renewal_date_array::timestamptz[]) as renewal_date)
INSERT
INTO users.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
SELECT u.id,
       mp.id,
       p.start_date,
       NULLIF(p.renewal_date, '1970-01-01 00:00:00+00'::timestamptz)
FROM prepared_data p
         JOIN users.users u ON u.email = p.customer_email
         JOIN membership.membership_plans mp ON mp.name = membership_plan_name;