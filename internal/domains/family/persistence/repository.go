package family

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	databaseErrors "api/internal/constants"
	"api/internal/di"
	db "api/internal/domains/family/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Repository struct {
	Queries *db.Queries
	Tx      *sql.Tx
}

func (r *Repository) GetTx() *sql.Tx {
	return r.Tx
}

func (r *Repository) WithTx(tx *sql.Tx) *Repository {
	return &Repository{
		Queries: r.Queries.WithTx(tx),
		Tx:      tx,
	}
}

func NewFamilyRepository(container *di.Container) *Repository {
	return &Repository{
		Queries: container.Queries.FamilyDb,
	}
}

// CreateLinkRequest creates a new parent link request
func (r *Repository) CreateLinkRequest(ctx context.Context, params db.CreateLinkRequestParams) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.CreateLinkRequest(ctx, params)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == databaseErrors.UniqueViolation {
				switch pqErr.Constraint {
				case "idx_parent_link_verification_code", "idx_parent_link_old_parent_code":
					return db.UsersParentLinkRequest{}, errLib.New("Verification code collision, please try again", http.StatusConflict)
				default:
					return db.UsersParentLinkRequest{}, errLib.New("A pending link request already exists for this child", http.StatusConflict)
				}
			}
		}
		log.Printf("Error creating link request: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// GetPendingRequestByChildId finds an active pending request for a child
func (r *Repository) GetPendingRequestByChildId(ctx context.Context, childID uuid.UUID) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.GetPendingRequestByChildId(ctx, childID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("No pending request found", http.StatusNotFound)
		}
		log.Printf("Error getting pending request: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// GetRequestByVerificationCode finds a request by its verification code
func (r *Repository) GetRequestByVerificationCode(ctx context.Context, code string) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.GetRequestByVerificationCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("Invalid or expired verification code", http.StatusNotFound)
		}
		log.Printf("Error getting request by code: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// GetRequestByOldParentCode finds a request by the old parent's verification code
func (r *Repository) GetRequestByOldParentCode(ctx context.Context, code string) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.GetRequestByOldParentCode(ctx, sql.NullString{String: code, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("Invalid or expired verification code", http.StatusNotFound)
		}
		log.Printf("Error getting request by old parent code: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// GetRequestById gets a request by ID
func (r *Repository) GetRequestById(ctx context.Context, id uuid.UUID) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.GetRequestById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("Request not found", http.StatusNotFound)
		}
		log.Printf("Error getting request by id: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// MarkVerified marks the primary verification as complete
func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.MarkVerified(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("Request already verified, completed, cancelled, or expired", http.StatusConflict)
		}
		log.Printf("Error marking request verified: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// MarkOldParentVerified marks the old parent's verification as complete
func (r *Repository) MarkOldParentVerified(ctx context.Context, id uuid.UUID) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.MarkOldParentVerified(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("Request already verified, completed, cancelled, or expired", http.StatusConflict)
		}
		log.Printf("Error marking old parent verified: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// CompleteRequest marks the request as completed
func (r *Repository) CompleteRequest(ctx context.Context, id uuid.UUID) (db.UsersParentLinkRequest, *errLib.CommonError) {
	request, err := r.Queries.CompleteRequest(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UsersParentLinkRequest{}, errLib.New("Request already completed or cancelled", http.StatusConflict)
		}
		log.Printf("Error completing request: %v", err)
		return db.UsersParentLinkRequest{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return request, nil
}

// CancelRequest cancels a pending request
func (r *Repository) CancelRequest(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.CancelRequest(ctx, id)
	if err != nil {
		log.Printf("Error cancelling request: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return nil
}

// UpdateChildParent updates the child's parent_id
func (r *Repository) UpdateChildParent(ctx context.Context, childID, parentID uuid.UUID) *errLib.CommonError {
	err := r.Queries.UpdateChildParent(ctx, db.UpdateChildParentParams{
		ID:       childID,
		ParentID: uuid.NullUUID{UUID: parentID, Valid: true},
	})
	if err != nil {
		log.Printf("Error updating child parent: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return nil
}

// RemoveChildParent removes the parent link from a child
func (r *Repository) RemoveChildParent(ctx context.Context, childID uuid.UUID) *errLib.CommonError {
	err := r.Queries.RemoveChildParent(ctx, childID)
	if err != nil {
		log.Printf("Error removing child parent: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return nil
}

// GetChildrenByParentId gets all children for a parent
func (r *Repository) GetChildrenByParentId(ctx context.Context, parentID uuid.UUID) ([]db.GetChildrenByParentIdRow, *errLib.CommonError) {
	children, err := r.Queries.GetChildrenByParentId(ctx, parentID)
	if err != nil {
		log.Printf("Error getting children: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return children, nil
}

// GetUserByEmail finds a user by email
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (db.GetUserByEmailRow, *errLib.CommonError) {
	user, err := r.Queries.GetUserByEmail(ctx, sql.NullString{String: email, Valid: true})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetUserByEmailRow{}, errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Error getting user by email: %v", err)
		return db.GetUserByEmailRow{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return user, nil
}

// GetUserById finds a user by ID
func (r *Repository) GetUserById(ctx context.Context, id uuid.UUID) (db.GetUserByIdRow, *errLib.CommonError) {
	user, err := r.Queries.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.GetUserByIdRow{}, errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Error getting user by id: %v", err)
		return db.GetUserByIdRow{}, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return user, nil
}

// GetPendingRequestsForUser gets all pending requests involving a user
func (r *Repository) GetPendingRequestsForUser(ctx context.Context, userID uuid.UUID) ([]db.GetPendingRequestsForUserRow, *errLib.CommonError) {
	requests, err := r.Queries.GetPendingRequestsForUser(ctx, userID)
	if err != nil {
		log.Printf("Error getting pending requests for user: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}
	return requests, nil
}

// IsExpired checks if a request has expired
func IsExpired(request db.UsersParentLinkRequest) bool {
	return request.ExpiresAt.Before(time.Now())
}

// IsComplete checks if a request is fully complete
func IsComplete(request db.UsersParentLinkRequest) bool {
	// Primary verification must be done
	if !request.VerifiedAt.Valid {
		return false
	}
	// If there's an old parent, they must also verify
	if request.OldParentID.Valid && !request.OldParentVerifiedAt.Valid {
		return false
	}
	return true
}
