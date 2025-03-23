-- name: CreateCourse :execrows
INSERT INTO public.courses (name, description, payg_price)
VALUES ($1, $2, $3);

-- name: GetCourseById :one
SELECT *
FROM public.courses
WHERE id = $1;

-- name: GetCourses :many
SELECT *
FROM public.courses;

-- name: UpdateCourse :execrows
UPDATE public.courses
SET name        = $1,
    description = $2,
    payg_price = $3,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $4;

-- name: DeleteCourse :execrows
DELETE
FROM public.courses
WHERE id = $1;