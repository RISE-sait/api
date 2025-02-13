package identity

import (
	"api/internal/di"
	identity "api/internal/domains/identity/persistence/repository"
	"api/internal/domains/identity/values"
	"fmt"

	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"net/http"
)

type ConfirmChildService struct {
	PendingChildrenAccountWaiversSigningRepository *identity.PendingChildAccountWaiverSigningRepository
	PendingChildAccountRepository                  *identity.PendingChildAccountRepository
	DB                                             *sql.DB
	HubspotService                                 *hubspot.HubSpotService
	CustomerRegistrationService                    *CustomerRegistrationService
}

func NewConfirmChildService(
	container *di.Container,
) *ConfirmChildService {
	return &ConfirmChildService{
		PendingChildrenAccountWaiversSigningRepository: identity.NewPendingChildAccountWaiverSigningRepository(container),
		CustomerRegistrationService:                    NewCustomerRegistrationService(container),
		PendingChildAccountRepository:                  identity.NewPendingChildAcountRepository(container),
		DB:                                             container.DB,
		HubspotService:                                 container.HubspotService,
	}
}

func (s *ConfirmChildService) ConfirmChild(
	ctx context.Context,
	childEmail, parentEmail string,
) *errLib.CommonError {

	tx, txErr := s.DB.BeginTx(ctx, nil)
	if txErr != nil {
		return errLib.New("Failed to start transaction", http.StatusInternalServerError)
	}

	// create accounts

	pendingChild, err := s.PendingChildAccountRepository.GetPendingChildAccountByChildEmail(ctx, childEmail)

	if err != nil {
		tx.Rollback()
		return err
	}

	pendingAccountWaivers, err := s.PendingChildrenAccountWaiversSigningRepository.GetWaiverSignings(ctx, childEmail)

	if err != nil {
		tx.Rollback()
		return err
	}

	waivers := make([]values.CustomerWaiverSigning, len(pendingAccountWaivers))

	for i, pendingAccountWaiver := range pendingAccountWaivers {
		waivers[i] = values.CustomerWaiverSigning{
			WaiverUrl:      pendingAccountWaiver.WaiverUrl,
			IsWaiverSigned: pendingAccountWaiver.IsSigned,
		}
	}

	customerRegistrationInfo := values.CustomerRegistrationInfo{
		UserInfo: values.UserInfo{
			FirstName: pendingChild.FirstName,
			LastName:  pendingChild.LastName,
			Email:     pendingChild.UserEmail,
		},
		Waivers:  waivers,
		Password: pendingChild.Password,
	}

	_, err = s.CustomerRegistrationService.RegisterCustomer(ctx, &customerRegistrationInfo)

	if err != nil {

		tx.Rollback()
		return err
	}

	if err := s.PendingChildrenAccountWaiversSigningRepository.DeletePendingWaiverSigningRecordByChildEmailTx(ctx, tx, childEmail); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.PendingChildAccountRepository.DeleteAccount(ctx, tx, childEmail); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	associatedParent, err := s.HubspotService.GetCustomerByEmail(parentEmail)

	if err != nil {
		err.Message = fmt.Sprintf("Created user account in database but failed to create customer on hubspot: %s", err.Message)
		return err
	}

	associatedChild, err := s.HubspotService.GetCustomerByEmail(childEmail)

	if err != nil {
		err.Message = fmt.Sprintf("Created user account in database but failed to create customer on hubspot: %s", err.Message)
		return err
	}

	if err := s.HubspotService.AssociateChildAndParent(associatedParent.ID, associatedChild.ID); err != nil {
		err.Message = fmt.Sprintf("Created user account in database and hubspot but failed to associate customer with family on hubspot: %s", err.Message)
		return err
	}

	return nil
}
