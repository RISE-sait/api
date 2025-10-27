package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"api/internal/di"
	staffActivityLogs "api/internal/domains/audit/staff_activity_logs/service"
	stripeService "api/internal/domains/payment/services/stripe"
	repo "api/internal/domains/user/persistence/repository"
	db "api/internal/domains/user/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"
	txUtils "api/utils/db"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/invoiceitem"
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
	CollectArrears     bool // whether to create invoice items for missed billing periods
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

		// 6. Collect arrears if requested
		var arrearsTotal int64
		if params.CollectArrears {
			arrears, arrearsErr := s.calculateAndCreateArrears(ctx, params.UserID, suspensionInfo.SuspendedAt.Time)
			if arrearsErr != nil {
				log.Printf("Warning: Failed to collect arrears for user %s: %v", params.UserID, arrearsErr)
				// Don't fail unsuspension if arrears collection fails
			} else {
				arrearsTotal = arrears
			}
		}

		// 7. Log staff activity
		extensionNote := ""
		if params.ExtendMembership {
			extensionNote = fmt.Sprintf(" (membership extended by %s)", suspensionDuration.Round(time.Hour*24))
		}

		arrearsNote := ""
		if params.CollectArrears && arrearsTotal > 0 {
			arrearsNote = fmt.Sprintf(" (arrears: $%.2f)", float64(arrearsTotal)/100)
		}

		activityDesc := fmt.Sprintf("Unsuspended user %s%s%s", params.UserID, extensionNote, arrearsNote)

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

// calculateAndCreateArrears calculates missed billing periods and creates Stripe invoice items
// Returns the total arrears amount in cents
func (s *SuspensionService) calculateAndCreateArrears(ctx context.Context, userID uuid.UUID, suspendedAt time.Time) (int64, error) {
	// Get user's Stripe customer ID
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	if err := s.db.QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID); err != nil {
		return 0, fmt.Errorf("failed to get Stripe customer ID: %w", err)
	}

	if !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		log.Printf("User %s has no Stripe customer ID - skipping arrears collection", userID)
		return 0, nil
	}

	// Get all suspended memberships
	queries := s.customerRepo.Queries
	suspendedMemberships, err := queries.GetUserSuspendedMemberships(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get suspended memberships: %w", err)
	}

	var totalArrears int64
	now := time.Now().UTC()

	// For each suspended membership, calculate missed billing periods
	for _, membership := range suspendedMemberships {
		if !membership.SuspendedAt.Valid {
			continue
		}

		// Calculate full months of suspension
		suspensionStart := membership.SuspendedAt.Time
		suspensionDuration := now.Sub(suspensionStart)
		monthsSuspended := int(suspensionDuration.Hours() / (24 * 30)) // Approximate months

		if monthsSuspended < 1 {
			log.Printf("Membership %s suspended for less than 1 month - skipping arrears", membership.ID)
			continue
		}

		// Get the membership plan price from Stripe
		// We need to get the subscription to find the price
		subscriptions, subErr := s.stripeService.GetCustomerSubscriptions(ctx)
		if subErr != nil {
			log.Printf("Warning: Failed to get subscriptions for arrears calculation: %v", subErr)
			continue
		}

		// Find the subscription matching this membership
		var monthlyAmount int64
		var stripePriceID string
		for _, sub := range subscriptions {
			// Match based on metadata or plan ID
			if sub.Items != nil && len(sub.Items.Data) > 0 {
				// Assume first item contains the membership price
				item := sub.Items.Data[0]
				if item.Price != nil {
					monthlyAmount = item.Price.UnitAmount
					stripePriceID = item.Price.ID
					break
				}
			}
		}

		if monthlyAmount == 0 {
			log.Printf("Warning: Could not determine monthly amount for membership %s", membership.ID)
			continue
		}

		// Calculate total arrears for this membership
		arrearsAmount := monthlyAmount * int64(monthsSuspended)
		totalArrears += arrearsAmount

		// Create Stripe invoice item for arrears
		invoiceItemErr := s.createStripeInvoiceItem(
			ctx,
			stripeCustomerID.String,
			arrearsAmount,
			stripePriceID,
			fmt.Sprintf("Arrears for %d month(s) during suspension (%s to %s)",
				monthsSuspended,
				suspensionStart.Format("2006-01-02"),
				now.Format("2006-01-02"),
			),
			map[string]string{
				"user_id":           userID.String(),
				"membership_id":     membership.ID.String(),
				"suspension_start":  suspensionStart.Format(time.RFC3339),
				"suspension_end":    now.Format(time.RFC3339),
				"months_suspended":  fmt.Sprintf("%d", monthsSuspended),
				"type":              "arrears",
			},
		)

		if invoiceItemErr != nil {
			log.Printf("Warning: Failed to create invoice item for membership %s: %v", membership.ID, invoiceItemErr)
			// Continue processing other memberships
		} else {
			log.Printf("Created arrears invoice item: $%.2f for membership %s", float64(arrearsAmount)/100, membership.ID)
		}
	}

	return totalArrears, nil
}

// createStripeInvoiceItem creates a Stripe invoice item (one-time charge added to next invoice)
func (s *SuspensionService) createStripeInvoiceItem(
	ctx context.Context,
	stripeCustomerID string,
	amount int64,
	stripePriceID string,
	description string,
	metadata map[string]string,
) error {
	// Check if Stripe is initialized
	if strings.ReplaceAll(stripe.Key, " ", "") == "" {
		return fmt.Errorf("Stripe not initialized")
	}

	log.Printf("Creating invoice item: customer=%s, amount=%d, description=%s", stripeCustomerID, amount, description)

	// Create invoice item params
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(stripeCustomerID),
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String("cad"), // CAD currency for Canadian customers
		Description: stripe.String(description),
	}

	// Add metadata
	for key, value := range metadata {
		params.AddMetadata(key, value)
	}

	// Create the invoice item
	_, err := invoiceitem.New(params)
	if err != nil {
		return fmt.Errorf("failed to create Stripe invoice item: %w", err)
	}

	log.Printf("Successfully created arrears invoice item for customer %s: $%.2f", stripeCustomerID, float64(amount)/100)
	return nil
}

// CollectArrearsManually manually collects arrears for a suspended user without unsuspending them
func (s *SuspensionService) CollectArrearsManually(ctx context.Context, userID uuid.UUID, collectedBy uuid.UUID) (int64, *errLib.CommonError) {
	// Get suspension info to verify user is suspended
	suspensionInfo, err := s.customerRepo.Queries.GetSuspensionInfo(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errLib.New("User not found", http.StatusNotFound)
		}
		log.Printf("Failed to get suspension info for user %s: %v", userID, err)
		return 0, errLib.New("Failed to get suspension info", http.StatusInternalServerError)
	}

	if !suspensionInfo.SuspendedAt.Valid {
		return 0, errLib.New("User is not suspended", http.StatusBadRequest)
	}

	// Calculate and create arrears
	arrears, arrearsErr := s.calculateAndCreateArrears(ctx, userID, suspensionInfo.SuspendedAt.Time)
	if arrearsErr != nil {
		log.Printf("Failed to collect arrears for user %s: %v", userID, arrearsErr)
		return 0, errLib.New("Failed to collect arrears", http.StatusInternalServerError)
	}

	// Log staff activity
	activityDesc := fmt.Sprintf("Manually collected arrears for suspended user %s: $%.2f", userID, float64(arrears)/100)
	if logErr := s.staffActivityLogsService.InsertStaffActivity(ctx, nil, collectedBy, activityDesc); logErr != nil {
		log.Printf("Warning: Failed to log arrears collection activity: %v", logErr)
	}

	return arrears, nil
}
