-- name: CreateJobPosting :one
INSERT INTO careers.job_postings (
    title, position, employment_type, location_type, description,
    responsibilities, requirements, nice_to_have,
    salary_min, salary_max, show_salary, status, closing_date, created_by
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11, $12, $13, $14
)
RETURNING *;

-- name: GetJobPostingById :one
SELECT * FROM careers.job_postings WHERE id = $1;

-- name: GetPublishedJobPostingById :one
SELECT * FROM careers.job_postings
WHERE id = $1 AND status = 'published';

-- name: ListPublishedJobPostings :many
SELECT * FROM careers.job_postings
WHERE status = 'published'
  AND (closing_date IS NULL OR closing_date > NOW())
ORDER BY published_at DESC;

-- name: ListAllJobPostings :many
SELECT * FROM careers.job_postings
ORDER BY created_at DESC;

-- name: UpdateJobPosting :one
UPDATE careers.job_postings SET
    title = $2,
    position = $3,
    employment_type = $4,
    location_type = $5,
    description = $6,
    responsibilities = $7,
    requirements = $8,
    nice_to_have = $9,
    salary_min = $10,
    salary_max = $11,
    show_salary = $12,
    closing_date = $13
WHERE id = $1
RETURNING *;

-- name: UpdateJobPostingStatus :one
UPDATE careers.job_postings SET
    status = $2,
    published_at = CASE WHEN sqlc.arg('status') = 'published' AND published_at IS NULL THEN NOW() ELSE published_at END
WHERE id = $1
RETURNING *;

-- name: DeleteJobPosting :execrows
DELETE FROM careers.job_postings WHERE id = $1;
