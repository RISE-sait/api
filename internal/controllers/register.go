package controllers

// import (
// 	userOptionalInfo "api/internal/dtos/user/optionalInfo"
// 	"api/internal/repositories"
// 	"api/internal/types/auth"
// 	"api/internal/utils"
// 	"api/internal/utils/validators"
// 	db "api/sqlc"
// 	"net/http"

// 	"github.com/jinzhu/copier"
// )

// type AccountRegistrationController struct {
// 	UsersRespository 		 *repositories.UsersRepository
// 	UserOptionalInfoRepository *repositories.UserOptionalInfoRepository
// 	StaffRepository            *repositories.StaffRepository
// }

// func (c *AccountRegistrationController) CreateTraditionalAccount(w http.ResponseWriter, r *http.Request) {
// 	var targetBody userOptionalInfo.CreateUserOptionalInfoRequest

// 	// Step 1: Decode and validate the request body.
// 	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	// Step 2: Prepare the database request for optional info (email/password).
// 	dbCreateUserOptionalInfoRequestBody := targetBody.ToDBParams()

// 	// Step 3: Begin a transaction for core registration (user + optional info).
// 	tx, err := c.UserOptionalInfoRepository.BeginTransaction(r.Context())
// 	if err != nil {
// 		httpErr := utils.CreateHTTPError("Failed to begin transaction", http.StatusInternalServerError)
// 		utils.RespondWithError(w, httpErr)
// 		return
// 	}

// 	defer func() {
// 		if p := recover(); p != nil {
// 			tx.Rollback()
// 			utils.RespondWithError(w, utils.CreateHTTPError("Transaction rolled back", http.StatusInternalServerError))
// 		}
// 	}()

// 	// Step 4: Insert into optional info (email/password).
// 	if err := c.UserOptionalInfoRepository.CreateUserOptionalInfoTx(tx, r.Context(), dbCreateUserOptionalInfoRequestBody); err != nil {
// 		tx.Rollback()
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	// Step 5: Commit the core transaction.
// 	if err := tx.Commit(); err != nil {
// 		utils.RespondWithError(w, utils.CreateHTTPError("Failed to commit transaction", http.StatusInternalServerError))
// 		return
// 	}

// 	// Step 6: Handle optional staff insertion if provided.
// 	var staffInfo *auth.StaffInfo
// 	role := "Athlete" // Default role.

// 	if targetBody.StaffInfo != nil {
// 		// Map staff info request to DB parameters.
// 		var dbCreateStaffParams db.CreateStaffParams
// 		if err := copier.Copy(&dbCreateStaffParams, targetBody.StaffInfo); err != nil {
// 			utils.RespondWithError(w, utils.CreateHTTPError("Error copying staff data", http.StatusInternalServerError))
// 			return
// 		}
// 		dbCreateStaffParams.Email = targetBody.Email

// 		// Insert staff info.
// 		if err := c.StaffRepository.CreateStaff(r.Context(), &dbCreateStaffParams); err != nil {
// 			utils.RespondWithError(w, err)
// 			return
// 		}

// 		// Populate staff info for JWT.
// 		staffInfo = &auth.StaffInfo{
// 			Role:     string(dbCreateStaffParams.Role),
// 			IsActive: dbCreateStaffParams.IsActive,
// 		}
// 		role = string(dbCreateStaffParams.Role)
// 	}

// 	// Step 7: Create JWT claims.
// 	userInfo := auth.UserInfo{
// 		Name:      targetBody.Name,
// 		Email:     targetBody.Email,
// 		StaffInfo: staffInfo,
// 	}

// 	// Step 8: Sign JWT.
// 	signedToken, signError := utils.SignJWT(userInfo)
// 	if signError != nil {
// 		utils.RespondWithError(w, signError)
// 		return
// 	}

// 	// Step 9: Set Authorization header and respond.
// 	w.Header().Set("Authorization", "Bearer "+signedToken)
// 	utils.RespondWithSuccess(w, nil, http.StatusCreated)
// }
