package event

import (
	values "api/internal/domains/event/values"

	"github.com/google/uuid"
)

//goland:noinspection GoNameStartsWithPackageName
type (
	DateResponseDto struct {
		StartAt string `json:"start_at"`
		EndAt   string `json:"end_at"`
	}

	PersonResponseDto struct {
		ID        uuid.UUID `json:"id"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
	}

	CustomerResponseDto struct {
		PersonResponseDto
		Email                  *string `json:"email,omitempty"`
		Phone                  *string `json:"phone,omitempty"`
		Gender                 *string `json:"gender,omitempty"`
		HasCancelledEnrollment bool    `json:"has_cancelled_enrollment"`
	}

	StaffResponseDto struct {
		PersonResponseDto
		Email    string  `json:"email,omitempty"`
		Phone    string  `json:"phone,omitempty"`
		Gender   *string `json:"gender,omitempty"`
		RoleName string  `json:"role_name"`
	}

	Participants struct {
		Customers []CustomerResponseDto `json:"customers"`
		Staff     []StaffResponseDto    `json:"staff"`
	}

	ProgramInfo struct {
		ID          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		Description *string   `json:"description,omitempty"`
	}

	LocationInfo struct {
		ID      uuid.UUID `json:"id"`
		Name    string    `json:"name"`
		Address string    `json:"address"`
	}

	TeamInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	EventResponseDto struct {
		ID        uuid.UUID         `json:"id"`
		Program   ProgramInfo       `json:"program"`
		Location  LocationInfo      `json:"location"`
		Capacity  int32             `json:"capacity"`
		CreatedBy PersonResponseDto `json:"created_by"`
		UpdatedBy PersonResponseDto `json:"updated_by"`
		Team      *TeamInfo         `json:"team,omitempty"`
		DateResponseDto
		*Participants
	}
)

func NewEventResponseDto(event values.ReadEventValues, includePeople bool) EventResponseDto {
	response := EventResponseDto{
		ID:        event.ID,
		Capacity:  event.Capacity,
		CreatedBy: PersonResponseDto(event.CreatedBy),
		UpdatedBy: PersonResponseDto(event.UpdatedBy),
		Location: LocationInfo{
			ID:      event.Location.ID,
			Name:    event.Location.Name,
			Address: event.Location.Address,
		},
		DateResponseDto: DateResponseDto{
			StartAt: event.StartAt.String(),
			EndAt:   event.EndAt.String(),
		},
		Program: ProgramInfo{
			ID:          event.Program.ID,
			Name:        event.Program.Name,
			Type:        event.Program.Type,
			Description: &event.Program.Description,
		},
	}

	if event.Team != nil {
		response.Team = &TeamInfo{
			ID:   event.Team.ID,
			Name: event.Team.Name,
		}
	}

	if includePeople {
		response.Participants = &Participants{
			Customers: mapCustomers(event.Customers),
			Staff:     mapStaffs(event.Staffs),
		}
	}

	return response
}

func mapCustomers(customers []values.Customer) []CustomerResponseDto {
	result := make([]CustomerResponseDto, 0, len(customers))
	for _, c := range customers {
		result = append(result, CustomerResponseDto{
			PersonResponseDto: PersonResponseDto{
				ID:        c.ID,
				FirstName: c.FirstName,
				LastName:  c.LastName,
			},
			Email:                  c.Email,
			Phone:                  c.Phone,
			Gender:                 c.Gender,
			HasCancelledEnrollment: c.HasCancelledEnrollment,
		})
	}
	return result
}

func mapStaffs(staffs []values.Staff) []StaffResponseDto {
	result := make([]StaffResponseDto, 0, len(staffs))
	for _, staff := range staffs {
		result = append(result, StaffResponseDto{
			PersonResponseDto: PersonResponseDto{
				ID:        staff.ID,
				FirstName: staff.FirstName,
				LastName:  staff.LastName,
			},
			Email:    staff.Email,
			Phone:    staff.Phone,
			Gender:   staff.Gender,
			RoleName: staff.RoleName,
		})
	}
	return result
}
