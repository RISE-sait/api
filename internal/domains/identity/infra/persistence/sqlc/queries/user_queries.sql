-- name: CreateUser :execrows
INSERT INTO users (email) VALUES ($1);
