package suspension

import (
	"net/http"
	"time"

	errLib "api/internal/libs/errors"
	"github.com/google/uuid"
)

type SuspendUserRequestDto struct {
	SuspensionReason   string  `json:"suspension_reason" validate:"required,min=10,max=500"`
	SuspensionDuration *string `json:"suspension_duration,omitempty"` // e.g., "720h" (30 days), "8760h" (1 year), null = indefinite
}

type UnsuspendUserRequestDto struct {
	ExtendMembership bool `json:"extend_membership"` // whether to extend renewal_date by suspension duration
}

type SuspensionInfoResponseDto struct {
	IsSuspended         bool       `json:"is_suspended"`
	SuspendedAt         *time.Time `json:"suspended_at,omitempty"`
	SuspensionReason    *string    `json:"suspension_reason,omitempty"`
	SuspendedBy         *uuid.UUID `json:"suspended_by,omitempty"`
	SuspensionExpiresAt *time.Time `json:"suspension_expires_at,omitempty"`
}

// ParseDuration converts a string duration to time.Duration
func (dto *SuspendUserRequestDto) ParseDuration() (*time.Duration, *errLib.CommonError) {
	if dto.SuspensionDuration == nil {
		return nil, nil // indefinite suspension
	}

	duration, err := time.ParseDuration(*dto.SuspensionDuration)
	if err != nil {
		return nil, errLib.New("Invalid suspension_duration format. Use Go duration format like '720h' (30 days), '8760h' (1 year)", http.StatusBadRequest)
	}

	// Validate reasonable duration (max 10 years)
	if duration > 87600*time.Hour {
		return nil, errLib.New("Suspension duration cannot exceed 10 years", http.StatusBadRequest)
	}

	if duration <= 0 {
		return nil, errLib.New("Suspension duration must be positive", http.StatusBadRequest)
	}

	return &duration, nil
}
