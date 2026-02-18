package family

import (
	"time"

	"github.com/google/uuid"
)

// RequestLinkResponse is returned after initiating a link request
type RequestLinkResponse struct {
	Message           string    `json:"message"`
	RequiresOldParent bool      `json:"requires_old_parent"`
	ExpiresAt         time.Time `json:"expires_at"`
}

// ConfirmLinkResponse is returned after confirming a link request
type ConfirmLinkResponse struct {
	Message   string `json:"message"`
	Status    string `json:"status"` // "complete", "pending_old_parent", "pending_new_parent"
	ChildID   string `json:"child_id,omitempty"`
	ChildName string `json:"child_name,omitempty"`
}

// ChildResponse represents a child in the parent's children list
type ChildResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email,omitempty"`
	LinkedAt  time.Time `json:"linked_at"`
}

// PendingRequestResponse represents a pending link request
type PendingRequestResponse struct {
	ID                   uuid.UUID  `json:"id"`
	ChildID              uuid.UUID  `json:"child_id"`
	ChildName            string     `json:"child_name"`
	ChildEmail           string     `json:"child_email,omitempty"`
	NewParentID          uuid.UUID  `json:"new_parent_id"`
	NewParentName        string     `json:"new_parent_name"`
	NewParentEmail       string     `json:"new_parent_email,omitempty"`
	OldParentID          *uuid.UUID `json:"old_parent_id,omitempty"`
	OldParentName        string     `json:"old_parent_name,omitempty"`
	OldParentEmail       string     `json:"old_parent_email,omitempty"`
	InitiatedBy          string     `json:"initiated_by"`
	CounterpartyVerified bool       `json:"counterparty_verified"`
	OldParentVerified    bool       `json:"old_parent_verified"`
	RequiresOldParent    bool       `json:"requires_old_parent"`
	ExpiresAt            time.Time  `json:"expires_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UserRole             string     `json:"user_role"`            // "child", "new_parent", or "old_parent"
	AwaitingUserAction   bool       `json:"awaiting_user_action"`
}
