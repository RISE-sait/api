package membershipPlans

import (
	db "api/sqlc"
	"database/sql"

	"github.com/google/uuid"
)

type CreateMembershipPlanRequest struct {
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	Name             string    `json:"name" validate:"required_and_notwhitespace"`
	Price            int64     `json:"price" validate:"required"`
	PaymentFrequency string    `json:"payment_frequency" validate:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
}

func (r *CreateMembershipPlanRequest) ToDBParams() *db.CreateMembershipPlanParams {

	dbParams := db.CreateMembershipPlanParams{

		Name:  r.Name,
		Price: r.Price,
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(r.PaymentFrequency),
			Valid:            true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: int32(r.AmtPeriods),
			Valid: true,
		},
		MembershipID: r.MembershipID,
	}

	return &dbParams
}

type UpdateMembershipPlanRequest struct {
	Name             string    `json:"name" validate:"required_and_notwhitespace"`
	Price            int64     `json:"price" validate:"required"`
	PaymentFrequency string    `json:"payment_frequency" validate:"payment_frequency"`
	AmtPeriods       int       `json:"amt_periods"`
	MembershipID     uuid.UUID `json:"membership_id" validate:"required"`
	ID               uuid.UUID `json:"id" validate:"required"`
}

func (r *UpdateMembershipPlanRequest) ToDBParams() *db.UpdateMembershipPlanParams {

	dbParams := db.UpdateMembershipPlanParams{

		Name:  r.Name,
		Price: r.Price,
		PaymentFrequency: db.NullPaymentFrequency{
			PaymentFrequency: db.PaymentFrequency(r.PaymentFrequency),
			Valid:            true,
		},
		AmtPeriods: sql.NullInt32{
			Int32: int32(r.AmtPeriods),
			Valid: true,
		},
		MembershipID: r.MembershipID,
	}

	return &dbParams
}
