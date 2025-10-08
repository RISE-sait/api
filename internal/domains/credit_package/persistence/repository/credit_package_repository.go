package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"api/internal/di"
	dto "api/internal/domains/credit_package/dto"
	db "api/internal/domains/user/persistence/sqlc/generated"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
)

type CreditPackageRepository struct {
	Queries *db.Queries
}

func NewCreditPackageRepository(container *di.Container) *CreditPackageRepository {
	return &CreditPackageRepository{
		Queries: container.Queries.UserDb,
	}
}

func (r *CreditPackageRepository) GetByID(ctx context.Context, id uuid.UUID) (*dto.CreditPackageResponse, *errLib.CommonError) {
	pkg, err := r.Queries.GetCreditPackageByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Credit package not found", http.StatusNotFound)
		}
		log.Printf("Failed to get credit package: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return r.mapToResponse(&pkg), nil
}

func (r *CreditPackageRepository) GetByStripePriceID(ctx context.Context, stripePriceID string) (*dto.CreditPackageResponse, *errLib.CommonError) {
	pkg, err := r.Queries.GetCreditPackageByStripePriceID(ctx, stripePriceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Credit package not found", http.StatusNotFound)
		}
		log.Printf("Failed to get credit package by Stripe price ID: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return r.mapToResponse(&pkg), nil
}

func (r *CreditPackageRepository) GetAll(ctx context.Context) ([]dto.CreditPackageResponse, *errLib.CommonError) {
	packages, err := r.Queries.GetAllCreditPackages(ctx)
	if err != nil {
		log.Printf("Failed to get all credit packages: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	response := make([]dto.CreditPackageResponse, len(packages))
	for i, pkg := range packages {
		response[i] = *r.mapToResponse(&pkg)
	}

	return response, nil
}

func (r *CreditPackageRepository) Create(ctx context.Context, req dto.CreateCreditPackageRequest) (*dto.CreditPackageResponse, *errLib.CommonError) {
	pkg, err := r.Queries.CreateCreditPackage(ctx, db.CreateCreditPackageParams{
		Name:              req.Name,
		Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
		StripePriceID:     req.StripePriceID,
		CreditAllocation:  req.CreditAllocation,
		WeeklyCreditLimit: req.WeeklyCreditLimit,
	})
	if err != nil {
		log.Printf("Failed to create credit package: %v", err)
		return nil, errLib.New("Failed to create credit package", http.StatusInternalServerError)
	}

	return r.mapToResponse(&pkg), nil
}

func (r *CreditPackageRepository) Update(ctx context.Context, id uuid.UUID, req dto.UpdateCreditPackageRequest) (*dto.CreditPackageResponse, *errLib.CommonError) {
	pkg, err := r.Queries.UpdateCreditPackage(ctx, db.UpdateCreditPackageParams{
		ID:                id,
		Name:              req.Name,
		Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
		StripePriceID:     req.StripePriceID,
		CreditAllocation:  req.CreditAllocation,
		WeeklyCreditLimit: req.WeeklyCreditLimit,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("Credit package not found", http.StatusNotFound)
		}
		log.Printf("Failed to update credit package: %v", err)
		return nil, errLib.New("Failed to update credit package", http.StatusInternalServerError)
	}

	return r.mapToResponse(&pkg), nil
}

func (r *CreditPackageRepository) Delete(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	err := r.Queries.DeleteCreditPackage(ctx, id)
	if err != nil {
		log.Printf("Failed to delete credit package: %v", err)
		return errLib.New("Failed to delete credit package", http.StatusInternalServerError)
	}

	return nil
}

func (r *CreditPackageRepository) GetCustomerActivePackage(ctx context.Context, customerID uuid.UUID) (*dto.CustomerActiveCreditPackageResponse, *errLib.CommonError) {
	result, err := r.Queries.GetCustomerActiveCreditPackage(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("No active credit package found", http.StatusNotFound)
		}
		log.Printf("Failed to get customer active credit package: %v", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return &dto.CustomerActiveCreditPackageResponse{
		CustomerID:        result.CustomerID,
		CreditPackageID:   result.CreditPackageID,
		PackageName:       result.PackageName,
		CreditAllocation:  result.CreditAllocation,
		WeeklyCreditLimit: result.WeeklyCreditLimit,
		PurchasedAt:       result.PurchasedAt,
	}, nil
}

func (r *CreditPackageRepository) SetCustomerActivePackage(ctx context.Context, customerID uuid.UUID, packageID uuid.UUID, weeklyLimit int32) *errLib.CommonError {
	err := r.Queries.SetCustomerActiveCreditPackage(ctx, db.SetCustomerActiveCreditPackageParams{
		CustomerID:        customerID,
		CreditPackageID:   packageID,
		WeeklyCreditLimit: weeklyLimit,
	})
	if err != nil {
		log.Printf("Failed to set customer active credit package: %v", err)
		return errLib.New("Failed to set active credit package", http.StatusInternalServerError)
	}

	return nil
}

func (r *CreditPackageRepository) mapToResponse(pkg *db.UsersCreditPackage) *dto.CreditPackageResponse {
	var description *string
	if pkg.Description.Valid {
		description = &pkg.Description.String
	}

	return &dto.CreditPackageResponse{
		ID:                pkg.ID,
		Name:              pkg.Name,
		Description:       description,
		StripePriceID:     pkg.StripePriceID,
		CreditAllocation:  pkg.CreditAllocation,
		WeeklyCreditLimit: pkg.WeeklyCreditLimit,
		CreatedAt:         pkg.CreatedAt,
		UpdatedAt:         pkg.UpdatedAt,
	}
}
