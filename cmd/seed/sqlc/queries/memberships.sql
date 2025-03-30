-- name: InsertMemberships :exec
INSERT INTO membership.memberships (name, description)
VALUES (unnest(@name_array::text[]), unnest(@description_array::text[]))
RETURNING id;

-- name: InsertMembershipPlans :exec
INSERT INTO membership.membership_plans (name, price, joining_fee, auto_renew, membership_id, payment_frequency,
                                         amt_periods)
SELECT name,
       price,
       joining_fee,
       auto_renew,
       (SELECT id FROM membership.memberships m WHERE m.name = membership_name),
       payment_frequency,
       CASE WHEN ap.amt_periods = 0 THEN NULL ELSE ap.amt_periods END
FROM unnest(@name_array::text[]) WITH ORDINALITY AS n(name, ord)
         JOIN
     unnest(@price_array::numeric[]) WITH ORDINALITY AS p(price, ord) ON n.ord = p.ord
         JOIN
     unnest(@joining_fee_array::numeric[]) WITH ORDINALITY AS j(joining_fee, ord) ON n.ord = j.ord
         JOIN
     unnest(@auto_renew_array::boolean[]) WITH ORDINALITY AS a(auto_renew, ord) ON n.ord = a.ord
         JOIN
     unnest(@membership_name_array::text[]) WITH ORDINALITY AS m(membership_name, ord) ON n.ord = m.ord
         JOIN
     unnest(@payment_frequency_array::payment_frequency[]) WITH ORDINALITY AS f(payment_frequency, ord) ON n.ord = f.ord
         JOIN
     unnest(@amt_periods_array::int[]) WITH ORDINALITY AS ap(amt_periods, ord) ON n.ord = ap.ord
RETURNING id;

-- name: InsertClientsMembershipPlans :exec
WITH prepared_data as (SELECT unnest(@customer_email_array::text[])       as customer_email,
                              unnest(@membership_plan_name::text[])     as membership_plan_name,
                              unnest(@start_date_array::timestamptz[])          as start_date,
                              unnest(@renewal_date_array::timestamptz[]) as renewal_date)
INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
SELECT u.id,
mp.id,
p.start_date,
       NULLIF(p.renewal_date, '1970-01-01 00:00:00+00'::timestamptz)
FROM prepared_data p
                 JOIN users.users u ON u.email = p.customer_email
JOIN membership.membership_plans mp ON mp.name = membership_plan_name;