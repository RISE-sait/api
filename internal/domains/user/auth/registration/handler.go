package controllers

import (
	users "api/internal/domains/user"
	"api/internal/domains/user/auth"
	"api/internal/domains/user/customers/services/hubspot"
	user_optional_info "api/internal/domains/user/optionalInfo"
	staffs "api/internal/domains/user/staff"
	"api/internal/utils"
	"api/internal/utils/validators"
	"database/sql"
	"net/http"
)

type AccountRegistrationController struct {
	UsersRespository           *users.Repository
	UserOptionalInfoRepository *user_optional_info.Repository
	StaffRepository            *staffs.Repository
	DB                         *sql.DB
	HubSpotService             *hubspot.HubSpotService
}

func (c *AccountRegistrationController) CreateTraditionalAccount(w http.ResponseWriter, r *http.Request) {

	var createUserOptionalInfoBody user_optional_info.CreateUserOptionalInfoRequest

	// Step 1: Decode and validate the request body.
	if err := validators.DecodeAndValidateRequestBody(r.Body, &createUserOptionalInfoBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	tx, err := c.DB.BeginTx(r.Context(), &sql.TxOptions{})

	if err != nil {
		utils.RespondWithError(w, utils.CreateHTTPError("Failed to begin transaction", http.StatusInternalServerError))
		return
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			utils.RespondWithError(w, utils.CreateHTTPError("Transaction rolled back", http.StatusInternalServerError))
		}
	}()

	// Step 2: Prepare the database request for optional info (email/password).
	dbCreateUserOptionalInfoRequestBody := createUserOptionalInfoBody.ToDBParams()

	// Step 4: Insert into users table.
	if err := c.UsersRespository.CreateUserTx(r.Context(), tx, dbCreateUserOptionalInfoRequestBody.Email); err != nil {
		tx.Rollback()
		utils.RespondWithError(w, err)
		return
	}

	// Step 5: Insert into optional info (email/password).
	if err := c.UserOptionalInfoRepository.CreateUserOptionalInfoTx(r.Context(), tx, *dbCreateUserOptionalInfoRequestBody); err != nil {
		tx.Rollback()
		utils.RespondWithError(w, err)
		return
	}

	// Step 5: Commit the core transaction.
	if err := tx.Commit(); err != nil {
		utils.RespondWithError(w, utils.CreateHTTPError("Failed to commit transaction", http.StatusInternalServerError))
		return
	}

	// Step 7: Create JWT claims.
	userInfo := auth.UserInfo{
		// Name:  createUserOptionalInfoBody.Name,
		Email: createUserOptionalInfoBody.Email,
	}

	// Step 8: Sign JWT.
	signedToken, signError := auth.SignJWT(userInfo)
	if signError != nil {
		utils.RespondWithError(w, signError)
		return
	}

	// Step 9: Set Authorization header and respond.
	w.Header().Set("Authorization", "Bearer "+signedToken)
	utils.RespondWithSuccess(w, nil, http.StatusCreated)
}
