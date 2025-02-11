-- Insert staff roles
INSERT INTO staff_roles (id, role_name) VALUES
  (gen_random_uuid(), 'INSTRUCTOR'),
  (gen_random_uuid(), 'ADMIN'),
  (gen_random_uuid(), 'SUPERADMIN'),
  (gen_random_uuid(), 'COACH');

-- Insert users
INSERT INTO users (id, email) VALUES
  (gen_random_uuid(), 'alice@example.com'),
  (gen_random_uuid(), 'bob@example.com'),
  (gen_random_uuid(), 'charlie@example.com'),
  (gen_random_uuid(), 'diana@example.com'),
  (gen_random_uuid(), 'ethan@example.com'),
  (gen_random_uuid(), 'frank@example.com'),
  (gen_random_uuid(), 'grace@example.com'),
  (gen_random_uuid(), 'hannah@example.com');

-- Insert user_optional_info
INSERT INTO user_optional_info (id, first_name, last_name, hashed_password) 
VALUES
  ((SELECT id FROM users WHERE email='alice@example.com'), 'Alice', 'Johnson', 'hashed_pw1'),
  ((SELECT id FROM users WHERE email='bob@example.com'), 'Bob', 'Smith', 'hashed_pw2'),
  ((SELECT id FROM users WHERE email='charlie@example.com'), 'Charlie', 'Brown', 'hashed_pw3'),
  ((SELECT id FROM users WHERE email='diana@example.com'), 'Diana', 'White', 'hashed_pw4'),
  ((SELECT id FROM users WHERE email='ethan@example.com'), 'Ethan', 'Black', 'hashed_pw5'),
  ((SELECT id FROM users WHERE email='frank@example.com'), 'Frank', 'Green', 'hashed_pw6'),
  ((SELECT id FROM users WHERE email='grace@example.com'), 'Grace', 'Blue', 'hashed_pw7'),
  ((SELECT id FROM users WHERE email='hannah@example.com'), 'Hannah', 'Gray', 'hashed_pw8');


-- Insert staff members
INSERT INTO staff (id, is_active, created_at, updated_at, role_id) VALUES
  ((SELECT id FROM users WHERE email='alice@example.com'), true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, (SELECT id FROM staff_roles WHERE role_name='INSTRUCTOR')),
  ((SELECT id FROM users WHERE email='bob@example.com'), true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, (SELECT id FROM staff_roles WHERE role_name='ADMIN')),
  ((SELECT id FROM users WHERE email='charlie@example.com'), true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, (SELECT id FROM staff_roles WHERE role_name='SUPERADMIN')),
  ((SELECT id FROM users WHERE email='diana@example.com'), true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, (SELECT id FROM staff_roles WHERE role_name='COACH'));

-- Insert facility types
INSERT INTO facility_types (id, name) VALUES
  (gen_random_uuid(), 'Gym'),
  (gen_random_uuid(), 'Swimming Pool'),
  (gen_random_uuid(), 'Tennis Court'),
  (gen_random_uuid(), 'Basketball Court'),
  (gen_random_uuid(), 'Yoga Studio');

-- Insert facilities
INSERT INTO facilities (id, name, location, facility_type_id) VALUES
  (gen_random_uuid(), 'Downtown Gym', '123 Main St', (SELECT id FROM facility_types WHERE name='Gym')),
  (gen_random_uuid(), 'City Pool', '456 Water Ave', (SELECT id FROM facility_types WHERE name='Swimming Pool')),
  (gen_random_uuid(), 'Tennis Club', '789 Court St', (SELECT id FROM facility_types WHERE name='Tennis Court')),
  (gen_random_uuid(), 'Basketball Arena', '321 Hoop Rd', (SELECT id FROM facility_types WHERE name='Basketball Court')),
  (gen_random_uuid(), 'Serenity Yoga', '654 Zen St', (SELECT id FROM facility_types WHERE name='Yoga Studio'));

-- Insert courses
INSERT INTO courses (id, name, description, created_at, updated_at) 
VALUES
  (gen_random_uuid(), 'Beginner Yoga', 'Introduction to yoga for all ages.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Advanced Swimming', 'For experienced swimmers looking to improve.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Basketball Training', 'Intensive training for basketball players.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Tennis for Beginners', 'Learn tennis from scratch with professional coaches.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Strength Training', 'Weightlifting and strength-building exercises.', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert events
INSERT INTO events (id, begin_time, end_time, course_id, facility_id, created_at, updated_at, day) VALUES
  (gen_random_uuid(), '08:00:00', '10:00:00', (SELECT id FROM courses WHERE name='Beginner Yoga'), (SELECT id FROM facilities WHERE name='Serenity Yoga'), NOW(), NOW(), 'MONDAY'),
  (gen_random_uuid(), '09:30:00', '11:30:00', (SELECT id FROM courses WHERE name='Advanced Swimming'), (SELECT id FROM facilities WHERE name='City Pool'), NOW(), NOW(), 'TUESDAY'),
  (gen_random_uuid(), '17:00:00', '19:00:00', (SELECT id FROM courses WHERE name='Basketball Training'), (SELECT id FROM facilities WHERE name='Basketball Arena'), NOW(), NOW(), 'WEDNESDAY'),
  (gen_random_uuid(), '10:00:00', '12:00:00', (SELECT id FROM courses WHERE name='Tennis for Beginners'), (SELECT id FROM facilities WHERE name='Tennis Club'), NOW(), NOW(), 'THURSDAY'),
  (gen_random_uuid(), '12:00:00', '14:00:00', (SELECT id FROM courses WHERE name='Strength Training'), (SELECT id FROM facilities WHERE name='Downtown Gym'), NOW(), NOW(), 'FRIDAY'),
  (gen_random_uuid(), '15:00:00', '17:00:00', NULL, (SELECT id FROM facilities WHERE name='Serenity Yoga'), NOW(), NOW(), 'SATURDAY'), -- Open slot
  (gen_random_uuid(), '16:00:00', '18:00:00', NULL, (SELECT id FROM facilities WHERE name='City Pool'), NOW(), NOW(), 'SUNDAY'); -- Open slot

  -- Insert customers
INSERT INTO customers (user_id, hubspot_id, credits) VALUES
  ((SELECT id FROM users WHERE email='alice@example.com'), 123456, 10),
  ((SELECT id FROM users WHERE email='bob@example.com'), 234567, 20),
  ((SELECT id FROM users WHERE email='charlie@example.com'), 345678, 15),
  ((SELECT id FROM users WHERE email='diana@example.com'), 456789, 30),
  ((SELECT id FROM users WHERE email='ethan@example.com'), 567890, 25),
  ((SELECT id FROM users WHERE email='frank@example.com'), 678901, 5),
  ((SELECT id FROM users WHERE email='grace@example.com'), 789012, 50),
  ((SELECT id FROM users WHERE email='hannah@example.com'), 890123, 40);

-- Insert customer events
INSERT INTO customer_events (customer_id, event_id, checked_in_at) VALUES
  ((SELECT user_id FROM customers WHERE hubspot_id = 123456), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Beginner Yoga') AND facility_id = (SELECT id FROM facilities WHERE name = 'Serenity Yoga') AND day = 'MONDAY'), '2025-03-02 08:00:00+00');
  ((SELECT user_id FROM customers WHERE hubspot_id = 234567), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Advanced Swimming') AND facility_id = (SELECT id FROM facilities WHERE name = 'City Pool') AND day = 'TUESDAY'), '2025-04-12 09:30:00+00'),
  ((SELECT user_id FROM customers WHERE hubspot_id = 345678), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Basketball Training') AND facility_id = (SELECT id FROM facilities WHERE name = 'Basketball Arena') AND day = 'WEDNESDAY'), '2025-05-20 17:00:00+00'),
  ((SELECT user_id FROM customers WHERE hubspot_id = 456789), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Tennis for Beginners') AND facility_id = (SELECT id FROM facilities WHERE name = 'Tennis Club') AND day = 'THURSDAY'), '2025-06-22 10:00:00+00'),
  ((SELECT user_id FROM customers WHERE hubspot_id = 567890), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Strength Training') AND facility_id = (SELECT id FROM facilities WHERE name = 'Downtown Gym') AND day = 'FRIDAY'), '2025-07-10 12:00:00+00'),
  ((SELECT user_id FROM customers WHERE hubspot_id = 678901), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Beginner Yoga') AND facility_id = (SELECT id FROM facilities WHERE name = 'Serenity Yoga') AND day = 'SATURDAY'), NULL),
  ((SELECT user_id FROM customers WHERE hubspot_id = 789012), (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Advanced Swimming') AND facility_id = (SELECT id FROM facilities WHERE name = 'City Pool') AND day = 'SUNDAY'), NULL);

INSERT INTO users (id, email) VALUES
  (gen_random_uuid(), 'jackson@example.com'),
  (gen_random_uuid(), 'katherine@example.com'),
  (gen_random_uuid(), 'lucas@example.com'),
  (gen_random_uuid(), 'mia@example.com'),
  (gen_random_uuid(), 'noah@example.com'),
  (gen_random_uuid(), 'olivia@example.com'),
  (gen_random_uuid(), 'patrick@example.com'),
  (gen_random_uuid(), 'quinn@example.com');
-- Insert more customers
INSERT INTO customers (user_id, hubspot_id, credits) VALUES
  ((SELECT id FROM users WHERE email='jackson@example.com'), 123457, 35),
  ((SELECT id FROM users WHERE email='katherine@example.com'), 234568, 60),
  ((SELECT id FROM users WHERE email='lucas@example.com'), 345679, 10),
  ((SELECT id FROM users WHERE email='mia@example.com'), 456780, 45),
  ((SELECT id FROM users WHERE email='noah@example.com'), 567891, 50),
  ((SELECT id FROM users WHERE email='olivia@example.com'), 678902, 5),
  ((SELECT id FROM users WHERE email='patrick@example.com'), 789013, 20),
  ((SELECT id FROM users WHERE email='quinn@example.com'), 890124, 30);

-- Insert more customer events
-- Insert 1st customer event
INSERT INTO customer_events (customer_id, event_id, checked_in_at) 
VALUES 
  ((SELECT user_id FROM customers WHERE hubspot_id = 123457), 
   (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Beginner Yoga') 
    AND facility_id = (SELECT id FROM facilities WHERE name = 'Serenity Yoga') 
    AND day = 'MONDAY'), '2025-03-02 08:00:00+00');

-- Insert 2nd customer event
INSERT INTO customer_events (customer_id, event_id, checked_in_at) 
VALUES 
  ((SELECT user_id FROM customers WHERE hubspot_id = 234568), 
   (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Advanced Swimming') 
    AND facility_id = (SELECT id FROM facilities WHERE name = 'City Pool') 
    AND day = 'TUESDAY'), '2025-04-13 09:30:00+00');

-- Insert 3rd customer event
INSERT INTO customer_events (customer_id, event_id, checked_in_at) 
VALUES 
  ((SELECT user_id FROM customers WHERE hubspot_id = 345679), 
   (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Basketball Training') 
    AND facility_id = (SELECT id FROM facilities WHERE name = 'Basketball Arena') 
    AND day = 'WEDNESDAY'), '2025-05-21 17:00:00+00');

-- Insert 4th customer event
INSERT INTO customer_events (customer_id, event_id, checked_in_at) 
VALUES 
  ((SELECT user_id FROM customers WHERE hubspot_id = 456780), 
   (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Tennis for Beginners') 
    AND facility_id = (SELECT id FROM facilities WHERE name = 'Tennis Club') 
    AND day = 'THURSDAY'), '2025-06-23 10:00:00+00');

-- Insert 5th customer event
INSERT INTO customer_events (customer_id, event_id, checked_in_at) 
VALUES 
  ((SELECT user_id FROM customers WHERE hubspot_id = 567891), 
   (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Strength Training') 
    AND facility_id = (SELECT id FROM facilities WHERE name = 'Downtown Gym') 
    AND day = 'FRIDAY'), '2025-07-11 12:00:00+00');
    
-- Insert 8th customer event
INSERT INTO customer_events (customer_id, event_id, attended_at) 
VALUES 
  ((SELECT user_id FROM customers WHERE hubspot_id = 890124), 
   (SELECT id FROM events WHERE course_id = (SELECT id FROM courses WHERE name = 'Strength Training') 
    AND facility_id = (SELECT id FROM facilities WHERE name = 'Downtown Gym') 
    AND day = 'FRIDAY'), '2025-07-12 12:00:00+00');

INSERT INTO memberships (id, name, description, start_date, end_date, created_at, updated_at)
VALUES
    (gen_random_uuid(), 'Basic Plan', 'Access to gym facilities and group classes', '2024-01-01', '2024-12-31', NOW(), NOW()),
    (gen_random_uuid(), 'Premium Plan', 'Includes personal training and sauna access', '2024-01-01', '2024-12-31', NOW(), NOW()),
    (gen_random_uuid(), 'Elite Plan', 'All-inclusive membership with unlimited guest passes', '2024-01-01', '2024-12-31', NOW(), NOW());

INSERT INTO customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date, status)
SELECT c.user_id, m.id, NOW() - INTERVAL '30 days', NOW() + INTERVAL '1 year', 'active'
FROM customers c
CROSS JOIN memberships m
LIMIT 10