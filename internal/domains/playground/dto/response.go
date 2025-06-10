package playground

import (
	values "api/internal/domains/playground/values"
	"time"
)
// ResponseDto represents the data transfer object for a session response.
type ResponseDto struct {
	ID         string `json:"id"`
	SystemID   string `json:"system_id"`
	SystemName string `json:"system_name"`
	CustomerID string `json:"customer_id"`
	CustomerFirstName string `json:"customer_first_name"`
	CustomerLastName  string `json:"customer_last_name"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
// NewResponse creates a new ResponseDto from a Session value.
func NewResponse(session values.Session) ResponseDto {
	return ResponseDto{
		ID:         session.ID.String(),
		SystemID:   session.SystemID.String(),
		SystemName: session.SystemName,
		CustomerID: session.CustomerID.String(),
		CustomerFirstName: session.CustomerFirstName,
		CustomerLastName:  session.CustomerLastName,
		StartTime:  session.StartTime.Format(time.RFC3339),
		EndTime:    session.EndTime.Format(time.RFC3339),
		CreatedAt:  session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  session.UpdatedAt.Format(time.RFC3339),
	}
}
