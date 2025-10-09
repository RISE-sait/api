package credit_package

import (
	"time"
	"github.com/google/uuid"
)

type CreditPackageResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Description       *string   `json:"description,omitempty"`
	StripePriceID     string    `json:"stripe_price_id"`
	CreditAllocation  int32     `json:"credit_allocation"`
	WeeklyCreditLimit int32     `json:"weekly_credit_limit"`
	Price             float64   `json:"price"`              // Price in dollars (e.g., 49.99)
	Currency          string    `json:"currency"`           // e.g., "CAD"
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CustomerActiveCreditPackageResponse struct {
	CustomerID        uuid.UUID  `json:"customer_id"`
	CreditPackageID   uuid.UUID  `json:"credit_package_id"`
	PackageName       string     `json:"package_name"`
	CreditAllocation  int32      `json:"credit_allocation"`
	WeeklyCreditLimit int32      `json:"weekly_credit_limit"`
	PurchasedAt       time.Time  `json:"purchased_at"`
}
