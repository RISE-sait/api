package persistence

import (
	"api/internal/di"
	db "api/internal/domains/customer/persistence/sqlc/generated"
	"api/internal/domains/customer/values"
	errLib "api/internal/libs/errors"
	"time"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type CustomersRepository struct {
	Queries *db.Queries
}

func NewCustomersRepository(container *di.Container) *CustomersRepository {
	return &CustomersRepository{
		Queries: container.Queries.CustomerDb,
	}
}

func (r *CustomersRepository) GetCustomers(ctx context.Context, eventIdPtr *uuid.UUID) ([]values.CustomerWithDetails, *errLib.CommonError) {

	var eventId uuid.UUID

	if eventIdPtr != nil {
		eventId = *eventIdPtr
	}

	dbCustomersByEvent, err := r.Queries.GetCustomers(ctx, uuid.NullUUID{
		UUID:  eventId,
		Valid: eventId != uuid.Nil,
	})

	if err != nil {
		log.Println("Failed to get customers: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	customers := make([]values.CustomerWithDetails, len(dbCustomersByEvent))

	for i, dbCustomer := range dbCustomersByEvent {

		var firstName *string
		var lastName *string

		var checkedInAt *time.Time

		if dbCustomer.FirstName.Valid {
			firstName = &dbCustomer.FirstName.String
			lastName = &dbCustomer.LastName.String
		}

		if dbCustomer.CheckedInAt.Valid {
			checkedInAt = &dbCustomer.CheckedInAt.Time
		}

		customerDetails := values.CustomerWithDetails{
			CustomerInfo: values.Customer{
				CustomerID: dbCustomer.CustomerID,
				FirstName:  firstName,
				LastName:   lastName,
				Email:      dbCustomer.Email,
				Phone:      dbCustomer.Phone.String,
			},
			// MembershipInfo: values.MembershipInfo{
			// 	Name: dbCustomer.MembershipName,
			// 	PlanInfo: values.MembershipPlanInfo{
			// 		Id:               dbCustomer.MembershipPlanID,
			// 		StartDate:        dbCustomer.MembershipPlanStartDate.Time.String(),
			// 		PlanRenewalDate:  dbCustomer.MembershipPlanRenewalDate.Time.String(),
			// 		Status:           string(dbCustomer.MembershipPlanStatus),
			// 		Name:             dbCustomer.MembershipPlanName,
			// 		UpdatedAt:        dbCustomer.MembershipPlanUpdatedAt.Time.String(),
			// 		PaymentFrequency: string(dbCustomer.MembershipPlanPaymentFrequency.PaymentFrequency),
			// 		AmtPeriods:       dbCustomer.MembershipPlanAmtPeriods.Int32,
			// 		Price:            dbCustomer.MembershipPlanPrice,
			// 	}},
			EventDetails: values.CustomerEventDetails{
				CheckedInAt: checkedInAt,
				IsCancelled: dbCustomer.IsEventBookingCancelled.Bool,
			},
		}

		customers[i] = customerDetails

	}

	return customers, nil
}

// func (r *CustomersRepository) GetMembershipPlansByCustomer(ctx context.Context, id uuid.UUID) ([]values.MembershipInfo, *errLib.CommonError) {

// 	dbMemberships, err := r.Queries.GetMembershipInfoByCustomer(ctx, id)

// 	if err != nil {
// 		log.Println("Failed to get membership infos: ", err.Error())
// 		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
// 	}

// 	membershipInfos := make([]values.MembershipInfo, len(dbMemberships))
// 	for i, dbMembership := range dbMemberships {

// 		membershipInfos[i] = values.MembershipInfo{
// 			Name: dbMembership.MembershipName,
// 			PlanInfo: values.MembershipPlanInfo{
// 				Id:               dbMembership.MembershipPlanID,
// 				UpdatedAt:        dbMembership.MembershipPlanUpdatedAt.Time.GoString(),
// 				AmtPeriods:       dbMembership.AmtPeriods.Int32,
// 				PaymentFrequency: string(dbMembership.PaymentFrequency.PaymentFrequency),
// 				Price:            dbMembership.Price,
// 				Name:             dbMembership.MembershipPlanName,
// 				PlanRenewalDate:  dbMembership.MembershipPlanRenewalDate.Time.GoString(),
// 				Status:           string(dbMembership.MembershipPlanStatus),
// 				StartDate:        dbMembership.MembershipPlanStartDate.Time.GoString(),
// 			},
// 		}

// 	}

// 	return membershipInfos, nil
// }
