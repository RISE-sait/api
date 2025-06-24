package discount

import (
	"time"

	"github.com/google/uuid"
)

type ResponseDto struct {
	ID              uuid.UUID  `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	DiscountPercent int        `json:"discount_percent"`
	IsUseUnlimited  bool       `json:"is_use_unlimited"`
	UsePerClient    int        `json:"use_per_client"`
	IsActive        bool       `json:"is_active"`
	ValidFrom       time.Time  `json:"valid_from"`
	ValidTo         *time.Time `json:"valid_to,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
