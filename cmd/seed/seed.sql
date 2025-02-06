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
INSERT INTO user_optional_info (id, name, hashed_password) VALUES
  ((SELECT id FROM users WHERE email='alice@example.com'), 'Alice Johnson', 'hashed_pw1'),
  ((SELECT id FROM users WHERE email='bob@example.com'), 'Bob Smith', 'hashed_pw2'),
  ((SELECT id FROM users WHERE email='charlie@example.com'), 'Charlie Brown', 'hashed_pw3'),
  ((SELECT id FROM users WHERE email='diana@example.com'), 'Diana White', 'hashed_pw4'),
  ((SELECT id FROM users WHERE email='ethan@example.com'), 'Ethan Black', 'hashed_pw5'),
  ((SELECT id FROM users WHERE email='frank@example.com'), 'Frank Green', 'hashed_pw6'),
  ((SELECT id FROM users WHERE email='grace@example.com'), 'Grace Blue', 'hashed_pw7'),
  ((SELECT id FROM users WHERE email='hannah@example.com'), 'Hannah Gray', 'hashed_pw8');

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
INSERT INTO courses (id, name, description, start_date, end_date, created_at, updated_at) VALUES
  (gen_random_uuid(), 'Beginner Yoga', 'Introduction to yoga for all ages.', '2025-03-01 08:00:00+00', '2025-06-01 08:00:00+00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Advanced Swimming', 'For experienced swimmers looking to improve.', '2025-04-10 09:30:00+00', '2025-08-10 09:30:00+00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Basketball Training', 'Intensive training for basketball players.', '2025-05-15 17:00:00+00', '2025-09-15 17:00:00+00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Tennis for Beginners', 'Learn tennis from scratch with professional coaches.', '2025-06-20 10:00:00+00', '2025-10-20 10:00:00+00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
  (gen_random_uuid(), 'Strength Training', 'Weightlifting and strength-building exercises.', '2025-07-05 12:00:00+00', '2025-11-05 12:00:00+00', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Insert schedules
INSERT INTO schedules (id, begin_time, end_time, course_id, facility_id, created_at, updated_at, day) VALUES
  (gen_random_uuid(), '08:00:00', '10:00:00', (SELECT id FROM courses WHERE name='Beginner Yoga'), (SELECT id FROM facilities WHERE name='Serenity Yoga'), NOW(), NOW(), 'MONDAY'),
  (gen_random_uuid(), '09:30:00', '11:30:00', (SELECT id FROM courses WHERE name='Advanced Swimming'), (SELECT id FROM facilities WHERE name='City Pool'), NOW(), NOW(), 'TUESDAY'),
  (gen_random_uuid(), '17:00:00', '19:00:00', (SELECT id FROM courses WHERE name='Basketball Training'), (SELECT id FROM facilities WHERE name='Basketball Arena'), NOW(), NOW(), 'WEDNESDAY'),
  (gen_random_uuid(), '10:00:00', '12:00:00', (SELECT id FROM courses WHERE name='Tennis for Beginners'), (SELECT id FROM facilities WHERE name='Tennis Club'), NOW(), NOW(), 'THURSDAY'),
  (gen_random_uuid(), '12:00:00', '14:00:00', (SELECT id FROM courses WHERE name='Strength Training'), (SELECT id FROM facilities WHERE name='Downtown Gym'), NOW(), NOW(), 'FRIDAY'),
  (gen_random_uuid(), '15:00:00', '17:00:00', NULL, (SELECT id FROM facilities WHERE name='Serenity Yoga'), NOW(), NOW(), 'SATURDAY'), -- Open slot
  (gen_random_uuid(), '16:00:00', '18:00:00', NULL, (SELECT id FROM facilities WHERE name='City Pool'), NOW(), NOW(), 'SUNDAY'); -- Open slot