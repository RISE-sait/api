package child

//
//import (
//	"api/internal/di"
//	"api/internal/domains/identity/persistence/repository"
//	errLib "api/internal/libs/errors"
//	"api/internal/services/hubspot"
//	"context"
//	"database/sql"
//	"net/http"
//)
//
//type ApprovalService struct {
//	ApproveChildRepo *identity_repository.ChildApprovalRepo
//	DB               *sql.DB
//	HubspotService   *hubspot.Repository
//}
//
//func NewChildApprovalService(
//	container *di.Container,
//) *ApprovalService {
//	return &ApprovalService{
//		ApproveChildRepo: identity_repository.NewChildApprovalRepo(container),
//		DB:               container.DB,
//		HubspotService:   container.HubspotService,
//	}
//}
//
//func (s *ApprovalService) ApproveChild(
//	ctx context.Context,
//	childEmail string,
//) *errLib.CommonError {
//
//	tx, txErr := s.DB.BeginTx(ctx, nil)
//	if txErr != nil {
//		return errLib.New("Failed to start transaction", http.StatusInternalServerError)
//	}
//
//	childInfo, err := s.ApproveChildRepo.ApproveChild(ctx, tx, childEmail)
//
//	if err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	associatedParent, err := s.HubspotService.GetUserByEmail(childInfo.ParentEmail)
//
//	if err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	associatedChild, err := s.HubspotService.GetUserByEmail(childEmail)
//
//	if err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	if err := s.HubspotService.AssociateChildAndParent(associatedParent.HubSpotId, associatedChild.HubSpotId); err != nil {
//		tx.Rollback()
//		return err
//	}
//
//	return nil
//}
//
//func (s *ApprovalService) rollback(tx *sql.Tx, err *errLib.CommonError) *errLib.CommonError {
//	tx.Rollback()
//	return err
//}
