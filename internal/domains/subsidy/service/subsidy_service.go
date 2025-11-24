package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"api/internal/di"
	"api/internal/domains/subsidy/dto"
	repo "api/internal/domains/subsidy/persistence/repository"
	db "api/internal/domains/subsidy/persistence/sqlc/generated"
	"api/internal/domains/payment/tracking"
	errLib "api/internal/libs/errors"
	txUtils "api/utils/db"
	"api/utils/email"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Business rule constants
const (
	MAX_SUBSIDY_AMOUNT        = 100000.0 // Maximum subsidy amount ($100,000)
	FLOAT_PRECISION_TOLERANCE = 0.01     // Tolerance for floating point comparisons
)

type SubsidyService struct {
	repo            *repo.SubsidyRepository
	db              *sql.DB
	paymentTracking *tracking.PaymentTrackingService
}

func NewSubsidyService(container *di.Container) *SubsidyService {
	return &SubsidyService{
		repo:            repo.NewSubsidyRepository(container),
		db:              container.DB,
		paymentTracking: tracking.NewPaymentTrackingService(container),
	}
}

func (s *SubsidyService) executeInTx(ctx context.Context, fn func(tx *sql.Tx) *errLib.CommonError) *errLib.CommonError {
	return txUtils.ExecuteInTx(ctx, s.db, fn)
}

// executeInSerializableTx executes a function within a SERIALIZABLE transaction to prevent race conditions
func (s *SubsidyService) executeInSerializableTx(ctx context.Context, fn func(tx *sql.Tx) *errLib.CommonError) *errLib.CommonError {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		log.Printf("Failed to begin serializable transaction: %v", err)
		return errLib.New("Failed to begin transaction", http.StatusInternalServerError)
	}

	// Execute the function
	if fnErr := fn(tx); fnErr != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
		}
		return fnErr
	}

	// Commit the transaction
	if commitErr := tx.Commit(); commitErr != nil {
		log.Printf("Failed to commit transaction: %v", commitErr)
		return errLib.New("Failed to commit transaction", http.StatusInternalServerError)
	}

	return nil
}

// ===== PROVIDER METHODS =====

func (s *SubsidyService) CreateProvider(ctx context.Context, req *dto.CreateProviderRequest) (*dto.ProviderResponse, *errLib.CommonError) {
	provider, err := s.repo.Queries.CreateProvider(ctx, db.CreateProviderParams{
		Name:         req.Name,
		ContactEmail: sqlString(req.ContactEmail),
		ContactPhone: sqlString(req.ContactPhone),
		IsActive:     sql.NullBool{Bool: true, Valid: true},
	})

	if err != nil {
		log.Printf("Failed to create provider: %v", err)
		return nil, errLib.New("Failed to create provider", http.StatusInternalServerError)
	}

	return mapProviderToResponse(provider), nil
}

func (s *SubsidyService) GetProvider(ctx context.Context, id uuid.UUID) (*dto.ProviderResponse, *errLib.CommonError) {
	provider, err := s.repo.Queries.GetProvider(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Provider not found", http.StatusNotFound)
		}
		log.Printf("Failed to get provider: %v", err)
		return nil, errLib.New("Failed to get provider", http.StatusInternalServerError)
	}

	return mapProviderToResponse(provider), nil
}

func (s *SubsidyService) ListProviders(ctx context.Context, isActive *bool) ([]dto.ProviderResponse, *errLib.CommonError) {
	providers, err := s.repo.Queries.ListProviders(ctx, sqlBool(isActive))
	if err != nil {
		log.Printf("Failed to list providers: %v", err)
		return nil, errLib.New("Failed to list providers", http.StatusInternalServerError)
	}

	result := make([]dto.ProviderResponse, len(providers))
	for i, p := range providers {
		result[i] = *mapProviderToResponse(p)
	}

	return result, nil
}

func (s *SubsidyService) GetProviderStats(ctx context.Context, id uuid.UUID) (*dto.ProviderStatsResponse, *errLib.CommonError) {
	stats, err := s.repo.Queries.GetProviderStats(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Provider not found", http.StatusNotFound)
		}
		log.Printf("Failed to get provider stats: %v", err)
		return nil, errLib.New("Failed to get provider stats", http.StatusInternalServerError)
	}

	return &dto.ProviderStatsResponse{
		ID:                stats.ID,
		Name:              stats.Name,
		TotalSubsidies:    stats.TotalSubsidies,
		TotalAmountIssued: interfaceToFloat(stats.TotalAmountIssued),
		TotalAmountUsed:   interfaceToFloat(stats.TotalAmountUsed),
		TotalRemaining:    interfaceToFloat(stats.TotalRemaining),
	}, nil
}

// ===== SUBSIDY CRUD METHODS =====

func (s *SubsidyService) CreateSubsidy(ctx context.Context, req *dto.CreateSubsidyRequest, staffID uuid.UUID, ipAddress string) (*dto.SubsidyResponse, *errLib.CommonError) {
	// SECURITY: Input validation
	if req.ApprovedAmount <= 0 {
		return nil, errLib.New("Approved amount must be positive", http.StatusBadRequest)
	}
	if req.ApprovedAmount > MAX_SUBSIDY_AMOUNT {
		return nil, errLib.New(fmt.Sprintf("Approved amount exceeds maximum limit ($%.0f)", MAX_SUBSIDY_AMOUNT), http.StatusBadRequest)
	}
	if req.Reason == "" {
		return nil, errLib.New("Reason is required for subsidy creation", http.StatusBadRequest)
	}

	var result *dto.SubsidyResponse
	err := s.executeInSerializableTx(ctx, func(tx *sql.Tx) *errLib.CommonError {
		queries := s.repo.Queries.WithTx(tx)

		// SECURITY: Verify provider exists and is active
		provider, provErr := queries.GetProvider(ctx, req.ProviderID)
		if provErr != nil {
			if provErr == sql.ErrNoRows {
				return errLib.New("Provider not found", http.StatusNotFound)
			}
			log.Printf("Failed to verify provider: %v", provErr)
			return errLib.New("Failed to verify provider", http.StatusInternalServerError)
		}

		if !provider.IsActive.Valid || !provider.IsActive.Bool {
			log.Printf("[SECURITY] Attempt to create subsidy with inactive provider: %s", req.ProviderID)
			return errLib.New("Provider is not active", http.StatusBadRequest)
		}

		// SECURITY: Check for existing active subsidy for this customer
		existingSubsidy, _ := queries.GetActiveSubsidyForCustomer(ctx, req.CustomerID)
		if existingSubsidy.ID != uuid.Nil {
			log.Printf("[SECURITY] Attempt to create duplicate active subsidy for customer %s (existing: %s)",
				req.CustomerID, existingSubsidy.ID)
			return errLib.New(fmt.Sprintf("Customer already has an active subsidy (ID: %s) with $%.2f remaining",
				existingSubsidy.ID, sqlStringToFloat(existingSubsidy.RemainingBalance)), http.StatusConflict)
		}

		// Create the subsidy with 'active' status (staff creates it directly)
		subsidy, dbErr := queries.CreateCustomerSubsidy(ctx, db.CreateCustomerSubsidyParams{
			CustomerID:     req.CustomerID,
			ProviderID:     uuid.NullUUID{UUID: req.ProviderID, Valid: true},
			ApprovedAmount: decimal.NewFromFloat(req.ApprovedAmount),
			Status:         "active",
			ApprovedBy:     uuid.NullUUID{UUID: staffID, Valid: true},
			ApprovedAt:     sql.NullTime{Time: time.Now(), Valid: true},
			ValidFrom:      time.Now(),
			ValidUntil:     sqlTime(req.ValidUntil),
			Reason:         sql.NullString{String: req.Reason, Valid: true},
			AdminNotes:     sql.NullString{String: req.AdminNotes, Valid: req.AdminNotes != ""},
		})

		if dbErr != nil {
			log.Printf("Failed to create customer subsidy: %v", dbErr)
			return errLib.New("Failed to create subsidy", http.StatusInternalServerError)
		}

		// Create comprehensive audit log with IP address
		_, auditErr := queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
			CustomerSubsidyID: uuid.NullUUID{UUID: subsidy.ID, Valid: true},
			Action:            "created",
			PerformedBy:       uuid.NullUUID{UUID: staffID, Valid: true},
			NewStatus:         sql.NullString{String: "active", Valid: true},
			AmountChanged:     sql.NullString{String: decimal.NewFromFloat(req.ApprovedAmount).String(), Valid: true},
			Notes:             sql.NullString{String: fmt.Sprintf("Subsidy created by staff. Provider: %s, Amount: $%.2f, Reason: %s", provider.Name, req.ApprovedAmount, req.Reason), Valid: true},
			IpAddress:         sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		})

		if auditErr != nil {
			log.Printf("Warning: Failed to create audit log: %v", auditErr)
		}

		// Get full details for response
		fullSubsidy, getErr := queries.GetCustomerSubsidy(ctx, subsidy.ID)
		if getErr != nil {
			return errLib.New("Failed to retrieve created subsidy", http.StatusInternalServerError)
		}

		result = mapSubsidyToResponse(fullSubsidy)
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("✅ [AUDIT] Created subsidy: ID=%s, Amount=$%.2f, Customer=%s, Provider=%s, Staff=%s",
		result.ID, req.ApprovedAmount, req.CustomerID, req.ProviderID, staffID)

	// Run fraud detection checks
	s.detectFraudOnCreation(req.CustomerID, result.ID, req.ApprovedAmount, staffID, ipAddress)

	// Send email notification to customer (async)
	go func() {
		// Extract customer name from result
		customerName := "Customer"
		if result.Customer != nil {
			customerName = result.Customer.Name
		}

		providerName := "Provider"
		if result.Provider != nil {
			providerName = result.Provider.Name
		}

		validUntil := "No expiration"
		if result.ValidUntil != nil {
			validUntil = result.ValidUntil.Format("January 2, 2006")
		}

		// Get customer email from the query result
		if result.Customer != nil && result.Customer.Email != "" {
			email.SendSubsidyApprovedEmail(result.Customer.Email, customerName, providerName, req.ApprovedAmount, validUntil)
		}
	}()

	return result, nil
}

func (s *SubsidyService) GetSubsidy(ctx context.Context, id uuid.UUID) (*dto.SubsidyDetailResponse, *errLib.CommonError) {
	subsidy, err := s.repo.Queries.GetCustomerSubsidy(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errLib.New("Subsidy not found", http.StatusNotFound)
		}
		log.Printf("Failed to get subsidy: %v", err)
		return nil, errLib.New("Failed to get subsidy", http.StatusInternalServerError)
	}

	// Get usage history
	usageHistory, err := s.repo.Queries.ListUsageTransactions(ctx, id)
	if err != nil {
		log.Printf("Failed to get usage history: %v", err)
		usageHistory = []db.ListUsageTransactionsRow{}
	}

	// Get audit log
	auditLog, err := s.repo.Queries.ListAuditLogsBySubsidy(ctx, uuid.NullUUID{UUID: id, Valid: true})
	if err != nil {
		log.Printf("Failed to get audit log: %v", err)
		auditLog = []db.ListAuditLogsBySubsidyRow{}
	}

	return &dto.SubsidyDetailResponse{
		SubsidyResponse: *mapSubsidyToResponse(subsidy),
		UsageHistory:    mapUsageTransactions(usageHistory),
		AuditLog:        mapAuditLogs(auditLog),
	}, nil
}

func (s *SubsidyService) ListSubsidies(ctx context.Context, filters dto.SubsidyFilters) (*dto.PaginatedResponse, *errLib.CommonError) {
	subsidies, err := s.repo.Queries.ListCustomerSubsidies(ctx, db.ListCustomerSubsidiesParams{
		CustomerID: sqlUUID(filters.CustomerID),
		ProviderID: sqlUUID(filters.ProviderID),
		Status:     sqlString(filters.Status),
		Offset:     int32((filters.Page - 1) * filters.Limit),
		Limit:      int32(filters.Limit),
	})

	if err != nil {
		log.Printf("Failed to list subsidies: %v", err)
		return nil, errLib.New("Failed to list subsidies", http.StatusInternalServerError)
	}

	total, err := s.repo.Queries.CountCustomerSubsidies(ctx, db.CountCustomerSubsidiesParams{
		CustomerID: sqlUUID(filters.CustomerID),
		ProviderID: sqlUUID(filters.ProviderID),
		Status:     sqlString(filters.Status),
	})

	if err != nil {
		log.Printf("Failed to count subsidies: %v", err)
		return nil, errLib.New("Failed to count subsidies", http.StatusInternalServerError)
	}

	result := make([]dto.SubsidyResponse, len(subsidies))
	for i, s := range subsidies {
		result[i] = *mapSubsidyListToResponse(s)
	}

	totalPages := int(total) / filters.Limit
	if int(total)%filters.Limit > 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		Data: result,
		Pagination: dto.PaginationMetadata{
			Total:      total,
			Page:       filters.Page,
			Limit:      filters.Limit,
			TotalPages: totalPages,
		},
	}, nil
}

// ===== BALANCE TRACKING METHODS =====

func (s *SubsidyService) GetActiveSubsidy(ctx context.Context, customerID uuid.UUID) (*dto.SubsidyResponse, *errLib.CommonError) {
	subsidy, err := s.repo.Queries.GetActiveSubsidyForCustomer(ctx, customerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active subsidy is not an error
		}
		log.Printf("Failed to get active subsidy: %v", err)
		return nil, errLib.New("Failed to get active subsidy", http.StatusInternalServerError)
	}

	return mapActiveSubsidyToResponse(subsidy), nil
}

func (s *SubsidyService) GetCustomerBalance(ctx context.Context, customerID uuid.UUID) (*dto.CustomerBalanceResponse, *errLib.CommonError) {
	subsidy, err := s.GetActiveSubsidy(ctx, customerID)
	if err != nil {
		return nil, err
	}

	if subsidy == nil {
		return &dto.CustomerBalanceResponse{
			HasActiveSubsidy: false,
			RemainingBalance: 0,
		}, nil
	}

	providerName := ""
	if subsidy.Provider != nil {
		providerName = subsidy.Provider.Name
	}

	return &dto.CustomerBalanceResponse{
		HasActiveSubsidy: true,
		ProviderName:     &providerName,
		RemainingBalance: subsidy.RemainingBalance,
		ValidUntil:       subsidy.ValidUntil,
	}, nil
}

func (s *SubsidyService) CalculateSubsidyAmount(subsidy *dto.SubsidyResponse, chargeAmount float64) float64 {
	if subsidy == nil || subsidy.RemainingBalance <= 0 {
		return 0
	}

	// Return the minimum of charge amount and remaining balance
	if chargeAmount <= subsidy.RemainingBalance {
		return chargeAmount
	}

	return subsidy.RemainingBalance
}

func (s *SubsidyService) DeactivateSubsidy(ctx context.Context, subsidyID, staffID uuid.UUID, reason, ipAddress string) *errLib.CommonError {
	// SECURITY: Input validation
	if reason == "" {
		return errLib.New("Reason is required for deactivation", http.StatusBadRequest)
	}

	return s.executeInSerializableTx(ctx, func(tx *sql.Tx) *errLib.CommonError {
		queries := s.repo.Queries.WithTx(tx)

		// SECURITY: Verify subsidy exists and get current status
		subsidy, getErr := queries.GetCustomerSubsidy(ctx, subsidyID)
		if getErr != nil {
			if getErr == sql.ErrNoRows {
				return errLib.New("Subsidy not found", http.StatusNotFound)
			}
			log.Printf("Failed to verify subsidy: %v", getErr)
			return errLib.New("Failed to verify subsidy", http.StatusInternalServerError)
		}

		// SECURITY: Verify subsidy can be deactivated
		if subsidy.Status == "depleted" {
			log.Printf("[SECURITY] Attempt to deactivate depleted subsidy: %s", subsidyID)
			return errLib.New("Cannot deactivate depleted subsidy", http.StatusBadRequest)
		}
		if subsidy.Status == "expired" {
			log.Printf("[SECURITY] Attempt to deactivate already expired subsidy: %s", subsidyID)
			return errLib.New("Subsidy is already expired", http.StatusBadRequest)
		}

		previousStatus := subsidy.Status
		remainingBalance := sqlStringToFloat(subsidy.RemainingBalance)

		// Deactivate the subsidy
		_, err := queries.DeactivateSubsidy(ctx, subsidyID)
		if err != nil {
			log.Printf("Failed to deactivate subsidy: %v", err)
			return errLib.New("Failed to deactivate subsidy", http.StatusInternalServerError)
		}

		// Create comprehensive audit log with IP address
		_, auditErr := queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
			CustomerSubsidyID: uuid.NullUUID{UUID: subsidyID, Valid: true},
			Action:            "deactivated",
			PerformedBy:       uuid.NullUUID{UUID: staffID, Valid: true},
			PreviousStatus:    sql.NullString{String: previousStatus, Valid: true},
			NewStatus:         sql.NullString{String: "expired", Valid: true},
			Notes:             sql.NullString{String: fmt.Sprintf("Manually deactivated by staff. Reason: %s. Remaining balance: $%.2f", reason, remainingBalance), Valid: true},
			IpAddress:         sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		})

		if auditErr != nil {
			log.Printf("Warning: Failed to create audit log: %v", auditErr)
		}

		log.Printf("✅ [AUDIT] Deactivated subsidy: ID=%s, Customer=%s, RemainingBalance=$%.2f, Staff=%s, Reason=%s",
			subsidyID, subsidy.CustomerID, remainingBalance, staffID, reason)

		return nil
	})
}

// ===== USAGE RECORDING METHODS =====

func (s *SubsidyService) RecordUsage(ctx context.Context, req *dto.RecordUsageRequest) (*dto.UsageTransactionResponse, *errLib.CommonError) {
	// SECURITY: Input validation
	if req.SubsidyApplied < 0 {
		return nil, errLib.New("Subsidy applied cannot be negative", http.StatusBadRequest)
	}
	if req.OriginalAmount < 0 {
		return nil, errLib.New("Original amount cannot be negative", http.StatusBadRequest)
	}
	if req.CustomerPaid < 0 {
		return nil, errLib.New("Customer paid cannot be negative", http.StatusBadRequest)
	}
	if req.SubsidyApplied > req.OriginalAmount {
		return nil, errLib.New("Subsidy applied cannot exceed original amount", http.StatusBadRequest)
	}

	var result *dto.UsageTransactionResponse

	// SECURITY: Use serializable isolation to prevent race conditions
	err := s.executeInSerializableTx(ctx, func(tx *sql.Tx) *errLib.CommonError {
		queries := s.repo.Queries.WithTx(tx)

		// SECURITY: Verify subsidy belongs to customer and has sufficient balance
		subsidy, getErr := queries.GetCustomerSubsidy(ctx, req.SubsidyID)
		if getErr != nil {
			if getErr == sql.ErrNoRows {
				return errLib.New("Subsidy not found", http.StatusNotFound)
			}
			return errLib.New("Failed to verify subsidy", http.StatusInternalServerError)
		}

		// SECURITY: Verify customer ownership
		if subsidy.CustomerID != req.CustomerID {
			log.Printf("[SECURITY] Subsidy usage attempt by wrong customer: subsidy %s belongs to %s, attempted by %s",
				req.SubsidyID, subsidy.CustomerID, req.CustomerID)
			return errLib.New("Access denied: subsidy does not belong to this customer", http.StatusForbidden)
		}

		// SECURITY: Verify subsidy is active
		if subsidy.Status != "active" && subsidy.Status != "approved" {
			return errLib.New(fmt.Sprintf("Subsidy is not active (status: %s)", subsidy.Status), http.StatusBadRequest)
		}

		// SECURITY: Verify sufficient balance
		remainingBalance := sqlStringToFloat(subsidy.RemainingBalance)
		if req.SubsidyApplied > remainingBalance+FLOAT_PRECISION_TOLERANCE {
			log.Printf("[SECURITY] Insufficient subsidy balance: requested %.2f, available %.2f", req.SubsidyApplied, remainingBalance)
			return errLib.New(fmt.Sprintf("Insufficient subsidy balance: requested $%.2f, available $%.2f",
				req.SubsidyApplied, remainingBalance), http.StatusBadRequest)
		}

		// Create usage transaction
		usage, dbErr := queries.CreateUsageTransaction(ctx, db.CreateUsageTransactionParams{
			CustomerSubsidyID:     req.SubsidyID,
			CustomerID:            req.CustomerID,
			TransactionType:       req.TransactionType,
			MembershipPlanID:      sqlUUID(req.MembershipPlanID),
			OriginalAmount:        decimal.NewFromFloat(req.OriginalAmount),
			SubsidyApplied:        decimal.NewFromFloat(req.SubsidyApplied),
			CustomerPaid:          decimal.NewFromFloat(req.CustomerPaid),
			StripeSubscriptionID:  sqlString(req.StripeSubscriptionID),
			StripeInvoiceID:       sqlString(req.StripeInvoiceID),
			StripePaymentIntentID: sqlString(req.StripePaymentIntentID),
			Description:           sql.NullString{String: req.Description, Valid: true},
		})

		if dbErr != nil {
			log.Printf("Failed to create usage transaction: %v", dbErr)
			return errLib.New("Failed to record usage", http.StatusInternalServerError)
		}

		// Update subsidy total_amount_used
		updatedSubsidy, updateErr := queries.UpdateSubsidyUsage(ctx, db.UpdateSubsidyUsageParams{
			ID:              req.SubsidyID,
			TotalAmountUsed: decimal.NewFromFloat(req.SubsidyApplied),
		})

		if updateErr != nil {
			log.Printf("Failed to update subsidy usage: %v", updateErr)
			return errLib.New("Failed to update subsidy", http.StatusInternalServerError)
		}

		// Check if subsidy is now depleted
		remainingBalance = sqlStringToFloat(updatedSubsidy.RemainingBalance)
		isDepleted := remainingBalance <= FLOAT_PRECISION_TOLERANCE

		if isDepleted {
			_, depletedErr := queries.MarkSubsidyAsDepleted(ctx, req.SubsidyID)
			if depletedErr != nil {
				log.Printf("Warning: Failed to mark subsidy as depleted: %v", depletedErr)
			} else {
				log.Printf("🔴 Subsidy depleted: %s", req.SubsidyID)
			}

			// Create audit log for depletion
			_, auditErr := queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
				CustomerSubsidyID: uuid.NullUUID{UUID: req.SubsidyID, Valid: true},
				Action:            "depleted",
				PreviousStatus:    sql.NullString{String: "active", Valid: true},
				NewStatus:         sql.NullString{String: "depleted", Valid: true},
				Notes:             sql.NullString{String: "Subsidy balance fully used", Valid: true},
			})

			if auditErr != nil {
				log.Printf("Warning: Failed to create audit log: %v", auditErr)
			}
		}

		// Run fraud detection on usage
		s.detectFraudOnUsage(req.CustomerID, req.SubsidyID, req.SubsidyApplied, remainingBalance, timeFromSQL(subsidy.CreatedAt))

		// Get customer info for email notifications
		customerEmail := stringFromSQL(subsidy.CustomerEmail)
		customerName := interfaceToString(subsidy.CustomerName)
		if customerName == "" {
			customerName = "Customer"
		}

		// Send appropriate email notification (async)
		go func() {
			if customerEmail != "" {
				if isDepleted {
					// Send depleted email
					email.SendSubsidyDepletedEmail(customerEmail, customerName, decimalToFloat(updatedSubsidy.TotalAmountUsed))
				} else if req.SubsidyApplied > 0 {
					// Send usage email (only if significant amount was used)
					email.SendSubsidyUsedEmail(customerEmail, customerName, req.SubsidyApplied, remainingBalance, req.TransactionType)
				}
			}
		}()

		// Track payment in centralized payment tracking system
		go func() {
			_, trackingErr := s.paymentTracking.TrackPayment(context.Background(), tracking.TrackPaymentParams{
				CustomerID:           req.CustomerID,
				CustomerEmail:        customerEmail,
				CustomerName:         customerName,
				TransactionType:      req.TransactionType,
				TransactionDate:      time.Now(),
				OriginalAmount:       req.OriginalAmount,
				DiscountAmount:       0,
				SubsidyAmount:        req.SubsidyApplied,
				CustomerPaid:         req.CustomerPaid,
				MembershipPlanID:     req.MembershipPlanID,
				SubsidyID:            &req.SubsidyID,
				StripeSubscriptionID: stringFromSQL(usage.StripeSubscriptionID),
				StripeInvoiceID:      stringFromSQL(usage.StripeInvoiceID),
				StripePaymentIntentID: stringFromSQL(usage.StripePaymentIntentID),
				PaymentStatus:        "completed",
				Currency:             "USD",
				Description:          stringFromSQL(usage.Description),
			})
			if trackingErr != nil {
				log.Printf("Warning: Failed to track subsidy payment: %v", trackingErr)
			}
		}()

		result = &dto.UsageTransactionResponse{
			ID:              usage.ID,
			Date:            timeFromSQL(usage.AppliedAt),
			TransactionType: usage.TransactionType,
			Description:     stringFromSQL(usage.Description),
			OriginalAmount:  decimalToFloat(usage.OriginalAmount),
			SubsidyApplied:  decimalToFloat(usage.SubsidyApplied),
			CustomerPaid:    decimalToFloat(usage.CustomerPaid),
			StripeInvoiceID: stringPtrFromSQL(usage.StripeInvoiceID),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("✅ Recorded subsidy usage: $%.2f applied, $%.2f paid by customer", req.SubsidyApplied, req.CustomerPaid)
	return result, nil
}

func (s *SubsidyService) GetCustomerUsageHistory(ctx context.Context, customerID uuid.UUID, page, limit int) (*dto.PaginatedResponse, *errLib.CommonError) {
	usage, err := s.repo.Queries.ListUsageTransactionsByCustomer(ctx, db.ListUsageTransactionsByCustomerParams{
		CustomerID: customerID,
		Limit:      int32(limit),
		Offset:     int32((page - 1) * limit),
	})

	if err != nil {
		log.Printf("Failed to get usage history: %v", err)
		return nil, errLib.New("Failed to get usage history", http.StatusInternalServerError)
	}

	total, err := s.repo.Queries.CountUsageTransactionsByCustomer(ctx, customerID)
	if err != nil {
		log.Printf("Failed to count usage: %v", err)
		return nil, errLib.New("Failed to count usage", http.StatusInternalServerError)
	}

	result := make([]dto.UsageTransactionResponse, len(usage))
	for i, u := range usage {
		result[i] = dto.UsageTransactionResponse{
			ID:                 u.ID,
			Date:               timeFromSQL(u.AppliedAt),
			TransactionType:    u.TransactionType,
			MembershipPlanName: stringPtrFromSQL(u.MembershipPlanName),
			Description:        stringFromSQL(u.Description),
			OriginalAmount:     decimalToFloat(u.OriginalAmount),
			SubsidyApplied:     decimalToFloat(u.SubsidyApplied),
			CustomerPaid:       decimalToFloat(u.CustomerPaid),
			StripeInvoiceID:    stringPtrFromSQL(u.StripeInvoiceID),
		}
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		Data: result,
		Pagination: dto.PaginationMetadata{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *SubsidyService) GetSubsidySummary(ctx context.Context) (*dto.SubsidySummaryResponse, *errLib.CommonError) {
	summary, err := s.repo.Queries.GetSubsidySummary(ctx)
	if err != nil {
		log.Printf("Failed to get subsidy summary: %v", err)
		return nil, errLib.New("Failed to get summary", http.StatusInternalServerError)
	}

	return &dto.SubsidySummaryResponse{
		ActiveCount:    summary.ActiveCount,
		PendingCount:   summary.PendingCount,
		DepletedCount:  summary.DepletedCount,
		TotalApproved:  interfaceToFloat(summary.TotalApproved),
		TotalUsed:      interfaceToFloat(summary.TotalUsed),
		TotalRemaining: interfaceToFloat(summary.TotalRemaining),
	}, nil
}

// ===== HELPER FUNCTIONS =====

func mapProviderToResponse(p db.SubsidiesProvider) *dto.ProviderResponse {
	return &dto.ProviderResponse{
		ID:           p.ID,
		Name:         p.Name,
		ContactEmail: stringPtrFromSQL(p.ContactEmail),
		ContactPhone: stringPtrFromSQL(p.ContactPhone),
		IsActive:     boolFromSQL(p.IsActive),
		CreatedAt:    timeFromSQL(p.CreatedAt),
		UpdatedAt:    timeFromSQL(p.UpdatedAt),
	}
}

func mapSubsidyToResponse(s db.GetCustomerSubsidyRow) *dto.SubsidyResponse {
	return &dto.SubsidyResponse{
		ID: s.ID,
		Customer: &dto.CustomerSummary{
			ID:    s.CustomerID,
			Name:  interfaceToString(s.CustomerName),
			Email: stringFromSQL(s.CustomerEmail),
		},
		Provider: &dto.ProviderSummary{
			ID:   uuidFromSQL(s.ProviderID),
			Name: stringFromSQL(s.ProviderName),
		},
		ApprovedAmount:   decimalToFloat(s.ApprovedAmount),
		TotalAmountUsed:  decimalToFloat(s.TotalAmountUsed),
		RemainingBalance: sqlStringToFloat(s.RemainingBalance),
		Status:           s.Status,
		ValidFrom:        s.ValidFrom,
		ValidUntil:       timePtrFromSQL(s.ValidUntil),
		Reason:           stringFromSQL(s.Reason),
		AdminNotes:       stringFromSQL(s.AdminNotes),
		ApprovedBy:       interfaceToString(s.ApprovedByName),
		ApprovedAt:       timePtrFromSQL(s.ApprovedAt),
		CreatedAt:        timeFromSQL(s.CreatedAt),
		UpdatedAt:        timeFromSQL(s.UpdatedAt),
	}
}

func mapSubsidyListToResponse(s db.ListCustomerSubsidiesRow) *dto.SubsidyResponse {
	return &dto.SubsidyResponse{
		ID: s.ID,
		Customer: &dto.CustomerSummary{
			ID:    s.CustomerID,
			Name:  interfaceToString(s.CustomerName),
			Email: stringFromSQL(s.CustomerEmail),
		},
		Provider: &dto.ProviderSummary{
			ID:   uuidFromSQL(s.ProviderID),
			Name: stringFromSQL(s.ProviderName),
		},
		ApprovedAmount:   decimalToFloat(s.ApprovedAmount),
		TotalAmountUsed:  decimalToFloat(s.TotalAmountUsed),
		RemainingBalance: sqlStringToFloat(s.RemainingBalance),
		Status:           s.Status,
		ValidFrom:        s.ValidFrom,
		ValidUntil:       timePtrFromSQL(s.ValidUntil),
		Reason:           stringFromSQL(s.Reason),
		AdminNotes:       stringFromSQL(s.AdminNotes),
		ApprovedBy:       interfaceToString(s.ApprovedByName),
		ApprovedAt:       timePtrFromSQL(s.ApprovedAt),
		CreatedAt:        timeFromSQL(s.CreatedAt),
		UpdatedAt:        timeFromSQL(s.UpdatedAt),
	}
}

func mapActiveSubsidyToResponse(s db.GetActiveSubsidyForCustomerRow) *dto.SubsidyResponse {
	return &dto.SubsidyResponse{
		ID: s.ID,
		Provider: &dto.ProviderSummary{
			ID:   uuidFromSQL(s.ProviderID),
			Name: stringFromSQL(s.ProviderName),
		},
		ApprovedAmount:   decimalToFloat(s.ApprovedAmount),
		TotalAmountUsed:  decimalToFloat(s.TotalAmountUsed),
		RemainingBalance: sqlStringToFloat(s.RemainingBalance),
		Status:           s.Status,
		ValidFrom:        s.ValidFrom,
		ValidUntil:       timePtrFromSQL(s.ValidUntil),
		CreatedAt:        timeFromSQL(s.CreatedAt),
		UpdatedAt:        timeFromSQL(s.UpdatedAt),
	}
}

func mapUsageTransactions(transactions []db.ListUsageTransactionsRow) []dto.UsageTransactionResponse {
	result := make([]dto.UsageTransactionResponse, len(transactions))
	for i, t := range transactions {
		result[i] = dto.UsageTransactionResponse{
			ID:                 t.ID,
			Date:               timeFromSQL(t.AppliedAt),
			TransactionType:    t.TransactionType,
			MembershipPlanName: stringPtrFromSQL(t.MembershipPlanName),
			Description:        stringFromSQL(t.Description),
			OriginalAmount:     decimalToFloat(t.OriginalAmount),
			SubsidyApplied:     decimalToFloat(t.SubsidyApplied),
			CustomerPaid:       decimalToFloat(t.CustomerPaid),
			StripeInvoiceID:    stringPtrFromSQL(t.StripeInvoiceID),
		}
	}
	return result
}

func mapAuditLogs(logs []db.ListAuditLogsBySubsidyRow) []dto.AuditLogResponse {
	result := make([]dto.AuditLogResponse, len(logs))
	for i, l := range logs {
		result[i] = dto.AuditLogResponse{
			ID:             l.ID,
			Action:         l.Action,
			PerformedBy:    interfaceToStringPtr(l.PerformedByName),
			PreviousStatus: stringPtrFromSQL(l.PreviousStatus),
			NewStatus:      stringPtrFromSQL(l.NewStatus),
			AmountChanged:  floatPtrFromDecimalSQL(l.AmountChanged),
			Notes:          stringPtrFromSQL(l.Notes),
			IPAddress:      stringPtrFromSQL(l.IpAddress),
			CreatedAt:      timeFromSQL(l.CreatedAt),
		}
	}
	return result
}

// SQL helper functions
func sqlString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func sqlBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func sqlUUID(u *uuid.UUID) uuid.NullUUID {
	if u == nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: *u, Valid: true}
}

func sqlTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func stringFromSQL(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func stringPtrFromSQL(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func boolFromSQL(b sql.NullBool) bool {
	if b.Valid {
		return b.Bool
	}
	return false
}

func uuidFromSQL(u uuid.NullUUID) uuid.UUID {
	if u.Valid {
		return u.UUID
	}
	return uuid.Nil
}

func uuidPtrFromSQL(u uuid.NullUUID) *uuid.UUID {
	if u.Valid {
		return &u.UUID
	}
	return nil
}

func timeFromSQL(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func timePtrFromSQL(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func interfaceToString(i interface{}) string {
	if i == nil {
		return ""
	}

	switch v := i.(type) {
	case string:
		return v
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return ""
	default:
		return fmt.Sprintf("%v", i)
	}
}

func interfaceToStringPtr(i interface{}) *string {
	if i == nil {
		return nil
	}

	switch v := i.(type) {
	case string:
		if v == "" {
			return nil
		}
		return &v
	case sql.NullString:
		if v.Valid {
			return &v.String
		}
		return nil
	default:
		s := fmt.Sprintf("%v", i)
		if s == "" || s == "<nil>" {
			return nil
		}
		return &s
	}
}

func decimalToFloat(d decimal.Decimal) float64 {
	f, _ := d.Float64()
	return f
}

func interfaceToFloat(i interface{}) float64 {
	if i == nil {
		return 0.0
	}

	switch v := i.(type) {
	case float64:
		return v
	case string:
		f, _ := decimal.NewFromString(v)
		return decimalToFloat(f)
	case decimal.Decimal:
		return decimalToFloat(v)
	default:
		return 0.0
	}
}

func sqlStringToFloat(s sql.NullString) float64 {
	if !s.Valid {
		return 0.0
	}
	d, err := decimal.NewFromString(s.String)
	if err != nil {
		return 0.0
	}
	return decimalToFloat(d)
}

func floatPtrFromDecimalSQL(s sql.NullString) *float64 {
	if s.Valid {
		var f float64
		fmt.Sscanf(s.String, "%f", &f)
		return &f
	}
	return nil
}
