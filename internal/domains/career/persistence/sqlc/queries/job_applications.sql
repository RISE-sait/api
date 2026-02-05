-- name: CreateJobApplication :one
INSERT INTO careers.job_applications (
    job_id, first_name, last_name, email, phone,
    resume_url, cover_letter, linkedin_url, portfolio_url
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9
)
RETURNING *;

-- name: GetJobApplicationById :one
SELECT * FROM careers.job_applications WHERE id = $1;

-- name: ListJobApplicationsByJobId :many
SELECT * FROM careers.job_applications
WHERE job_id = $1
ORDER BY created_at DESC;

-- name: ListAllJobApplications :many
SELECT * FROM careers.job_applications
ORDER BY created_at DESC;

-- name: UpdateJobApplicationStatus :one
UPDATE careers.job_applications SET
    status = $2,
    reviewed_by = $3
WHERE id = $1
RETURNING *;

-- name: UpdateJobApplicationNotes :one
UPDATE careers.job_applications SET
    internal_notes = $2
WHERE id = $1
RETURNING *;

-- name: UpdateJobApplicationRating :one
UPDATE careers.job_applications SET
    rating = $2
WHERE id = $1
RETURNING *;
