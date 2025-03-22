package haircut

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/haircut/persistence/sqlc/generated"
	values "api/internal/domains/haircut/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type ServiceRepository struct {
	Queries *db.Queries
}

func NewHaircutServiceRepository(dbQueries *db.Queries) *ServiceRepository {
	return &ServiceRepository{
		Queries: dbQueries,
	}
}

func (r *ServiceRepository) CreateHaircutService(c context.Context, details values.CreateHaircutServiceValues) *errLib.CommonError {

	dbParams := db.CreateHaircutServiceParams{
		Name:          details.Name,
		Description:   sql.NullString{},
		Price:         details.Price,
		DurationInMin: details.DurationInMin,
	}

	affectedRows, err := r.Queries.CreateHaircutService(c, dbParams)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("A haircut service with the same name already exists", http.StatusConflict)
		}

		log.Printf("Failed to create service: %+v. Error: %v", details, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Failed to create service. Unknown reason. Please try again.", http.StatusInternalServerError)
	}

	return nil
}

func (r *Repository) GetHaircutServices(ctx context.Context) ([]values.ReadHaircutServicesValues, *errLib.CommonError) {

	dbServices, err := r.Queries.GetHaircutServices(ctx)

	if err != nil {
		log.Println("Failed to get services: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	services := make([]values.ReadHaircutServicesValues, len(dbServices))
	for i, dbService := range dbServices {

		service := values.ReadHaircutServicesValues{
			ID: dbService.ID,
			ServiceValuesBase: values.ServiceValuesBase{
				Name:          dbService.Name,
				Price:         dbService.Price,
				DurationInMin: dbService.DurationInMin,
			},
		}

		if dbService.Description.Valid {
			service.Description = &dbService.Description.String
		}

		services[i] = service

	}

	return services, nil
}

func (r *Repository) UpdateHaircutService(c context.Context, details values.UpdateHaircutServicesValues) *errLib.CommonError {
	params := db.UpdateHaircutServiceParams{
		Name:          details.Name,
		DurationInMin: details.DurationInMin,
		Price:         details.Price,
		ID:            details.ID,
	}

	if details.Description != nil {
		params.Description = sql.NullString{
			String: *details.Description,
			Valid:  true,
		}
	}

	affectedRows, err := r.Queries.UpdateHaircutService(c, params)

	if err != nil {
		log.Printf("Failed to update service: %+v. Error: %v", params, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Service not found", http.StatusNotFound)
	}

	return nil

}

func (r *Repository) DeleteHaircutService(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteHaircutService(c, id)

	if err != nil {
		log.Printf("Failed to delete service with Id: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Service not found", http.StatusNotFound)
	}

	return nil
}
