-- name: InsertLocations :many
INSERT INTO location.locations (name, address)
VALUES (unnest(@name_array::text[]), unnest(@address_array::text[]))
RETURNING id;

-- name: InsertPractices :many
INSERT INTO public.practices (name, description, level, capacity)
VALUES (unnest(@name_array::text[]),
        unnest(@description_array::text[]),
        unnest(@level_array::practice_level[]),
        unnest(@capacity_array::int[]))
RETURNING id;

-- name: InsertCourses :many
INSERT INTO course.courses (name, description, capacity)
VALUES (unnest(@name_array::text[]),
        unnest(@description_array::text[]),
        unnest(@capacity_array::int[]))
RETURNING id;

-- name: InsertGames :many
INSERT INTO public.games (name)
VALUES (unnest(@name_array::text[]))
RETURNING id;

-- -- name: InsertEvents :many
-- INSERT INTO public.events (event_start_at, event_end_at, practice_id, course_id, game_id, location_id)
-- SELECT unnest(@event_start_at_array::timestamptz[]),
--        unnest(@event_end_at_array::timestamptz[]),
--        unnest(
--                ARRAY(
--                        SELECT CASE
--                                   WHEN practice_id = '00000000-0000-0000-0000-000000000000'
--                                       THEN NULL
--                                   ELSE practice_id
--                                   END
--                        FROM unnest(@practice_id_array::uuid[]) AS practice_id
--                )
--        ),
--        unnest(
--                ARRAY(
--                        SELECT CASE
--                                   WHEN course_id = '00000000-0000-0000-0000-000000000000'
--                                       THEN NULL
--                                   ELSE course_id
--                                   END
--                        FROM unnest(@course_id_array::uuid[]) AS course_id
--                )
--        ),
--        unnest(
--                ARRAY(
--                        SELECT CASE
--                                   WHEN game_id = '00000000-0000-0000-0000-000000000000'
--                                       THEN NULL
--                                   ELSE game_id
--                                   END
--                        FROM unnest(@game_id_array::uuid[]) AS game_id
--                )
--        ),
--        unnest(@location_id_array::uuid[])
-- ON CONFLICT DO NOTHING
-- RETURNING id;

-- name: InsertMemberships :many
INSERT INTO membership.memberships (name, description)
VALUES (unnest(@name_array::text[]), unnest(@description_array::text[]))
RETURNING id;

-- name: InsertMembershipPlans :many
INSERT INTO membership.membership_plans (name, price, joining_fee, auto_renew, membership_id, payment_frequency,
                                         amt_periods)
SELECT name,
       price,
       joining_fee,
       auto_renew,
       (SELECT id FROM membership.memberships m WHERE m.name = membership_name),
       payment_frequency,
       amt_periods
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

-- name: InsertCourseMembershipsEligibility :exec
INSERT INTO public.course_membership (course_id, membership_id, is_eligible, price_per_booking)
VALUES (unnest(@course_id_array::uuid[]),
        unnest(@membership_id_array::uuid[]),
        unnest(@is_eligible_array::bool[]),
        unnest(@price_per_booking_array::numeric[]));

-- name: InsertPracticeMembershipsEligibility :exec
INSERT INTO public.practice_membership (practice_id, membership_id, is_eligible, price_per_booking)
VALUES (unnest(@practice_id_array::uuid[]),
        unnest(@membership_id_array::uuid[]),
        unnest(@is_eligible_array::bool[]),
        unnest(@price_per_booking_array::numeric[]));


-- name: InsertClients :many
WITH prepared_data AS (SELECT unnest(@country_alpha2_code_array::text[])            AS country_alpha2_code,
                              unnest(@first_name_array::text[])                     AS first_name,
                              unnest(@last_name_array::text[])                      AS last_name,
                              unnest(@age_array::int[])                             AS age,
                              unnest(
                                      ARRAY(
                                              SELECT CASE
                                                         WHEN parent_id = '00000000-0000-0000-0000-000000000000'
                                                             THEN NULL
                                                         ELSE parent_id
                                                         END
                                              FROM unnest(@parent_id_array::uuid[]) AS parent_id
                                      )
                              )                                                     AS parent_id,
                              unnest(
                                      ARRAY(
                                              SELECT CASE
                                                         WHEN gender = 'N'
                                                             THEN NULL
                                                         ELSE gender
                                                         END
                                              FROM unnest(@gender_array::char[]) AS gender
                                      )
                              ) AS gender,
                              unnest(@phone_array::text[])                          AS phone,
                              unnest(@email_array::text[])                          AS email,
                              unnest(@has_marketing_email_consent_array::boolean[]) AS has_marketing_email_consent,
                              unnest(@has_sms_consent_array::boolean[])             AS has_sms_consent)
INSERT
INTO users.users (country_alpha2_code,
                  first_name,
                  last_name,
                  age,
                  gender,
                  parent_id,
                  phone,
                  email,
                  has_marketing_email_consent,
                  has_sms_consent)
SELECT country_alpha2_code,
       first_name,
       last_name,
       age,
       gender,
       parent_id,
       phone,
       email,
       has_marketing_email_consent,
       has_sms_consent
FROM prepared_data
RETURNING id;


-- name: InsertAthletes :many
INSERT
INTO users.athletes (id)
VALUES (unnest(@id_array::uuid[]))
RETURNING id;

-- -- name: InsertClientsMembershipPlans :exec
-- WITH prepared_data AS (SELECT unnest(@customer_id_array::uuid[])  AS customer_id,
--                               unnest(
--                                       ARRAY(
--                                               SELECT CASE
--                                                          WHEN membership_plan_id = '00000000-0000-0000-0000-000000000000'
--                                                              THEN NULL
--                                                          ELSE membership_plan_id
--                                                          END
--                                               FROM unnest(@membership_plan_id_array::uuid[]) AS membership_plan_id
--                                       )
--                               ),
--                                   unnest(
--                                       ARRAY(
--                                               SELECT CASE
--                                                          WHEN start_date = '0001-01-01 00:00:00 UTC'
--                                                              THEN NULL
--                                                          ELSE start_date
--                                                          END
--                                               FROM unnest(@start_date_array::timestamptz[]) AS start_date
--                                       )
--                               )                                   AS start_date,
--                               unnest(
--                                       ARRAY(
--                                               SELECT CASE
--                                                          WHEN renewal_date = '0001-01-01 00:00:00 UTC'
--                                                              THEN NULL
--                                                          ELSE renewal_date
--                                                          END
--                                               FROM unnest(@renewal_date_array::timestamptz[]) AS renewal_date
--                                       )
--                               )                                   AS renewal_date)
-- INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
-- VALUES (  customer_id, membership_plan_id, start_date, renewal_date);

-- name: InsertClientsMembershipPlans :many
INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
VALUES (unnest(@customer_id::uuid[]),
        unnest(@plans_array::uuid[]),
        unnest(@start_date_array::timestamptz[]),
        unnest(@renewal_date_array::timestamptz[]))
RETURNING id;

-- name: InsertCustomersEnrollments :many
WITH prepared_data AS (SELECT unnest(@customer_id_array::uuid[])  AS customer_id,
                              unnest(@event_id_array::uuid[])     AS event_id,
                              unnest(
                                      ARRAY(
                                              SELECT CASE
                                                         WHEN checked_in_at = '0001-01-01 00:00:00 UTC'
                                                             THEN NULL
                                                         ELSE checked_in_at
                                                         END
                                              FROM unnest(@checked_in_at_array::timestamptz[]) AS checked_in_at
                                      )
                              )                                   AS checked_in_at,
                              unnest(@is_cancelled_array::bool[]) AS is_cancelled)
INSERT
INTO public.customer_enrollment(customer_id, event_id, checked_in_at, is_cancelled)
SELECT customer_id,
       event_id,
       checked_in_at,
       is_cancelled
FROM prepared_data
RETURNING id;

-- name: InsertStaffRoles :exec
INSERT INTO users.staff_roles (role_name)
VALUES ('admin'),
       ('superadmin'),
       ('coach'),
       ('instructor'),
       ('receptionist'),
       ('barber');

-- name: InsertStaff :exec
WITH staff_data AS (SELECT e.email,
                           ia.is_active,
                           rn.role_name
                    FROM unnest(@emails::text[]) WITH ORDINALITY AS e(email, idx)
                             JOIN
                         unnest(@is_active_array::bool[]) WITH ORDINALITY AS ia(is_active, idx)
                         ON e.idx = ia.idx
                             JOIN
                         unnest(@role_name_array::text[]) WITH ORDINALITY AS rn(role_name, idx)
                         ON e.idx = rn.idx)
INSERT
INTO users.staff (id, is_active, role_id)
SELECT u.id,
       sd.is_active,
       sr.id
FROM staff_data sd
         JOIN
     users.users u ON u.email = sd.email
         JOIN
     users.staff_roles sr ON sr.role_name = sd.role_name;