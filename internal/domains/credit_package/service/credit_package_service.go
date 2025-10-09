package service

import (
	"context"
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
}

func NewCreditPackageService(container *di.Container) *CreditPackageService {
	return &CreditPackageService{
		CreditPackageRepo: repo.NewCreditPackageRepository(container),
		CreditService:     userServices.NewCustomerCreditService(container),
		StripeService:     stripe.NewPriceService(),
	}
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
func (s *CreditPackageService) CheckoutCreditPackage(ctx context.Context, packageID uuid.UUID, successURL string) (string, *errLib.CommonError) {
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

	// Create Stripe checkout session for one-time payment
	// Pass packageID in metadata so webhook can identify the purchase
	packageIDStr := packageID.String()
	checkoutURL, err := stripe.CreateOneTimePayment(ctx, pkg.StripePriceID, 1, &packageIDStr, successURL)
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
	return s.CreditPackageRepo.Create(ctx, req)
}

func (s *CreditPackageService) UpdatePackage(ctx context.Context, id uuid.UUID, req dto.UpdateCreditPackageRequest) (*dto.CreditPackageResponse, *errLib.CommonError) {
	return s.CreditPackageRepo.Update(ctx, id, req)
}

func (s *CreditPackageService) DeletePackage(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	return s.CreditPackageRepo.Delete(ctx, id)
}
