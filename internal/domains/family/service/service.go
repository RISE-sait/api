package family

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	"api/internal/di"
	dto "api/internal/domains/family/dto"
	repo "api/internal/domains/family/persistence"
	db "api/internal/domains/family/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"
	txUtils "api/utils/db"
	"api/utils/email"

	"github.com/google/uuid"
)

const (
	codeLength     = 6
	codeExpiration = 24 * time.Hour
)

type Service struct {
	repo *repo.Repository
	db   *sql.DB
}

func NewService(container *di.Container) *Service {
	return &Service{
		repo: repo.NewFamilyRepository(container),
		db:   container.DB,
	}
}

func (s *Service) executeInTx(ctx context.Context, fn func(repo *repo.Repository) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, func(tx *sql.Tx) *errLib.CommonError {
		return fn(s.repo.WithTx(tx))
	})
}

// generateCode generates a random 6-digit verification code
func generateCode() (string, error) {
	code := ""
	for i := 0; i < codeLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", n.Int64())
	}
	return code, nil
}

// RequestLink initiates a parent-child link request
// The caller's current parent status determines the flow:
// - If caller has no parent: they're a parent wanting to add a child
// - If caller has a parent: they're a child wanting to change/add a parent
func (s *Service) RequestLink(ctx context.Context, targetEmail string) (*dto.RequestLinkResponse, *errLib.CommonError) {
	callerID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Get caller's info
	caller, err := s.repo.GetUserById(ctx, callerID)
	if err != nil {
		return nil, err
	}

	// Get target user by email
	target, err := s.repo.GetUserByEmail(ctx, targetEmail)
	if err != nil {
		if err.HTTPCode == http.StatusNotFound {
			return nil, errLib.New("No user found with that email address", http.StatusNotFound)
		}
		return nil, err
	}

	// Can't link to yourself
	if caller.ID == target.ID {
		return nil, errLib.New("Cannot link to yourself", http.StatusBadRequest)
	}

	var childID, newParentID uuid.UUID
	var oldParentID uuid.NullUUID
	var initiatedBy string
	var requiresOldParent bool

	if caller.ParentID.Valid {
		// Caller is a child (has a parent) - they want to link to a new parent
		childID = caller.ID
		newParentID = target.ID
		initiatedBy = "athlete"

		// Check if target is also a child (has a parent)
		if target.ParentID.Valid {
			return nil, errLib.New("Cannot link to another child user", http.StatusBadRequest)
		}

		// If child already has a parent, this is a transfer
		if caller.ParentID.UUID != uuid.Nil {
			oldParentID = caller.ParentID
			requiresOldParent = true
		}
	} else {
		// Caller is a parent (no parent_id) - they want to add a child
		newParentID = caller.ID
		childID = target.ID
		initiatedBy = "parent"

		// If target already has a parent, this is a transfer
		if target.ParentID.Valid && target.ParentID.UUID != uuid.Nil {
			oldParentID = target.ParentID
			requiresOldParent = true
		}
	}

	// Check for existing pending request for this child
	existingRequest, existingErr := s.repo.GetPendingRequestByChildId(ctx, childID)
	if existingErr == nil && !repo.IsExpired(existingRequest) {
		return nil, errLib.New("A pending link request already exists for this child", http.StatusConflict)
	}

	// Generate verification codes
	verificationCode, codeErr := generateCode()
	if codeErr != nil {
		log.Printf("Error generating verification code: %v", codeErr)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	var oldParentCode sql.NullString
	if requiresOldParent {
		code, codeErr := generateCode()
		if codeErr != nil {
			log.Printf("Error generating old parent code: %v", codeErr)
			return nil, errLib.New("Internal server error", http.StatusInternalServerError)
		}
		// Ensure old parent code differs from primary code
		if code == verificationCode {
			code, codeErr = generateCode()
			if codeErr != nil {
				log.Printf("Error regenerating old parent code: %v", codeErr)
				return nil, errLib.New("Internal server error", http.StatusInternalServerError)
			}
		}
		oldParentCode = sql.NullString{String: code, Valid: true}
	}

	expiresAt := time.Now().Add(codeExpiration)

	// Create the request
	params := db.CreateLinkRequestParams{
		ChildID:          childID,
		NewParentID:      newParentID,
		OldParentID:      oldParentID,
		InitiatedBy:      initiatedBy,
		VerificationCode: verificationCode,
		OldParentCode:    oldParentCode,
		ExpiresAt:        expiresAt,
	}

	_, err = s.repo.CreateLinkRequest(ctx, params)
	if err != nil {
		return nil, err
	}

	// Send verification emails
	child, childErr := s.repo.GetUserById(ctx, childID)
	if childErr != nil {
		log.Printf("Error fetching child for email: %v", childErr)
	}
	newParent, parentErr := s.repo.GetUserById(ctx, newParentID)
	if parentErr != nil {
		log.Printf("Error fetching new parent for email: %v", parentErr)
	}

	if childErr == nil && parentErr == nil {
		if initiatedBy == "athlete" {
			// Child initiated - send code to new parent
			if newParent.Email.Valid {
				go email.SendParentLinkRequestEmail(
					newParent.Email.String,
					newParent.FirstName,
					child.FirstName+" "+child.LastName,
					verificationCode,
				)
			}
		} else {
			// Parent initiated - send code to child
			if child.Email.Valid {
				go email.SendChildLinkRequestEmail(
					child.Email.String,
					child.FirstName,
					newParent.FirstName+" "+newParent.LastName,
					verificationCode,
				)
			}
		}
	}

	// If transfer, send code to old parent
	if requiresOldParent && oldParentCode.Valid {
		oldParent, oldParentErr := s.repo.GetUserById(ctx, oldParentID.UUID)
		if oldParentErr != nil {
			log.Printf("Error fetching old parent for email: %v", oldParentErr)
		} else if oldParent.Email.Valid && childErr == nil && parentErr == nil {
			go email.SendTransferApprovalEmail(
				oldParent.Email.String,
				oldParent.FirstName,
				child.FirstName+" "+child.LastName,
				newParent.FirstName+" "+newParent.LastName,
				oldParentCode.String,
			)
		}
	}

	var message string
	if initiatedBy == "athlete" {
		message = fmt.Sprintf("Verification code sent to %s", targetEmail)
	} else {
		message = fmt.Sprintf("Verification code sent to child at %s", targetEmail)
	}
	if requiresOldParent {
		message += ". Current parent must also approve the transfer."
	}

	return &dto.RequestLinkResponse{
		Message:           message,
		RequiresOldParent: requiresOldParent,
		ExpiresAt:         expiresAt,
	}, nil
}

// ConfirmLink confirms a link request with a verification code
func (s *Service) ConfirmLink(ctx context.Context, code string) (*dto.ConfirmLinkResponse, *errLib.CommonError) {
	callerID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	// Try to find request by primary verification code
	request, err := s.repo.GetRequestByVerificationCode(ctx, code)
	isOldParentCode := false

	if err != nil {
		// Try old parent code
		request, err = s.repo.GetRequestByOldParentCode(ctx, code)
		if err != nil {
			return nil, errLib.New("Invalid or expired verification code", http.StatusNotFound)
		}
		isOldParentCode = true
	}

	// Verify the caller is the expected verifier
	if isOldParentCode {
		// Old parent is verifying
		if !request.OldParentID.Valid || request.OldParentID.UUID != callerID {
			return nil, errLib.New("This verification code is not for you", http.StatusForbidden)
		}
	} else {
		// Primary verification
		if request.InitiatedBy == "athlete" {
			// Parent should be verifying
			if request.NewParentID != callerID {
				return nil, errLib.New("This verification code is not for you", http.StatusForbidden)
			}
		} else {
			// Child should be verifying
			if request.ChildID != callerID {
				return nil, errLib.New("This verification code is not for you", http.StatusForbidden)
			}
		}
	}

	var response *dto.ConfirmLinkResponse

	txErr := s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		var updatedRequest db.UsersParentLinkRequest
		var markErr *errLib.CommonError

		if isOldParentCode {
			updatedRequest, markErr = txRepo.MarkOldParentVerified(ctx, request.ID)
		} else {
			updatedRequest, markErr = txRepo.MarkVerified(ctx, request.ID)
		}

		if markErr != nil {
			return markErr
		}

		// Check if all verifications are complete
		if repo.IsComplete(updatedRequest) {
			// Update the child's parent_id
			if updateErr := txRepo.UpdateChildParent(ctx, updatedRequest.ChildID, updatedRequest.NewParentID); updateErr != nil {
				return updateErr
			}

			// Mark request as completed
			if _, completeErr := txRepo.CompleteRequest(ctx, updatedRequest.ID); completeErr != nil {
				return completeErr
			}

			child, childErr := txRepo.GetUserById(ctx, updatedRequest.ChildID)
			if childErr != nil {
				log.Printf("Error fetching child for completion email: %v", childErr)
			}
			newParent, parentErr := txRepo.GetUserById(ctx, updatedRequest.NewParentID)
			if parentErr != nil {
				log.Printf("Error fetching new parent for completion email: %v", parentErr)
			}

			// Send completion emails
			if childErr == nil && parentErr == nil && child.Email.Valid {
				go email.SendLinkCompleteEmail(
					child.Email.String,
					child.FirstName,
					newParent.FirstName+" "+newParent.LastName,
				)
			}

			childName := ""
			if childErr == nil {
				childName = child.FirstName + " " + child.LastName
			}

			response = &dto.ConfirmLinkResponse{
				Message:   "Link complete! Parent-child relationship has been established.",
				Status:    "complete",
				ChildID:   updatedRequest.ChildID.String(),
				ChildName: childName,
			}
		} else {
			// Still waiting for more verifications
			if !updatedRequest.VerifiedAt.Valid {
				response = &dto.ConfirmLinkResponse{
					Message: "Your verification is recorded. Waiting for the other party to verify.",
					Status:  "pending_new_parent",
				}
			} else {
				response = &dto.ConfirmLinkResponse{
					Message: "Your verification is recorded. Waiting for the current parent to approve the transfer.",
					Status:  "pending_old_parent",
				}
			}
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return response, nil
}

// CancelRequest cancels a pending link request
func (s *Service) CancelRequest(ctx context.Context) *errLib.CommonError {
	callerID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return err
	}

	// Find pending requests for this user
	requests, err := s.repo.GetPendingRequestsForUser(ctx, callerID)
	if err != nil {
		return err
	}

	if len(requests) == 0 {
		return errLib.New("No pending link request found", http.StatusNotFound)
	}

	// Collect IDs of requests the user is allowed to cancel
	var toCancel []uuid.UUID
	for _, req := range requests {
		if req.ChildID == callerID || (req.InitiatedBy == "parent" && req.NewParentID == callerID) {
			toCancel = append(toCancel, req.ID)
		}
	}

	if len(toCancel) == 0 {
		return errLib.New("No cancellable link request found", http.StatusNotFound)
	}

	// Cancel all in a single transaction
	return s.executeInTx(ctx, func(txRepo *repo.Repository) *errLib.CommonError {
		for _, id := range toCancel {
			if cancelErr := txRepo.CancelRequest(ctx, id); cancelErr != nil {
				return cancelErr
			}
		}
		return nil
	})
}

// GetPendingRequests gets all pending requests involving the user
func (s *Service) GetPendingRequests(ctx context.Context) ([]dto.PendingRequestResponse, *errLib.CommonError) {
	callerID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	requests, err := s.repo.GetPendingRequestsForUser(ctx, callerID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.PendingRequestResponse, len(requests))
	for i, req := range requests {
		var userRole string
		var awaitingAction bool

		if req.ChildID == callerID {
			userRole = "child"
			// Child awaits action only if parent initiated and not yet verified
			awaitingAction = req.InitiatedBy == "parent" && !req.VerifiedAt.Valid
		} else if req.NewParentID == callerID {
			userRole = "new_parent"
			// New parent awaits action if child initiated and not yet verified
			awaitingAction = req.InitiatedBy == "athlete" && !req.VerifiedAt.Valid
		} else if req.OldParentID.Valid && req.OldParentID.UUID == callerID {
			userRole = "old_parent"
			// Old parent awaits action if not yet verified
			awaitingAction = !req.OldParentVerifiedAt.Valid
		}

		var oldParentID *uuid.UUID
		var oldParentName string
		if req.OldParentID.Valid {
			oldParentID = &req.OldParentID.UUID
			oldParentName = nullStringToString(req.OldParentFirstName) + " " + nullStringToString(req.OldParentLastName)
		}

		result[i] = dto.PendingRequestResponse{
			ID:                    req.ID,
			ChildID:               req.ChildID,
			ChildName:             req.ChildFirstName + " " + req.ChildLastName,
			ChildEmail:            nullStringToString(req.ChildEmail),
			NewParentID:           req.NewParentID,
			NewParentName:         req.NewParentFirstName + " " + req.NewParentLastName,
			NewParentEmail:        nullStringToString(req.NewParentEmail),
			OldParentID:           oldParentID,
			OldParentName:         oldParentName,
			OldParentEmail:        nullStringToString(req.OldParentEmail),
			InitiatedBy:           req.InitiatedBy,
			CounterpartyVerified:  req.VerifiedAt.Valid,
			OldParentVerified:     req.OldParentVerifiedAt.Valid,
			RequiresOldParent:     req.OldParentID.Valid,
			ExpiresAt:             req.ExpiresAt,
			CreatedAt:             req.CreatedAt,
			UserRole:              userRole,
			AwaitingUserAction:    awaitingAction,
		}
	}

	return result, nil
}

// GetChildren gets all children for a parent
func (s *Service) GetChildren(ctx context.Context) ([]dto.ChildResponse, *errLib.CommonError) {
	callerID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildrenByParentId(ctx, callerID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ChildResponse, len(children))
	for i, child := range children {
		result[i] = dto.ChildResponse{
			ID:        child.ID,
			FirstName: child.FirstName,
			LastName:  child.LastName,
			Email:     nullStringToString(child.Email),
			LinkedAt:  child.LinkedAt,
		}
	}

	return result, nil
}

// AdminUnlink removes a parent-child link (admin only)
func (s *Service) AdminUnlink(ctx context.Context, childID uuid.UUID) *errLib.CommonError {
	// Verify child exists and has a parent
	child, err := s.repo.GetUserById(ctx, childID)
	if err != nil {
		return err
	}

	if !child.ParentID.Valid || child.ParentID.UUID == uuid.Nil {
		return errLib.New("This user has no parent link to remove", http.StatusBadRequest)
	}

	return s.repo.RemoveChildParent(ctx, childID)
}

// IsParentOfChild checks if a user is the parent of a specific child
func (s *Service) IsParentOfChild(ctx context.Context, parentID, childID uuid.UUID) (bool, *errLib.CommonError) {
	child, err := s.repo.GetUserById(ctx, childID)
	if err != nil {
		return false, err
	}

	if !child.ParentID.Valid {
		return false, nil
	}

	return child.ParentID.UUID == parentID, nil
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// VerifyParentChildAccess verifies that the given parent ID has access to the given child ID.
// Returns the childID if access is granted, or an error if not.
// This is used by other handlers to support the child_id query parameter.
func (s *Service) VerifyParentChildAccess(ctx context.Context, parentID, childID uuid.UUID) *errLib.CommonError {
	child, err := s.repo.GetUserById(ctx, childID)
	if err != nil {
		return err
	}

	if !child.ParentID.Valid || child.ParentID.UUID != parentID {
		return errLib.New("You do not have permission to access this child's data", http.StatusForbidden)
	}

	return nil
}
