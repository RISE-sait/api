-- name: InsertIntoOutbox :execrows
INSERT INTO audit.outbox (sql_statement, status)
VALUES ($1, $2);