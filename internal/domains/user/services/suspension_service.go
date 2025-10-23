package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	stripeService "api/internal/domains/payment/services/stripe"
	repo "api/internal/domains/user/persistence/repository"
	db "api/internal/domains/user/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	txUtils "api/utils/db"

	"github.com/google/uuid"
)

type SuspensionService struct {
	customerRepo             *repo.CustomerRepository
	staffActivityLogsService *staffActivityLogs.Service
	stripeService            *stripeService.SubscriptionService
	db                       *sql.DB
}

func NewSuspensionService(container *di.Container) *SuspensionService {
	return &SuspensionService{
		customerRepo:             repo.NewCustomerRepository(container),
		staffActivityLogsService: staffActivityLogs.NewService(container),
		stripeService:            stripeService.NewSubscriptionService(container),
		db:                       container.DB,
	}
}

type SuspendUserParams struct {
	UserID             uuid.UUID
	SuspendedBy        uuid.UUID
	SuspensionReason   string
	SuspensionDuration *time.Duration // nil = indefinite suspension
}

type UnsuspendUserParams struct {
	UserID             uuid.UUID
	UnsuspendedBy      uuid.UUID
	ExtendMembership   bool // whether to extend renewal_date by suspension duration
}

func (s *SuspensionService) executeInTx(ctx context.Context, fn func(tx *sql.Tx) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, fn)
}

// SuspendUser suspends a user and their active memberships
func (s *SuspensionService) SuspendUser(ctx context.Context, params SuspendUserParams) *errLib.CommonError {
	return s.executeInTx(ctx, func(tx *sql.Tx) *errLib.CommonError {
		suspendedAt := time.Now().UTC()
		var suspensionExpiresAt *time.Time

		if params.SuspensionDuration != nil {
			expiresAt := suspendedAt.Add(*params.SuspensionDuration)
			suspensionExpiresAt = &expiresAt
		}

		// 1. Suspend the user
		queries := s.customerRepo.WithTx(tx).Queries
		_, err := queries.SuspendUser(ctx, db.SuspendUserParams{
			UserID:              params.UserID,
			SuspendedAt:         sql.NullTime{Time: suspendedAt, Valid: true},
			SuspensionReason:    sql.NullString{String: params.SuspensionReason, Valid: true},
			SuspendedBy:         uuid.NullUUID{UUID: params.SuspendedBy, Valid: true},
			SuspensionExpiresAt: sql.NullTime{Time: func() time.Time { if suspensionExpiresAt != nil { return *suspensionExpiresAt }; return time.Time{} }(), Valid: suspensionExpiresAt != nil},
		})
		if err != nil {
			log.Printf("Failed to suspend user %s: %v", params.UserID, err)
			return errLib.New("Failed to suspend user", http.StatusInternalServerError)
		}

		// 2. Suspend all active memberships
		_, err = queries.SuspendUserMemberships(ctx, db.SuspendUserMembershipsParams{
			UserID:      params.UserID,
			SuspendedAt: sql.NullTime{Time: suspendedAt, Valid: true},
		})
		if err != nil {
			log.Printf("Failed to suspend memberships for user %s: %v", params.UserID, err)
			return errLib.New("Failed to suspend user memberships", http.StatusInternalServerError)
		}

		// 3. Pause Stripe subscriptions
		if pauseErr := s.pauseStripeSubscriptions(ctx, params.UserID); pauseErr != nil {
			log.Printf("Warning: Failed to pause Stripe subscriptions for user %s: %v", params.UserID, pauseErr)
			// Don't fail the entire operation if Stripe fails - log and continue
		}

		// 4. Log staff activity
		durationStr := "indefinite"
		if suspensionExpiresAt != nil {
			durationStr = fmt.Sprintf("until %s", suspensionExpiresAt.Format(time.RFC3339))
		}

		activityDesc := fmt.Sprintf("Suspended user %s (%s) - Reason: %s",
			params.UserID, durationStr, params.SuspensionReason)

		if logErr := s.staffActivityLogsService.InsertStaffActivity(ctx, tx, params.SuspendedBy, activityDesc); logErr != nil {
			log.Printf("Warning: Failed to log suspension activity: %v", logErr)
		}

		return nil
	})
}

// UnsuspendUser unsuspends a user and their memberships
func (s *SuspensionService) UnsuspendUser(ctx context.Context, params UnsuspendUserParams) *errLib.CommonError {
	return s.executeInTx(ctx, func(tx *sql.Tx) *errLib.CommonError {
		queries := s.customerRepo.WithTx(tx).Queries

		// 1. Get suspension info to calculate duration
		suspensionInfo, err := queries.GetSuspensionInfo(ctx, params.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errLib.New("User not found", http.StatusNotFound)
			}
			log.Printf("Failed to get suspension info for user %s: %v", params.UserID, err)
			return errLib.New("Failed to get suspension info", http.StatusInternalServerError)
		}

		if !suspensionInfo.SuspendedAt.Valid {
			return errLib.New("User is not suspended", http.StatusBadRequest)
		}

		suspensionDuration := time.Since(suspensionInfo.SuspendedAt.Time)

		// 2. Extend membership renewal dates if requested
		if params.ExtendMembership {
			suspendedMemberships, err := queries.GetUserSuspendedMemberships(ctx, params.UserID)
			if err != nil {
				log.Printf("Failed to get suspended memberships for user %s: %v", params.UserID, err)
				return errLib.New("Failed to get user memberships", http.StatusInternalServerError)
			}

			for _, membership := range suspendedMemberships {
				if membership.RenewalDate.Valid {
					newRenewalDate := membership.RenewalDate.Time.Add(suspensionDuration)
					_, err = queries.ExtendMembershipRenewalDate(ctx, db.ExtendMembershipRenewalDateParams{
						UserID:         params.UserID,
						NewRenewalDate: sql.NullTime{Time: newRenewalDate, Valid: true},
					})
					if err != nil {
						log.Printf("Failed to extend renewal date for membership: %v", err)
						// Continue with unsuspension even if extension fails
					}
				}
			}
		}

		// 3. Unsuspend the user
		_, err = queries.UnsuspendUser(ctx, params.UserID)
		if err != nil {
			log.Printf("Failed to unsuspend user %s: %v", params.UserID, err)
			return errLib.New("Failed to unsuspend user", http.StatusInternalServerError)
		}

		// 4. Unsuspend all memberships
		_, err = queries.UnsuspendUserMemberships(ctx, params.UserID)
		if err != nil {
			log.Printf("Failed to unsuspend memberships for user %s: %v", params.UserID, err)
			return errLib.New("Failed to unsuspend user memberships", http.StatusInternalServerError)
		}

		// 5. Resume Stripe subscriptions
		if resumeErr := s.resumeStripeSubscriptions(ctx, params.UserID); resumeErr != nil {
			log.Printf("Warning: Failed to resume Stripe subscriptions for user %s: %v", params.UserID, resumeErr)
			// Don't fail the entire operation if Stripe fails - log and continue
		}

		// 6. Log staff activity
		extensionNote := ""
		if params.ExtendMembership {
			extensionNote = fmt.Sprintf(" (membership extended by %s)", suspensionDuration.Round(time.Hour*24))
		}

		activityDesc := fmt.Sprintf("Unsuspended user %s%s", params.UserID, extensionNote)

		if logErr := s.staffActivityLogsService.InsertStaffActivity(ctx, tx, params.UnsuspendedBy, activityDesc); logErr != nil {
			log.Printf("Warning: Failed to log unsuspension activity: %v", logErr)
		}

		return nil
	})
}

// pauseStripeSubscriptions pauses all active Stripe subscriptions for a user
func (s *SuspensionService) pauseStripeSubscriptions(ctx context.Context, userID uuid.UUID) error {
	subscriptions, err := s.stripeService.GetCustomerSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get customer subscriptions: %w", err)
	}

	for _, subscription := range subscriptions {
		if subscription.Status == "active" || subscription.Status == "trialing" {
			// Pause indefinitely (nil resumeAt) - will be resumed manually when user is unsuspended
			if _, err := s.stripeService.PauseSubscription(ctx, subscription.ID, nil); err != nil {
				return fmt.Errorf("failed to pause subscription %s: %w", subscription.ID, err)
			}
		}
	}

	return nil
}

// resumeStripeSubscriptions resumes all paused Stripe subscriptions for a user
func (s *SuspensionService) resumeStripeSubscriptions(ctx context.Context, userID uuid.UUID) error {
	subscriptions, err := s.stripeService.GetCustomerSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get customer subscriptions: %w", err)
	}

	for _, subscription := range subscriptions {
		if subscription.Status == "paused" {
			if _, err := s.stripeService.ResumeSubscription(ctx, subscription.ID); err != nil {
				return fmt.Errorf("failed to resume subscription %s: %w", subscription.ID, err)
			}
		}
	}

	return nil
}

// GetSuspensionInfo retrieves suspension information for a user
func (s *SuspensionService) GetSuspensionInfo(ctx context.Context, userID uuid.UUID) (*db.GetSuspensionInfoRow, *errLib.CommonError) {
	suspensionInfo, err := s.customerRepo.Queries.GetSuspensionInfo(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to get suspension info for user %s: %v", userID, err)
		return nil, errLib.New("Failed to get suspension info", http.StatusInternalServerError)
	}

	return &suspensionInfo, nil
}
