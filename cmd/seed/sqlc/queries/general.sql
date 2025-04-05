-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest(@name_array::text[]), unnest(@address_array::text[]))
RETURNING id;

-- name: InsertTeams :many
WITH prepared_data AS (SELECT unnest(@coach_email_array::text[])          AS coach,
                              unnest(@capacity_array::int[])             AS capacity,
                              unnest(@name_array::text[]) AS name)
INSERT
INTO athletic.teams(capacity, coach_id, name)
SELECT capacity, u.id, name
FROM prepared_data
JOIN users.users u ON u.email = coach
RETURNING id;


-- name: InsertWaivers :exec
INSERT INTO waiver.waiver(waiver_url, waiver_name)
VALUES ('https://storage.googleapis.com/rise-sports/waivers/code.pdf', 'code_pdf'),
       ('https://storage.googleapis.com/rise-sports/waivers/tetris.pdf', 'tetris_pdf');

-- name: InsertCoachStats :exec
INSERT INTO athletic.coach_stats (coach_id, wins, losses)
VALUES ((SELECT id FROM users.users WHERE email = 'viktor.djurasic+1@abcfitness.com'),
        1,
        1),
       ((SELECT id FROM users.users WHERE email = 'coach@test.com'),
        1,
        2);
