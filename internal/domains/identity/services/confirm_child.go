package identity

import (
	"api/cmd/server/di"
	identity "api/internal/domains/identity/persistence/repository"
	"fmt"
	"strings"

	errLib "api/internal/libs/errors"
	"api/internal/services/hubspot"
	"context"
	"database/sql"
	"net/http"
)

type ConfirmChildService struct {
	UsersRepository                                *identity.UserRepository
	CredentialsRepository                          *identity.UserCredentialsRepository
	WaiverRepository                               *identity.WaiverSigningRepository
	PendingChildrenAccountWaiversSigningRepository *identity.PendingChildAccountWaiverSigningRepository
	PendingChildAccountRepository                  *identity.PendingChildAccountRepository
	DB                                             *sql.DB
	HubspotService                                 *hubspot.HubSpotService
}

func NewConfirmChildService(
	container *di.Container,
) *ConfirmChildService {
	return &ConfirmChildService{
		UsersRepository:               identity.NewUserRepository(container),
		PendingChildAccountRepository: identity.NewPendingChildAcountRepository(container),
		DB:                            container.DB,
		HubspotService:                container.HubspotService,
		CredentialsRepository:         identity.NewUserCredentialsRepository(container),
		WaiverRepository:              identity.NewWaiverSigningRepository(container),
		PendingChildrenAccountWaiversSigningRepository: identity.NewPendingChildAccountWaiverSigningRepository(container),
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

	child, err := s.PendingChildAccountRepository.GetPendingChildAccountByChildEmail(ctx, childEmail)

	if err != nil {
		tx.Rollback()
		return err
	}

	if err := s.UsersRepository.CreateUserTx(ctx, tx, childEmail); err != nil {
		tx.Rollback()
		return err
	}

	// if password is not empty, create password
	if child.Password != "" {
		if err := s.CredentialsRepository.CreatePasswordTx(ctx, tx, childEmail, child.Password); err != nil {
			tx.Rollback()
			return err
		}
	}

	waiversSignStatuses, err := s.PendingChildrenAccountWaiversSigningRepository.GetWaiverSignings(ctx, childEmail)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, waiverSignStatus := range waiversSignStatuses {
		if !waiverSignStatus.IsSigned {
			tx.Rollback()
			return errLib.New("Waiver not signed", http.StatusConflict)
		}
		if err := s.WaiverRepository.CreateWaiverSigningRecordTx(ctx, tx, childEmail, waiverSignStatus.WaiverUrl, waiverSignStatus.IsSigned); err != nil {
			tx.Rollback()
			return err
		}
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

	// create hubspot customer

	hubspotCustomer := hubspot.HubSpotCustomerCreateBody{
		Properties: hubspot.HubSpotCustomerProps{
			FirstName: strings.Split(child.UserEmail, "@")[0],
			LastName:  strings.Split(child.UserEmail, "@")[0],
			Email:     child.UserEmail,
		},
	}

	if err := s.HubspotService.CreateCustomer(hubspotCustomer); err != nil {
		err.Message = fmt.Sprintf("Created user account in database but failed to create customer on hubspot: %s", err.Message)
		return err
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

	if err := s.HubspotService.AssociateChildAndParent(associatedChild.ID, associatedParent.ID); err != nil {
		err.Message = fmt.Sprintf("Created user account in database and hubspot but failed to associate customer with family on hubspot: %s", err.Message)
		return err
	}

	return nil
}
