package dto

import (
	"time"

	"github.com/google/uuid"
)

// ===== PROVIDER DTOs =====

type CreateProviderRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=200"`
	ContactEmail *string `json:"contact_email" validate:"omitempty,email"`
	ContactPhone *string `json:"contact_phone" validate:"omitempty,min=10,max=20"`
}

type UpdateProviderRequest struct {
	Name         *string `json:"name" validate:"omitempty,min=2,max=200"`
	ContactEmail *string `json:"contact_email" validate:"omitempty,email"`
	ContactPhone *string `json:"contact_phone" validate:"omitempty,min=10,max=20"`
	IsActive     *bool   `json:"is_active"`
}

type ProviderResponse struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	ContactEmail *string    `json:"contact_email,omitempty"`
	ContactPhone *string    `json:"contact_phone,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ProviderStatsResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	TotalSubsidies    int64     `json:"total_subsidies"`
	TotalAmountIssued float64   `json:"total_amount_issued"`
	TotalAmountUsed   float64   `json:"total_amount_used"`
	TotalRemaining    float64   `json:"total_remaining"`
}

// ===== CUSTOMER SUBSIDY DTOs =====

type CreateSubsidyRequest struct {
	CustomerID     uuid.UUID  `json:"customer_id" validate:"required"`
	ProviderID     uuid.UUID  `json:"provider_id" validate:"required"`
	ApprovedAmount float64    `json:"approved_amount" validate:"required,gt=0"`
	ValidUntil     *time.Time `json:"valid_until"`
	Reason         string     `json:"reason" validate:"required,min=5,max=500"`
	AdminNotes     string     `json:"admin_notes" validate:"max=1000"`
}

type UpdateSubsidyRequest struct {
	ValidUntil *time.Time `json:"valid_until"`
	AdminNotes *string    `json:"admin_notes" validate:"omitempty,max=1000"`
}

type DeactivateSubsidyRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500"`
}

type SubsidyFilters struct {
	CustomerID *uuid.UUID
	ProviderID *uuid.UUID
	Status     *string
	Page       int
	Limit      int
}

type SubsidyResponse struct {
	ID               uuid.UUID        `json:"id"`
	Customer         *CustomerSummary `json:"customer,omitempty"`
	Provider         *ProviderSummary `json:"provider,omitempty"`
	ApprovedAmount   float64          `json:"approved_amount"`
	TotalAmountUsed  float64          `json:"total_amount_used"`
	RemainingBalance float64          `json:"remaining_balance"`
	Status           string           `json:"status"`
	ValidFrom        time.Time        `json:"valid_from"`
	ValidUntil       *time.Time       `json:"valid_until,omitempty"`
	Reason           string           `json:"reason,omitempty"`
	AdminNotes       string           `json:"admin_notes,omitempty"`
	ApprovedBy       *uuid.UUID       `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time       `json:"approved_at,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type SubsidyDetailResponse struct {
	SubsidyResponse
	UsageHistory []UsageTransactionResponse `json:"usage_history"`
	AuditLog     []AuditLogResponse         `json:"audit_log,omitempty"`
}

type CustomerSummary struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type ProviderSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ===== USAGE TRANSACTION DTOs =====

type RecordUsageRequest struct {
	SubsidyID            uuid.UUID  `json:"subsidy_id" validate:"required,uuid"`
	CustomerID           uuid.UUID  `json:"customer_id" validate:"required,uuid"`
	TransactionType      string     `json:"transaction_type" validate:"required"`
	MembershipPlanID     *uuid.UUID `json:"membership_plan_id" validate:"omitempty,uuid"`
	OriginalAmount       float64    `json:"original_amount" validate:"required,gte=0"`
	SubsidyApplied       float64    `json:"subsidy_applied" validate:"required,gte=0"`
	CustomerPaid         float64    `json:"customer_paid" validate:"required,gte=0"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id"`
	StripeInvoiceID      *string    `json:"stripe_invoice_id"`
	StripePaymentIntentID *string   `json:"stripe_payment_intent_id"`
	Description          string     `json:"description"`
}

type UsageTransactionResponse struct {
	ID                   uuid.UUID  `json:"id"`
	Date                 time.Time  `json:"date"`
	TransactionType      string     `json:"transaction_type"`
	MembershipPlanName   *string    `json:"membership_plan_name,omitempty"`
	Description          string     `json:"description"`
	OriginalAmount       float64    `json:"original_amount"`
	SubsidyApplied       float64    `json:"subsidy_applied"`
	CustomerPaid         float64    `json:"customer_paid"`
	StripeInvoiceID      *string    `json:"stripe_invoice_id,omitempty"`
}

type UsageStatsResponse struct {
	TransactionCount    int64   `json:"transaction_count"`
	TotalSubsidyUsed    float64 `json:"total_subsidy_used"`
	TotalCustomerPaid   float64 `json:"total_customer_paid"`
	TotalOriginalAmount float64 `json:"total_original_amount"`
}

// ===== AUDIT LOG DTOs =====

type AuditLogResponse struct {
	ID              uuid.UUID  `json:"id"`
	Action          string     `json:"action"`
	PerformedBy     *string    `json:"performed_by,omitempty"`
	PreviousStatus  *string    `json:"previous_status,omitempty"`
	NewStatus       *string    `json:"new_status,omitempty"`
	AmountChanged   *float64   `json:"amount_changed,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	IPAddress       *string    `json:"ip_address,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ===== SUMMARY DTOs =====

type SubsidySummaryResponse struct {
	ActiveCount    int64   `json:"active_count"`
	PendingCount   int64   `json:"pending_count"`
	DepletedCount  int64   `json:"depleted_count"`
	TotalApproved  float64 `json:"total_approved"`
	TotalUsed      float64 `json:"total_used"`
	TotalRemaining float64 `json:"total_remaining"`
}

// ===== PAGINATION DTOs =====

type PaginatedResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

type PaginationMetadata struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"total_pages"`
}

// ===== BALANCE CHECK DTO (for customers) =====

type CustomerBalanceResponse struct {
	HasActiveSubsidy bool       `json:"has_active_subsidy"`
	ProviderName     *string    `json:"provider_name,omitempty"`
	RemainingBalance float64    `json:"remaining_balance"`
	ValidUntil       *time.Time `json:"valid_until,omitempty"`
}
