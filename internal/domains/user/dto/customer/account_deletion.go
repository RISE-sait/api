package customer

// AccountDeletionRequest represents the request to delete a customer account
type AccountDeletionRequest struct {
	ConfirmDeletion bool `json:"confirm_deletion" binding:"required" example:"true"`
}