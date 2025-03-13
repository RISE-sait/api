package purchase

import (
	db "api/internal/domains/membership/persistence/sqlc/generated"
	"api/internal/domains/purchase/values"
	errLib "api/internal/libs/errors"
	"api/internal/libs/validators"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type MembershipPlanRequestDto struct {
	MembershipPlanID uuid.UUID `json:"membership_plan_id"`
	StartDate        string    `json:"start_date" validate:"required"`
	Status           string    `json:"status"`
}

func (dto *MembershipPlanRequestDto) validate() *errLib.CommonError {
	if err := validators.ValidateDto(dto); err != nil {
		return err
	}
	return nil
}

func (dto *MembershipPlanRequestDto) ToPurchaseRequestInfo(customerId uuid.UUID) (values.MembershipPlanPurchaseInfo, *errLib.CommonError) {

	var vo values.MembershipPlanPurchaseInfo

	if err := dto.validate(); err != nil {
		return vo, err
	}

	startDate, err := validators.ParseDateTime(dto.StartDate)

	if err != nil {
		return vo, err
	}

	if !db.MembershipStatus(dto.Status).Valid() {

		var validStatuses []string

		for i, status := range db.AllMembershipStatusValues() {
			validStatuses[i] = string(status)
		}

		return vo, errLib.New(
			fmt.Sprintf("invalid status: %s. valid statuses are: %s", dto.Status, strings.Join(validStatuses, ", ")),
			http.StatusBadRequest,
		)
	}

	return values.MembershipPlanPurchaseInfo{
		CustomerId:       customerId,
		MembershipPlanId: dto.MembershipPlanID,
		StartDate:        startDate,
		Status:           dto.Status,
	}, nil
}
