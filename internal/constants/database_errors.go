package database_errors

const (
	UniqueViolation           = "23505" // Postgres error code for unique violation
	ForeignKeyViolation       = "23503" // Postgres error code for foreign key violation
	NotNullViolation          = "23502" // Postgres error code for not null violation
	InvalidTextRepresentation = "22P02" // Error code for invalid input syntax, including enums
)
