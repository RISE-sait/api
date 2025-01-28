package database_errors

const (
	UniqueViolation     = "23505" // Postgres error code for unique violation
	ForeignKeyViolation = "23503" // Postgres error code for foreign key violation
)
