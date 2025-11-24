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
		PhotoURL    *string   `json:"photo_url,omitempty"`
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

	CourtInfo struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	EventResponseDto struct {
		ID                        uuid.UUID         `json:"id"`
		Program                   ProgramInfo       `json:"program"`
		Location                  LocationInfo      `json:"location"`
		Capacity                  int32             `json:"capacity"`
		CreatedBy                 PersonResponseDto `json:"created_by"`
		UpdatedBy                 PersonResponseDto `json:"updated_by"`
		Team                      *TeamInfo         `json:"team,omitempty"`
		Court                     *CourtInfo        `json:"court,omitempty"`
		RequiredMembershipPlanIDs []uuid.UUID       `json:"required_membership_plan_ids,omitempty"`
		PriceID                   *string           `json:"price_id,omitempty"`
		CreditCost                *int32            `json:"credit_cost,omitempty"`
		DateResponseDto
		*Participants
	}
)

func NewEventResponseDto(event values.ReadEventValues, includePeople bool, includeContactInfo bool) EventResponseDto {
	response := EventResponseDto{
		ID:                        event.ID,
		Capacity:                  event.Capacity,
		CreatedBy:                 PersonResponseDto(event.CreatedBy),
		UpdatedBy:                 PersonResponseDto(event.UpdatedBy),
		RequiredMembershipPlanIDs: event.RequiredMembershipPlanIDs,
		PriceID:                   event.PriceID,
		CreditCost:                event.CreditCost,
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
			PhotoURL:    event.Program.PhotoURL,
		},
	}

	if event.Team != nil {
		response.Team = &TeamInfo{
			ID:   event.Team.ID,
			Name: event.Team.Name,
		}
	}

	if event.Court != nil {
		response.Court = &CourtInfo{
			ID:   event.Court.ID,
			Name: event.Court.Name,
		}
	}

	if includePeople {
		response.Participants = &Participants{
			Customers: mapCustomers(event.Customers, includeContactInfo),
			Staff:     mapStaffs(event.Staffs, includeContactInfo),
		}
	}

	return response
}

func mapCustomers(customers []values.Customer, includeContactInfo bool) []CustomerResponseDto {
	result := make([]CustomerResponseDto, 0, len(customers))
	for _, c := range customers {
		dto := CustomerResponseDto{
			PersonResponseDto: PersonResponseDto{
				ID:        c.ID,
				FirstName: c.FirstName,
				LastName:  c.LastName,
			},
			HasCancelledEnrollment: c.HasCancelledEnrollment,
		}

		// Only include contact info if requested (admin/receptionist access)
		if includeContactInfo {
			dto.Email = c.Email
			dto.Phone = c.Phone
			dto.Gender = c.Gender
		}

		result = append(result, dto)
	}
	return result
}

func mapStaffs(staffs []values.Staff, includeContactInfo bool) []StaffResponseDto {
	result := make([]StaffResponseDto, 0, len(staffs))
	for _, staff := range staffs {
		dto := StaffResponseDto{
			PersonResponseDto: PersonResponseDto{
				ID:        staff.ID,
				FirstName: staff.FirstName,
				LastName:  staff.LastName,
			},
			RoleName: staff.RoleName,
		}

		// Only include contact info if requested (admin/receptionist access)
		if includeContactInfo {
			dto.Email = staff.Email
			dto.Phone = staff.Phone
			dto.Gender = staff.Gender
		}

		result = append(result, dto)
	}
	return result
}
