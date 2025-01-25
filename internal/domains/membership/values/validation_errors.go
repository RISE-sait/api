package values

const (
	errEmptyName          = "name cannot be empty or whitespace"
	errEmptyDescription   = "description cannot be empty or whitespace"
	errNameTooLong        = "name cannot exceed 100 characters"
	errStartDateRequired  = "start date is required"
	errEndDateRequired    = "end date is required"
	errInvalidDateRange   = "end date cannot be before start date"
	errPastStartDate      = "start date cannot be in the past"
	errDescriptionTooLong = "name cannot exceed 300 characters"
)
