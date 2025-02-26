package staff

import (
	values "api/internal/domains/staff/values"
	"github.com/google/uuid"
)

type Staff struct {
	ID uuid.UUID
	values.Details
}
