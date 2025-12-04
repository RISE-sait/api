package service

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/credit_package/dto"
	repo "api/internal/domains/credit_package/persistence/repository"
	"api/internal/domains/payment/services/stripe"
	userServices "api/internal/domains/user/services"
	errLib "api/internal/libs/errors"
	contextUtils "api/utils/context"

	"github.com/google/uuid"
)

type CreditPackageService struct {
	CreditPackageRepo *repo.CreditPackageRepository
	CreditService     *userServices.CustomerCreditService
	StripeService     *stripe.PriceService
	ProductService    *stripe.ProductService
	DB                *sql.DB
}

func NewCreditPackageService(container *di.Container) *CreditPackageService {
	return &CreditPackageService{
		CreditPackageRepo: repo.NewCreditPackageRepository(container),
		CreditService:     userServices.NewCustomerCreditService(container),
		StripeService:     stripe.NewPriceService(),
		ProductService:    stripe.NewProductService(),
		DB:                container.DB,
	}
}

// getExistingStripeCustomerID retrieves the existing Stripe customer ID for a user from the database
func (s *CreditPackageService) getExistingStripeCustomerID(ctx context.Context, userID uuid.UUID) *string {
	var stripeCustomerID sql.NullString
	query := "SELECT stripe_customer_id FROM users.users WHERE id = $1"
	err := s.DB.QueryRowContext(ctx, query, userID).Scan(&stripeCustomerID)
	if err != nil || !stripeCustomerID.Valid || stripeCustomerID.String == "" {
		return nil
	}
	return &stripeCustomerID.String
}

// GetAllPackages returns all available credit packages with live Stripe pricing
func (s *CreditPackageService) GetAllPackages(ctx context.Context) ([]dto.CreditPackageResponse, *errLib.CommonError) {
	packages, err := s.CreditPackageRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Enrich with Stripe price data
	for i, pkg := range packages {
		if pkg.StripePriceID != "" {
			stripePrice, stripeErr := s.StripeService.GetPrice(pkg.StripePriceID)
			if stripeErr != nil {
				log.Printf("Warning: Failed to fetch Stripe price for package %s (price_id: %s): %v",
					pkg.ID, pkg.StripePriceID, stripeErr)
				// Continue with empty price data if Stripe fails
				continue
			}

			// Update package with live Stripe data (convert cents to dollars)
			packages[i].Price = float64(stripePrice.UnitAmount) / 100
			packages[i].Currency = string(stripePrice.Currency)
		}
	}

	return packages, nil
}

// GetPackageByID returns a specific credit package by ID with live Stripe pricing
func (s *CreditPackageService) GetPackageByID(ctx context.Context, id uuid.UUID) (*dto.CreditPackageResponse, *errLib.CommonError) {
	pkg, err := s.CreditPackageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Enrich with Stripe price data
	if pkg.StripePriceID != "" {
		stripePrice, stripeErr := s.StripeService.GetPrice(pkg.StripePriceID)
		if stripeErr != nil {
			log.Printf("Warning: Failed to fetch Stripe price for package %s (price_id: %s): %v",
				pkg.ID, pkg.StripePriceID, stripeErr)
			// Return package with empty price data if Stripe fails
			return pkg, nil
		}

		// Update package with live Stripe data (convert cents to dollars)
		pkg.Price = float64(stripePrice.UnitAmount) / 100
		pkg.Currency = string(stripePrice.Currency)
	}

	return pkg, nil
}

// CheckoutCreditPackage creates a Stripe checkout session for purchasing a credit package
func (s *CreditPackageService) CheckoutCreditPackage(ctx context.Context, packageID uuid.UUID, successURL string, cancelURL string) (string, *errLib.CommonError) {
	// Get customer ID from context
	customerID, err := contextUtils.GetUserID(ctx)
	if err != nil {
		return "", err
	}

	// VALIDATION: Customer must have 0 credits to purchase
	currentBalance, err := s.CreditService.GetCustomerCredits(ctx, customerID)
	if err != nil {
		log.Printf("Failed to get customer credit balance: %v", err)
		return "", err
	}

	if currentBalance > 0 {
		return "", errLib.New("Cannot purchase credit package while you have remaining credits. Please use your existing credits first.", http.StatusBadRequest)
	}

	// Get package details
	pkg, err := s.CreditPackageRepo.GetByID(ctx, packageID)
	if err != nil {
		return "", err
	}

	// Get existing Stripe customer ID to reuse (industry standard: one user = one Stripe customer)
	existingCustomerID := s.getExistingStripeCustomerID(ctx, customerID)

	// Create Stripe checkout session for one-time payment
	// Pass packageID in metadata so webhook can identify the purchase
	packageIDStr := packageID.String()
	checkoutURL, err := stripe.CreateOneTimePayment(ctx, pkg.StripePriceID, 1, &packageIDStr, nil, successURL, cancelURL, existingCustomerID)
	if err != nil {
		log.Printf("Failed to create Stripe checkout session: %v", err)
		return "", err
	}

	return checkoutURL, nil
}

// GetCustomerActivePackage returns the customer's currently active credit package
func (s *CreditPackageService) GetCustomerActivePackage(ctx context.Context, customerID uuid.UUID) (*dto.CustomerActiveCreditPackageResponse, *errLib.CommonError) {
	return s.CreditPackageRepo.GetCustomerActivePackage(ctx, customerID)
}

// Admin CRUD operations

func (s *CreditPackageService) CreatePackage(ctx context.Context, req dto.CreateCreditPackageRequest) (*dto.CreditPackageResponse, *errLib.CommonError) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// If no StripePriceID provided, create product/price in Stripe
	if req.StripePriceID == "" {
		// Set defaults
		currency := req.Currency
		if currency == "" {
			currency = "cad"
		}

		// Create Stripe Product + one-time Price (credit packages are one-time purchases)
		priceID, productID, err := s.ProductService.CreateProductWithOneTimePrice(
			req.Name,
			req.Description,
			*req.UnitAmount,
			currency,
		)
		if err != nil {
			return nil, err
		}

		req.StripePriceID = priceID
		log.Printf("[STRIPE] Created Stripe product %s with price %s for credit package '%s'", productID, priceID, req.Name)
	}

	return s.CreditPackageRepo.Create(ctx, req)
}

func (s *CreditPackageService) UpdatePackage(ctx context.Context, id uuid.UUID, req dto.UpdateCreditPackageRequest) (*dto.CreditPackageResponse, *errLib.CommonError) {
	return s.CreditPackageRepo.Update(ctx, id, req)
}

func (s *CreditPackageService) DeletePackage(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	// First, get the package to retrieve Stripe price ID before deleting
	pkg, err := s.CreditPackageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if any customers currently have this package active with remaining credits
	var activeWithCreditsCount int
	checkActiveQuery := `
		SELECT COUNT(*) FROM users.customer_credit_packages ccp
		JOIN users.customer_credits cc ON cc.customer_id = ccp.customer_id
		WHERE ccp.credit_package_id = $1 AND cc.balance > 0
	`
	if dbErr := s.DB.QueryRowContext(ctx, checkActiveQuery, id).Scan(&activeWithCreditsCount); dbErr != nil {
		log.Printf("Failed to check active credit packages for package %s: %v", id, dbErr)
		return errLib.New("Failed to check active credit packages", http.StatusInternalServerError)
	}

	if activeWithCreditsCount > 0 {
		return errLib.New("Cannot delete package: customers have remaining credits from this package", http.StatusBadRequest)
	}

	// Delete customer_credit_packages entries for this package (customers with 0 credits)
	deleteCustomerPackagesQuery := `DELETE FROM users.customer_credit_packages WHERE credit_package_id = $1`
	if _, dbErr := s.DB.ExecContext(ctx, deleteCustomerPackagesQuery, id); dbErr != nil {
		log.Printf("Failed to delete customer credit packages for package %s: %v", id, dbErr)
		return errLib.New("Failed to delete customer credit packages", http.StatusInternalServerError)
	}

	// Deactivate Stripe product and price (don't fail if Stripe fails)
	if pkg.StripePriceID != "" {
		s.ProductService.DeactivatePrice(pkg.StripePriceID)
		s.ProductService.DeactivateProductFromPrice(pkg.StripePriceID)
		log.Printf("[STRIPE] Deactivated Stripe price %s for credit package '%s'", pkg.StripePriceID, pkg.Name)
	}

	return s.CreditPackageRepo.Delete(ctx, id)
}
