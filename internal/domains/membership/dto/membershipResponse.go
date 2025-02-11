package membership

import (
	"github.com/google/uuid"
)

type MembershipResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
}
