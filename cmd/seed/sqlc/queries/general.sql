-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest(@name_array::text[]), unnest(@address_array::text[]))
RETURNING id;

-- name: InsertPractices :exec
INSERT INTO practices (name, description, level)
VALUES (unnest(@name_array::text[]),
        unnest(@description_array::text[]),
        unnest(@level_array::practice_level[]))
RETURNING id;

-- name: InsertCourses :exec
INSERT INTO courses (name, description)
VALUES (unnest(@name_array::text[]),
        unnest(@description_array::text[]))
RETURNING id;

-- name: InsertGames :exec
INSERT INTO games (name)
VALUES (unnest(@name_array::text[]))
RETURNING id;

-- name: InsertWaivers :exec
INSERT INTO waiver.waiver(waiver_url, waiver_name)
VALUES ('https://www.youtube.com/', 'youtube'),
       ('https://www.youtube.com/watch?v=5GTFt8JNwHU', 'video');