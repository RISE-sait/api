-- Insert staff roles
-- INSERT INTO users.staff_roles (id, role_name) VALUES
--   (gen_random_uuid(), 'INSTRUCTOR'),
--   (gen_random_uuid(), 'ADMIN'),
--   (gen_random_uuid(), 'SUPERADMIN'),
--   (gen_random_uuid(), 'COACH');

-- Insert users
INSERT INTO users.users (hubspot_id)
VALUES (103323445125),
       (103322588816),
       (103323445021),
       (103322588701);

-- Insert staff members
INSERT INTO users.staff (id, role_id, is_active)
VALUES ((SELECT id FROM users.users LIMIT 1 OFFSET 0),
        (SELECT id FROM users.staff_roles WHERE role_name = 'INSTRUCTOR'), true),
       ((SELECT id FROM users.users LIMIT 1 OFFSET 1), (SELECT id FROM users.staff_roles WHERE role_name = 'ADMIN'),
        true),
       ((SELECT id FROM users.users LIMIT 1 OFFSET 2),
        (SELECT id FROM users.staff_roles WHERE role_name = 'SUPERADMIN'), true),
       ((SELECT id FROM users.users LIMIT 1 OFFSET 3), (SELECT id FROM users.staff_roles WHERE role_name = 'COACH'),
        true)
ON CONFLICT (id) DO NOTHING;

-- Insert facility types
INSERT INTO location.facility_categories (name)
VALUES ('Gym'),
       ('Swimming Pool'),
       ('Tennis Court'),
       ('Basketball Court'),
       ('Yoga Studio');


-- Insert facilities
INSERT INTO location.facilities (name, address, facility_category_id)
VALUES ('Downtown Gym', '123 Main St', (SELECT id FROM location.facility_categories WHERE name = 'Gym')),
       ('City Pool', '456 Water Ave', (SELECT id FROM location.facility_categories WHERE name = 'Swimming Pool')),
       ('Tennis Club', '789 Court St', (SELECT id FROM location.facility_categories WHERE name = 'Tennis Court')),
       ('Basketball Arena', '321 Hoop Rd',
        (SELECT id FROM location.facility_categories WHERE name = 'Basketball Court')),
       ('Serenity Yoga', '654 Zen St', (SELECT id FROM location.facility_categories WHERE name = 'Yoga Studio'));

-- Insert locations
INSERT INTO location.locations (name, facility_id)
VALUES ('Downtown Gym Location', (SELECT id FROM location.facilities WHERE name = 'Downtown Gym')),
       ('City Pool Location', (SELECT id FROM location.facilities WHERE name = 'City Pool')),
       ('Tennis Club Location', (SELECT id FROM location.facilities WHERE name = 'Tennis Club')),
       ('Basketball Arena Location', (SELECT id FROM location.facilities WHERE name = 'Basketball Arena')),
       ('Serenity Yoga Location', (SELECT id FROM location.facilities WHERE name = 'Serenity Yoga'));


-- Insert courses
INSERT INTO course.courses (name, description, capacity)
VALUES ('Beginner Yoga', 'A relaxing yoga course for beginners.', 20),
       ('Advanced Swimming', 'Improve your swimming techniques with professional coaching.', 15),
       ('Tennis Fundamentals', 'Learn the basics of tennis from experienced instructors.', 12),
       ('Basketball Skills Camp', 'Enhance your basketball skills with expert guidance.', 25),
       ('Strength Training 101', 'A foundational strength training program for all levels.', 18);

-- Insert practices
INSERT INTO public.practices (name, description, level, capacity)
VALUES
    ('Yoga for Beginners', 'A gentle introduction to yoga, perfect for those new to the practice.', 'beginner', 20),
    ('Intermediate Swimming', 'Build on your swimming skills with intermediate-level techniques.', 'intermediate', 15),
    ('Advanced Tennis', 'Refine your tennis game with advanced drills and strategies.', 'advanced', 12),
    ('Basketball Training', 'Focused training on basketball shooting techniques and teamwork.', 'intermediate', 25),
    ('Strength and Conditioning', 'An intense workout program aimed at improving strength and endurance.', 'all', 18);


INSERT INTO public.games (name, video_link)
VALUES ('Football Match 2025', 'https://example.com/video/football-2025'),
       ('Basketball Championship', 'https://example.com/video/basketball-championship'),
       ('Tennis Grand Slam', 'https://example.com/video/tennis-grand-slam'),
       ('Swimming Relay 2025', 'https://example.com/video/swimming-relay'),
       ('Yoga Live Session', 'https://example.com/video/yoga-live');

-- Insert mock events
INSERT INTO public.events (event_start_at, event_end_at, session_start_time, session_end_time, practice_id, course_id, game_id, location_id, day)
VALUES
    ('2025-03-10 10:00:00+00', '2025-03-10 11:00:00+00',
     '10:00:00+00', '15:00:00+00',
     (SELECT id FROM public.practices WHERE name = 'Yoga for Beginners'),
     (SELECT id FROM course.courses WHERE name = 'Beginner Yoga'),
     (SELECT id FROM public.games WHERE name = 'Football Match 2025'),
     (SELECT id FROM location.locations WHERE name = 'Downtown Gym Location'),
     'MONDAY'),

    ('2025-03-15 09:00:00+00', '2025-03-15 10:00:00+00',
     '09:00:00+00', '12:00:00+00',
     (SELECT id FROM public.practices WHERE name = 'Intermediate Swimming'),
     (SELECT id FROM course.courses WHERE name = 'Advanced Swimming'),
     (SELECT id FROM public.games WHERE name = 'Basketball Championship'),
     (SELECT id FROM location.locations WHERE name = 'City Pool Location'),
     'SATURDAY'),

    ('2025-03-20 08:00:00+00', '2025-03-20 10:00:00+00',
     '08:00:00+00', '11:00:00+00',
     (SELECT id FROM public.practices WHERE name = 'Advanced Tennis'),
     (SELECT id FROM course.courses WHERE name = 'Tennis Fundamentals'),
     (SELECT id FROM public.games WHERE name = 'Tennis Grand Slam'),
     (SELECT id FROM location.locations WHERE name = 'Tennis Club Location'),
     'THURSDAY'),

    ('2025-03-25 13:00:00+00', '2025-03-25 14:30:00+00',
     '13:00:00+00', '16:00:00+00',
     (SELECT id FROM public.practices WHERE name = 'Basketball Training'),
     (SELECT id FROM course.courses WHERE name = 'Basketball Skills Camp'),
     (SELECT id FROM public.games WHERE name = 'Basketball Championship'),
     (SELECT id FROM location.locations WHERE name = 'Basketball Arena Location'),
     'TUESDAY'),

    ('2025-03-30 17:00:00+00', '2025-03-30 19:00:00+00',
     '17:00:00+00', '20:00:00+00',
     (SELECT id FROM public.practices WHERE name = 'Strength and Conditioning'),
     (SELECT id FROM course.courses WHERE name = 'Strength Training 101'),
     (SELECT id FROM public.games WHERE name = 'Swimming Relay 2025'),
     (SELECT id FROM location.locations WHERE name = 'Serenity Yoga Location'),
     'SUNDAY');

-- Insert memberships
INSERT INTO membership.memberships (name, description)
VALUES ('Basic Membership', 'Access to basic gym facilities'),
       ('Premium Membership', 'Access to all gym facilities and special classes'),
       ('VIP Membership', 'Exclusive access to VIP areas and events');

-- Insert membership plans
INSERT INTO membership.membership_plans (name, price, joining_fee, membership_id, payment_frequency, amt_periods)
VALUES
    -- Basic Membership Plans
    ('Basic - Football', 50, 10, (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'), 'month', 12),
    ('Basic - Basketball', 50, 10, (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'), 'month', 12),
    ('Basic - Swimming', 50, 10, (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'), 'month', 12),
    ('Basic - Yoga', 50, 10, (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'), 'month', 12),

    -- Premium Membership Plans
    ('Premium - Football', 75, 10, (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'), 'month', 12),
    ('Premium - Basketball', 75, 10, (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'), 'month', 12),
    ('Premium - Swimming', 75, 10, (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'), 'month', 12),
    ('Premium - Yoga', 75, 10, (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'), 'month', 12),

    -- VIP Membership Plans
    ('VIP - Football', 100, 10, (SELECT id FROM membership.memberships WHERE name = 'VIP Membership'), 'month', 12),
    ('VIP - Basketball', 100, 10, (SELECT id FROM membership.memberships WHERE name = 'VIP Membership'), 'month', 12),
    ('VIP - Swimming', 100, 10, (SELECT id FROM membership.memberships WHERE name = 'VIP Membership'), 'month', 12),
    ('VIP - Yoga', 100, 10, (SELECT id FROM membership.memberships WHERE name = 'VIP Membership'), 'month', 12);


-- Insert mock customer membership plans
INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date, status)
VALUES
    -- Mock data for 'Basic Football' plan
    ((SELECT id FROM users.users LIMIT 1 OFFSET 0),
     (SELECT id FROM membership.membership_plans WHERE name = 'Basic - Football'),
     '2025-03-01 00:00:00+00',
     '2026-03-01 00:00:00+00',
     'active'),

    -- Mock data for 'Premium Yoga' plan
    ((SELECT id FROM users.users LIMIT 1 OFFSET 1),
     (SELECT id FROM membership.membership_plans WHERE name = 'Premium - Yoga'),
     '2025-03-01 00:00:00+00',
     '2026-03-01 00:00:00+00',
     'active'),

    -- Mock data for 'VIP Basketball' plan
    ((SELECT id FROM users.users LIMIT 1 OFFSET 2),
     (SELECT id FROM membership.membership_plans WHERE name = 'VIP - Basketball'),
     '2025-03-01 00:00:00+00',
     '2026-03-01 00:00:00+00',
     'active'),

    -- Mock data for 'Basic Tennis' plan
    ((SELECT id FROM users.users LIMIT 1 OFFSET 3),
     (SELECT id FROM membership.membership_plans WHERE name = 'Basic - Football'),
     '2025-03-01 00:00:00+00',
     '2026-03-01 00:00:00+00',
     'active');

-- Insert mock customer enrollments
INSERT INTO public.customer_enrollment (customer_id, event_id, created_at, updated_at, checked_in_at, is_cancelled)
VALUES
    -- Mock enrollment for 'Yoga for Beginners' event
    ((SELECT id FROM users.users LIMIT 1 OFFSET 0),
     (SELECT id FROM public.events WHERE event_start_at = '2025-03-10 10:00:00+00'),
     '2025-03-01 00:00:00+00',
     '2025-03-01 00:00:00+00',
     '2025-03-10 10:30:00+00',
     false),

    -- Mock enrollment for 'Intermediate Swimming' event
    ((SELECT id FROM users.users LIMIT 1 OFFSET 1),
     (SELECT id FROM public.events WHERE event_start_at = '2025-03-15 09:00:00+00'),
     '2025-03-01 00:00:00+00',
     '2025-03-01 00:00:00+00',
     NULL,
     false),

    -- Mock enrollment for 'Advanced Tennis' event
    ((SELECT id FROM users.users LIMIT 1 OFFSET 2),
     (SELECT id FROM public.events WHERE event_start_at = '2025-03-20 08:00:00+00'),
     '2025-03-01 00:00:00+00',
     '2025-03-01 00:00:00+00',
     NULL,
     true),

    -- Mock enrollment for 'Basketball Training' event
    ((SELECT id FROM users.users LIMIT 1 OFFSET 3),
     (SELECT id FROM public.events WHERE event_start_at = '2025-03-25 13:00:00+00'),
     '2025-03-01 00:00:00+00',
     '2025-03-01 00:00:00+00',
     '2025-03-25 13:30:00+00',
     false);


-- Insert mock event staff assignments
INSERT INTO public.event_staff (event_id, staff_id)
VALUES
    -- Assign 'INSTRUCTOR' to 'Yoga for Beginners' event
    ((SELECT id FROM public.events WHERE event_start_at = '2025-03-10 10:00:00+00'),
     (SELECT id FROM users.staff WHERE role_id = (SELECT id FROM users.staff_roles WHERE role_name = 'INSTRUCTOR') LIMIT 1)),

    -- Assign 'ADMIN' to 'Intermediate Swimming' event
    ((SELECT id FROM public.events WHERE event_start_at = '2025-03-15 09:00:00+00'),
     (SELECT id FROM users.staff WHERE role_id = (SELECT id FROM users.staff_roles WHERE role_name = 'ADMIN') LIMIT 1)),

    -- Assign 'SUPERADMIN' to 'Advanced Tennis' event
    ((SELECT id FROM public.events WHERE event_start_at = '2025-03-20 08:00:00+00'),
     (SELECT id FROM users.staff WHERE role_id = (SELECT id FROM users.staff_roles WHERE role_name = 'SUPERADMIN') LIMIT 1)),

    -- Assign 'COACH' to 'Basketball Training' event
    ((SELECT id FROM public.events WHERE event_start_at = '2025-03-25 13:00:00+00'),
     (SELECT id FROM users.staff WHERE role_id = (SELECT id FROM users.staff_roles WHERE role_name = 'COACH') LIMIT 1)),

    -- Assign 'INSTRUCTOR' to 'Strength and Conditioning' event
    ((SELECT id FROM public.events WHERE event_start_at = '2025-03-30 17:00:00+00'),
     (SELECT id FROM users.staff WHERE role_id = (SELECT id FROM users.staff_roles WHERE role_name = 'INSTRUCTOR') LIMIT 1));


-- Insert mock course memberships
INSERT INTO public.course_membership (course_id, membership_id, price_per_booking, is_eligible)
VALUES
    -- Mock data for 'Basic Membership' for 'Beginner Yoga'
    ((SELECT id FROM course.courses WHERE name = 'Beginner Yoga'),
     (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'),
     15.00, true),

    -- Mock data for 'Premium Membership' for 'Advanced Swimming'
    ((SELECT id FROM course.courses WHERE name = 'Advanced Swimming'),
     (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'),
     20.00, true),

    -- Mock data for 'VIP Membership' for 'Tennis Fundamentals'
    ((SELECT id FROM course.courses WHERE name = 'Tennis Fundamentals'),
     (SELECT id FROM membership.memberships WHERE name = 'VIP Membership'),
     30.00, true),

    -- Mock data for 'Basic Membership' for 'Strength Training 101'
    ((SELECT id FROM course.courses WHERE name = 'Strength Training 101'),
     (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'),
     12.00, false),

    -- Mock data for 'Premium Membership' for 'Basketball Skills Camp'
    ((SELECT id FROM course.courses WHERE name = 'Basketball Skills Camp'),
     (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'),
     18.00, true);


-- Insert mock practice memberships
-- Insert mock practice memberships
INSERT INTO public.practice_membership (practice_id, membership_id, price_per_booking, is_eligible)
VALUES
    -- Mock data for 'Basic Membership' for 'Yoga for Beginners' practice
    ((SELECT id FROM public.practices WHERE name = 'Yoga for Beginners'),
     (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'),
     10.00, true),

    -- Mock data for 'Premium Membership' for 'Intermediate Swimming' practice
    ((SELECT id FROM public.practices WHERE name = 'Intermediate Swimming'),
     (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'),
     12.00, true),

    -- Mock data for 'VIP Membership' for 'Advanced Tennis' practice
    ((SELECT id FROM public.practices WHERE name = 'Advanced Tennis'),
     (SELECT id FROM membership.memberships WHERE name = 'VIP Membership'),
     15.00, true),

    -- Mock data for 'Basic Membership' for 'Basketball Training' practice
    ((SELECT id FROM public.practices WHERE name = 'Basketball Training'),
     (SELECT id FROM membership.memberships WHERE name = 'Basic Membership'),
     10.00, false),

    ((SELECT id FROM public.practices WHERE name = 'Strength and Conditioning'),
     (SELECT id FROM membership.memberships WHERE name = 'Premium Membership'),
     20.00, true);

-- Mock data for 'Premium Membership' f


INSERT INTO users.users (hubspot_id)
VALUES (103336460047),
       (103324806199);

-- Insert staff members
INSERT INTO users.staff (id, role_id, is_active)
VALUES
    ((SELECT id FROM users.users WHERE hubspot_id = '103336460047'::text),
     (SELECT id FROM users.staff_roles WHERE role_name = 'BARBER'), true),

    ((SELECT id FROM users.users WHERE hubspot_id = '103324806199'::text),
     (SELECT id FROM users.staff_roles WHERE role_name = 'BARBER'), true);


-- Insert mock data into barber.barber_events
INSERT INTO barber.barber_events (begin_date_time, end_date_time, customer_id, barber_id, created_at, updated_at)
VALUES
    ('2025-03-02 10:00:00+00', '2025-03-02 11:00:00+00',
     (SELECT id FROM users.users WHERE hubspot_id = '103323445125'::text),
     (SELECT id FROM users.users WHERE hubspot_id = '103336460047'::text),
     now(), now()),
    ('2025-03-02 11:30:00+00', '2025-03-02 12:30:00+00',
     (SELECT id FROM users.users WHERE hubspot_id = '103323445125'::text),
     (SELECT id FROM users.users WHERE hubspot_id = '103336460047'::text),
     now(), now());
