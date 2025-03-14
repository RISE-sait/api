package event_staff

import (
	db "api/internal/domains/event_staff/persistence/sqlc/generated"
	values "api/internal/domains/event_staff/values"
	staffValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"

	"context"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type EventStaffsRepository struct {
	Queries *db.Queries
}

var _ EventStaffsRepositoryInterface = (*EventStaffsRepository)(nil)

func NewEventStaffsRepository(dbQueries *db.Queries) *EventStaffsRepository {
	return &EventStaffsRepository{
		Queries: dbQueries,
	}
}

func (r *EventStaffsRepository) AssignStaffToEvent(c context.Context, input values.EventStaff) *errLib.CommonError {

	dbParams := db.AssignStaffToEventParams{
		EventID: input.EventID,
		StaffID: input.StaffID,
	}

	_, err := r.Queries.AssignStaffToEvent(c, dbParams)

	if err != nil {
		log.Printf("Failed to assign staff to event: %+v. Error: %v", input.EventID, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil
}

func (r *EventStaffsRepository) GetStaffsAssignedToEvent(ctx context.Context, eventId uuid.UUID) ([]staffValues.ReadValues, *errLib.CommonError) {

	dbStaffs, err := r.Queries.GetStaffsAssignedToEvent(ctx, eventId)

	if err != nil {
		log.Println("Failed to get staffs: ", err.Error())
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	staffs := make([]staffValues.ReadValues, len(dbStaffs))
	for i, dbStaff := range dbStaffs {

		staffs[i] = staffValues.ReadValues{
			ID:        dbStaff.ID,
			HubspotID: dbStaff.HubspotID.String,
			IsActive:  dbStaff.IsActive,
			CreatedAt: dbStaff.CreatedAt,
			UpdatedAt: dbStaff.UpdatedAt,
			RoleName:  dbStaff.RoleName,
		}
	}

	return staffs, nil
}

func (r *EventStaffsRepository) UnassignedStaffFromEvent(c context.Context, values values.EventStaff) *errLib.CommonError {

	params := db.UnassignStaffFromEventParams{
		EventID: values.EventID,
		StaffID: values.StaffID,
	}

	_, err := r.Queries.UnassignStaffFromEvent(c, params)

	if err != nil {
		log.Printf("Failed to unassign staff: %+v. Error: %v", values.StaffID, err.Error())
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return nil

}
