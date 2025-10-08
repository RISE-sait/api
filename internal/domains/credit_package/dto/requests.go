package credit_package

import "github.com/google/uuid"

type CreateCreditPackageRequest struct {
	Name              string `json:"name" validate:"required,min=3,max=150"`
	Description       string `json:"description"`
	StripePriceID     string `json:"stripe_price_id" validate:"required"`
	CreditAllocation  int32  `json:"credit_allocation" validate:"required,min=1"`
	WeeklyCreditLimit int32  `json:"weekly_credit_limit" validate:"min=0"`
}

type UpdateCreditPackageRequest struct {
	Name              string `json:"name" validate:"required,min=3,max=150"`
	Description       string `json:"description"`
	StripePriceID     string `json:"stripe_price_id" validate:"required"`
	CreditAllocation  int32  `json:"credit_allocation" validate:"required,min=1"`
	WeeklyCreditLimit int32  `json:"weekly_credit_limit" validate:"min=0"`
}

type CreditPackageIDParam struct {
	ID uuid.UUID `json:"id" validate:"required"`
}
