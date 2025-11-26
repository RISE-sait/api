package credit_package

import (
	errLib "api/internal/libs/errors"
	"net/http"

	"github.com/google/uuid"
)

// CreateCreditPackageRequest represents a request to create a new credit package.
// Either stripe_price_id OR pricing details (unit_amount + currency) must be provided.
// @Description Request body for creating a credit package
type CreateCreditPackageRequest struct {
	// Name of the credit package (required, 3-150 characters)
	Name string `json:"name" validate:"required,min=3,max=150" example:"10 Credit Pack"`
	// Description of the package
	Description string `json:"description" example:"Get 10 credits for facility access"`
	// Existing Stripe price ID (optional if pricing details provided)
	StripePriceID string `json:"stripe_price_id" example:"price_1234567890"`
	// Number of credits included in this package (required, min 1)
	CreditAllocation int32 `json:"credit_allocation" validate:"required,min=1" example:"10"`
	// Weekly limit on credit usage (0 = unlimited)
	WeeklyCreditLimit int32 `json:"weekly_credit_limit" validate:"min=0" example:"5"`
	// Price in cents (required if stripe_price_id not provided)
	UnitAmount *int64 `json:"unit_amount" example:"2999"`
	// Currency code (defaults to "cad" if not provided)
	Currency string `json:"currency" example:"cad"`
}

// Validate checks that either stripe_price_id or pricing details are provided
func (r *CreateCreditPackageRequest) Validate() *errLib.CommonError {
	if r.StripePriceID == "" {
		// If no Stripe price ID, require pricing details
		if r.UnitAmount == nil || *r.UnitAmount <= 0 {
			return errLib.New("Either stripe_price_id or unit_amount must be provided", http.StatusBadRequest)
		}
	}
	return nil
}

// UpdateCreditPackageRequest represents a request to update an existing credit package.
// @Description Request body for updating a credit package
type UpdateCreditPackageRequest struct {
	// Name of the credit package (required, 3-150 characters)
	Name string `json:"name" validate:"required,min=3,max=150" example:"10 Credit Pack"`
	// Description of the package
	Description string `json:"description" example:"Get 10 credits for facility access"`
	// Stripe price ID (required for updates)
	StripePriceID string `json:"stripe_price_id" validate:"required" example:"price_1234567890"`
	// Number of credits included in this package (required, min 1)
	CreditAllocation int32 `json:"credit_allocation" validate:"required,min=1" example:"10"`
	// Weekly limit on credit usage (0 = unlimited)
	WeeklyCreditLimit int32 `json:"weekly_credit_limit" validate:"min=0" example:"5"`
}

type CreditPackageIDParam struct {
	ID uuid.UUID `json:"id" validate:"required"`
}
