# Plan: Auto-Create Stripe Products When Creating Membership Plans

## Goal
When an admin creates a membership plan in the app, automatically create the corresponding Stripe Product and Price, instead of requiring manual creation in the Stripe dashboard first.

---

## Current Flow (What Happens Now)
1. Admin manually creates a Product and Price in Stripe Dashboard
2. Admin copies the `stripe_price_id` (e.g., `price_1ABC123...`)
3. Admin creates membership plan in the app, pasting the `stripe_price_id`
4. Database stores the `stripe_price_id` reference

## New Flow (What We Want)
1. Admin creates membership plan in the app with pricing details (amount, currency, interval)
2. App automatically creates Stripe Product + Price via API
3. App stores the returned `stripe_price_id` in the database
4. (Optional) If joining fee is provided, create a second one-time Price for that too

---

## Files to Modify

### 1. Stripe Service - Add Product/Price Creation Functions
**File:** `internal/domains/payment/services/stripe/stripe.go`

Add these new functions after the existing `PriceService`:

```go
// ProductService handles Stripe product and price creation
type ProductService struct{}

func NewProductService() *ProductService {
    return &ProductService{}
}

// CreateProductWithRecurringPrice creates a Stripe Product and a recurring Price
// Returns the stripe_price_id to store in the database
func (s *ProductService) CreateProductWithRecurringPrice(
    productName string,        // e.g., "Monthly Membership - Gold"
    productDescription string, // Optional description
    unitAmount int64,          // Price in cents (e.g., 5000 for $50.00)
    currency string,           // e.g., "cad" or "usd"
    interval string,           // "month", "year", "week", "day"
    intervalCount int64,       // e.g., 1 for every month, 3 for every 3 months
) (stripePriceID string, stripeProductID string, err *errLib.CommonError)

// CreateOneTimePrice creates a one-time Price for an existing Product (for joining fees)
// Returns the stripe_price_id for the one-time fee
func (s *ProductService) CreateOneTimePrice(
    stripeProductID string, // The product ID to attach this price to
    unitAmount int64,       // Price in cents
    currency string,        // e.g., "cad"
    nickname string,        // e.g., "Joining Fee"
) (stripePriceID string, err *errLib.CommonError)
```

**Implementation details:**
- Use `github.com/stripe/stripe-go/v81/product` for product creation
- Use `github.com/stripe/stripe-go/v81/price` for price creation
- Add timeout handling like existing Stripe functions
- Log the created IDs for debugging

---

### 2. Update DTO - Change Request Structure
**File:** `internal/domains/membership/dto/membership_plan/requests.go`

**Current DTO:**
```go
type PlanRequestDto struct {
    MembershipID        uuid.UUID `json:"membership_id" validate:"required"`
    Name                string    `json:"name" validate:"notwhitespace"`
    AmtPeriods          *int32    `json:"amt_periods" validate:"omitempty,gt=0"`
    StripePriceID       string    `json:"stripe_price_id" validate:"required,notwhitespace"`  // REMOVE required
    StripeJoiningFeesID string    `json:"stripe_joining_fees_id"`
    CreditAllocation    *int32    `json:"credit_allocation" validate:"omitempty,gte=0"`
    WeeklyCreditLimit   *int32    `json:"weekly_credit_limit" validate:"omitempty,gte=0"`
}
```

**New DTO:**
```go
type PlanRequestDto struct {
    MembershipID        uuid.UUID `json:"membership_id" validate:"required"`
    Name                string    `json:"name" validate:"notwhitespace"`
    AmtPeriods          *int32    `json:"amt_periods" validate:"omitempty,gt=0"`
    CreditAllocation    *int32    `json:"credit_allocation" validate:"omitempty,gte=0"`
    WeeklyCreditLimit   *int32    `json:"weekly_credit_limit" validate:"omitempty,gte=0"`

    // Option 1: Provide existing Stripe Price ID (legacy/manual way)
    StripePriceID       string    `json:"stripe_price_id"`
    StripeJoiningFeesID string    `json:"stripe_joining_fees_id"`

    // Option 2: Provide pricing details to auto-create in Stripe (new way)
    UnitAmount          *int64    `json:"unit_amount"`           // Price in cents (e.g., 5000 = $50.00)
    Currency            string    `json:"currency"`              // "cad" or "usd", defaults to "cad"
    BillingInterval     string    `json:"billing_interval"`      // "month", "year", "week", "day"
    IntervalCount       *int64    `json:"interval_count"`        // defaults to 1
    JoiningFeeAmount    *int64    `json:"joining_fee_amount"`    // Optional one-time fee in cents
}
```

**Validation logic in `ToCreateValueObjects()`:**
- If `StripePriceID` is provided, use it directly (existing behavior)
- If `StripePriceID` is empty, require `UnitAmount`, `Currency`, and `BillingInterval`
- Return error if neither option is fully provided

---

### 3. Update Values Struct
**File:** `internal/domains/membership/values/plans.go`

Add new fields to `PlanCreateValues`:
```go
type PlanCreateValues struct {
    PlanDetails
    // New fields for Stripe auto-creation
    UnitAmount       *int64  // Price in cents
    Currency         string  // "cad" or "usd"
    BillingInterval  string  // "month", "year", etc.
    IntervalCount    *int64  // defaults to 1
    JoiningFeeAmount *int64  // optional one-time fee
}
```

---

### 4. Update Service - Add Stripe Creation Logic
**File:** `internal/domains/membership/services/membership_plan.go`

Modify `CreateMembershipPlan` function:

```go
func (s *PlanService) CreateMembershipPlan(ctx context.Context, details values.PlanCreateValues) *errLib.CommonError {
    return s.executeInTx(ctx, func(txRepo *repo.PlansRepository) *errLib.CommonError {

        // NEW: If no StripePriceID provided, create product/price in Stripe
        if details.StripePriceID == "" {
            if details.UnitAmount == nil || details.BillingInterval == "" {
                return errLib.New("Either stripe_price_id or pricing details (unit_amount, billing_interval) required", http.StatusBadRequest)
            }

            // Set defaults
            currency := details.Currency
            if currency == "" {
                currency = "cad"
            }
            intervalCount := int64(1)
            if details.IntervalCount != nil {
                intervalCount = *details.IntervalCount
            }

            // Create Stripe Product + recurring Price
            priceID, productID, err := s.productService.CreateProductWithRecurringPrice(
                details.Name,           // Product name
                "",                     // Description (optional)
                *details.UnitAmount,    // Price in cents
                currency,
                details.BillingInterval,
                intervalCount,
            )
            if err != nil {
                return err
            }

            details.StripePriceID = priceID
            log.Printf("Created Stripe product %s with price %s for plan %s", productID, priceID, details.Name)

            // If joining fee provided, create one-time price
            if details.JoiningFeeAmount != nil && *details.JoiningFeeAmount > 0 {
                joiningFeePriceID, err := s.productService.CreateOneTimePrice(
                    productID,
                    *details.JoiningFeeAmount,
                    currency,
                    "Joining Fee",
                )
                if err != nil {
                    return err
                }
                details.StripeJoiningFeesID = joiningFeePriceID
                details.JoiningFee = int(*details.JoiningFeeAmount)
                log.Printf("Created Stripe joining fee price %s", joiningFeePriceID)
            }
        }

        // Existing code continues...
        if err := txRepo.CreateMembershipPlan(ctx, details); err != nil {
            return err
        }
        // ... rest of existing code
    })
}
```

Also update the `PlanService` struct to include the new service:
```go
type PlanService struct {
    repo                     *repo.PlansRepository
    staffActivityLogsService *staffActivityLogs.Service
    stripeService            *stripeService.PriceService
    productService           *stripeService.ProductService  // NEW
    db                       *sql.DB
}
```

And update `NewPlanService`:
```go
func NewPlanService(container *di.Container) *PlanService {
    return &PlanService{
        repo:                     repo.NewMembershipPlansRepository(container),
        staffActivityLogsService: staffActivityLogs.NewService(container),
        stripeService:            stripeService.NewPriceService(),
        productService:           stripeService.NewProductService(),  // NEW
        db:                       container.DB,
    }
}
```

---

### 5. Add Stripe SDK Import
**File:** `internal/domains/payment/services/stripe/stripe.go`

Add import at the top:
```go
import (
    // ... existing imports
    "github.com/stripe/stripe-go/v81/product"  // NEW
)
```

---

## Implementation Steps (In Order)

### Step 1: Create ProductService in Stripe package
Add to `internal/domains/payment/services/stripe/stripe.go`:
- `ProductService` struct
- `NewProductService()` constructor
- `CreateProductWithRecurringPrice()` function
- `CreateOneTimePrice()` function

### Step 2: Update the DTO
Modify `internal/domains/membership/dto/membership_plan/requests.go`:
- Add new fields to `PlanRequestDto`
- Update `ToCreateValueObjects()` with validation logic

### Step 3: Update the Values struct
Modify `internal/domains/membership/values/plans.go`:
- Add new fields to `PlanCreateValues`

### Step 4: Update the Service
Modify `internal/domains/membership/services/membership_plan.go`:
- Add `productService` to `PlanService` struct
- Update `NewPlanService()` to initialize it
- Update `CreateMembershipPlan()` with Stripe creation logic

### Step 5: Regenerate SQLC (if needed)
```bash
cd internal/domains/membership/persistence/sqlc && sqlc generate
```

### Step 6: Test
Test both flows:
1. **Legacy flow:** Provide `stripe_price_id` directly (should work as before)
2. **New flow:** Provide `unit_amount`, `currency`, `billing_interval` (should create in Stripe)

---

## API Request Examples

### Legacy Way (still supported):
```json
POST /memberships/plans
{
    "membership_id": "uuid-here",
    "name": "Monthly Gold",
    "stripe_price_id": "price_1ABC123...",
    "stripe_joining_fees_id": "price_1XYZ789...",
    "amt_periods": 12,
    "credit_allocation": 100,
    "weekly_credit_limit": 10
}
```

### New Way (auto-create in Stripe):
```json
POST /memberships/plans
{
    "membership_id": "uuid-here",
    "name": "Monthly Gold",
    "unit_amount": 5000,
    "currency": "cad",
    "billing_interval": "month",
    "interval_count": 1,
    "joining_fee_amount": 15000,
    "amt_periods": 12,
    "credit_allocation": 100,
    "weekly_credit_limit": 10
}
```

This will:
1. Create Stripe Product named "Monthly Gold"
2. Create recurring Price of $50.00 CAD/month attached to that product
3. Create one-time Price of $150.00 CAD for joining fee
4. Store both price IDs in database

---

## Rollback Considerations
- If database insert fails after Stripe creation, the Stripe products/prices will still exist (orphaned)
- Consider: Add cleanup logic to archive/delete Stripe prices on failure
- Or: Accept orphaned prices (they don't cost anything, just clutter)

---

## Future Enhancements (Not in this PR)
- Add `stripe_product_id` column to database for better tracking
- Add endpoint to update Stripe prices when plan is updated
- Add endpoint to archive/deactivate Stripe prices when plan is deleted
