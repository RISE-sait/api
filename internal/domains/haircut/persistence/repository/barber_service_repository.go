package haircut

import (
	databaseErrors "api/internal/constants"
	db "api/internal/domains/haircut/persistence/sqlc/generated"
	values "api/internal/domains/haircut/values"
	errLib "api/internal/libs/errors"
	"context"
	"errors"
	"github.com/lib/pq"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type BarberServiceRepository struct {
	Queries *db.Queries
}

func NewBarberServiceRepository(dbQueries *db.Queries) *BarberServiceRepository {
	return &BarberServiceRepository{
		Queries: dbQueries,
	}
}

func (r *BarberServiceRepository) CreateBarberService(c context.Context, barberId, serviceId uuid.UUID) *errLib.CommonError {

	dbParams := db.CreateBarberServiceParams{
		BarberID:  barberId,
		ServiceID: serviceId,
	}

	affectedRows, err := r.Queries.CreateBarberService(c, dbParams)

	if err != nil {

		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == databaseErrors.UniqueViolation {
			return errLib.New("The barber with the haircut service already exists", http.StatusConflict)
		}

		log.Printf("Failed to create service. Barber ID: %v. Service ID: %v. Error: %v", barberId, serviceId, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if affectedRows == 0 {
		return errLib.New("Failed to create service. Unknown reason. Please try again.", http.StatusInternalServerError)
	}

	return nil
}

func (r *BarberServiceRepository) GetBarberServices(ctx context.Context) ([]values.ReadBarberServicesValues, *errLib.CommonError) {

	dbServices, err := r.Queries.GetBarberServices(ctx)

	if err != nil {
		log.Println("Failed to get services: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	services := make([]values.ReadBarberServicesValues, len(dbServices))
	for i, dbService := range dbServices {

		service := values.ReadBarberServicesValues{
			ID:        dbService.ID,
			CreatedAt: dbService.CreatedAt,
			UpdatedAt: dbService.UpdatedAt,
			BarberServiceValuesBase: values.BarberServiceValuesBase{
				ServiceTypeID: dbService.ServiceID,
				BarberID:      dbService.BarberID,
			},
			HaircutName: dbService.HaircutName,
			BarberName:  dbService.BarberName,
		}

		services[i] = service

	}

	return services, nil
}

func (r *BarberServiceRepository) DeleteBarberService(c context.Context, id uuid.UUID) *errLib.CommonError {
	row, err := r.Queries.DeleteBarberService(c, id)

	if err != nil {
		log.Printf("Failed to delete barber service with Id: %s. Error: %s", id, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if row == 0 {
		return errLib.New("Barber service not found", http.StatusNotFound)
	}

	return nil
}
