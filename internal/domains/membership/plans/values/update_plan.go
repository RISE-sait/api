package values

import (
	errLib "api/internal/libs/errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type MembershipPlanUpdate struct {
	ID               uuid.UUID
	MembershipID     uuid.UUID
	Name             string
	Price            int64
	PaymentFrequency string
	AmtPeriods       int
}

func NewMembershipPlanUpdate(id, membershipID uuid.UUID, name, paymentFrequency string, price int64, amtPeriods int) *MembershipPlanUpdate {
	return &MembershipPlanUpdate{
		ID:               id,
		MembershipID:     membershipID,
		Name:             name,
		Price:            price,
		PaymentFrequency: paymentFrequency,
		AmtPeriods:       amtPeriods,
	}
}

func (mp MembershipPlanUpdate) Validate() *errLib.CommonError {
	mp.Name = strings.TrimSpace(mp.Name)

	// Validate Name
	if mp.Name == "" {
		return errLib.New(errEmptyName, http.StatusBadRequest)
	}
	if len(mp.Name) > 100 {
		return errLib.New(errNameTooLong, http.StatusBadRequest)
	}

	// Validate MembershipID
	if mp.MembershipID == uuid.Nil {
		return errLib.New(errEmptyMembershipID, http.StatusBadRequest)
	}

	// Validate Price
	if mp.Price <= 0 {
		return errLib.New(errInvalidPrice, http.StatusBadRequest)
	}

	// Validate Payment Frequency
	if mp.PaymentFrequency == "" {
		return errLib.New(errEmptyPaymentFrequency, http.StatusBadRequest)
	}

	// Validate AmtPeriods
	if mp.AmtPeriods <= 0 {
		return errLib.New(errInvalidAmtPeriods, http.StatusBadRequest)
	}

	return nil
}
