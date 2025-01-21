package traditional_login

import (
	userOptionalInfo "api/internal/dtos/user/optionalInfo"
	"api/internal/repositories"
	"api/internal/types"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"
	"strings"
)

type TraditionalLoginController struct {
	UserOptionalInfoRepository *repositories.UserOptionalInfoRepository
	StaffRepository            *repositories.StaffRepository
}

func NewTraditionalLoginController(traditionalAccsRepo *repositories.UserOptionalInfoRepository, staffRepository *repositories.StaffRepository) *TraditionalLoginController {
	return &TraditionalLoginController{
		UserOptionalInfoRepository: traditionalAccsRepo,
		StaffRepository:            staffRepository,
	}
}

func (c *TraditionalLoginController) GetUser(w http.ResponseWriter, r *http.Request) {

	var targetBody userOptionalInfo.GetUserOptionalInfoRequest

	if err := validators.DecodeAndValidateRequestBody(r.Body, &targetBody); err != nil {
		utils.RespondWithError(w, err)
		return
	}

	params := targetBody.ToDBParams()

	if isUserExist := c.UserOptionalInfoRepository.IsUserOptionalInfoExist(r.Context(), params); !isUserExist {

		httpErr := utils.CreateHTTPError("No user with the associated credentials found", http.StatusNotFound)

		utils.RespondWithError(w, httpErr)
		return
	}

	staffInfo := types.StaffInfo{
		Role:     "Athlete",
		IsActive: false,
	}

	name := strings.Split(params.Email, "@")[0]

	staff, err := c.StaffRepository.GetStaffByEmail(r.Context(), params.Email)

	if err == nil {

		staffInfo.Role = string(staff.Role)

		staffInfo.IsActive = staff.IsActive

		if staff.Name.Valid {
			name = staff.Name.String
		}
	}

	// Create user info for JWT
	userInfo := types.UserInfo{
		Email:     params.Email,
		Name:      name,
		StaffInfo: &staffInfo,
	}

	// Sign JWT
	signedToken, signError := utils.SignJWT(userInfo)
	if signError != nil {
		utils.RespondWithError(w, signError)
		return
	}

	// Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	utils.RespondWithSuccess(w, nil, http.StatusOK)
}
