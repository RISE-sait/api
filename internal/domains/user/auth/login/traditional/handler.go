package auth

import (
	"api/internal/domains/user/auth"
	user_optional_info "api/internal/domains/user/optionalInfo"
	staffs "api/internal/domains/user/staff"
	"api/internal/utils"
	"api/internal/utils/validators"
	"net/http"
	"strings"
)

type TraditionalLoginHandler struct {
	UserOptionalInfoRepository *user_optional_info.Repository
	StaffRepository            *staffs.Repository
}

func NewTraditionalLoginController(traditionalAccsRepo *user_optional_info.Repository, staffRepository *staffs.Repository) *TraditionalLoginHandler {
	return &TraditionalLoginHandler{
		UserOptionalInfoRepository: traditionalAccsRepo,
		StaffRepository:            staffRepository,
	}
}

func (c *TraditionalLoginHandler) GetUser(w http.ResponseWriter, r *http.Request) {

	var targetBody user_optional_info.GetUserOptionalInfoRequest

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

	staffInfo := auth.StaffInfo{
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
	userInfo := auth.UserInfo{
		Email:     params.Email,
		Name:      name,
		StaffInfo: &staffInfo,
	}

	// Sign JWT
	signedToken, signError := auth.SignJWT(userInfo)
	if signError != nil {
		utils.RespondWithError(w, signError)
		return
	}

	// Set Authorization header and respond
	w.Header().Set("Authorization", "Bearer "+signedToken)
	utils.RespondWithSuccess(w, nil, http.StatusOK)
}
