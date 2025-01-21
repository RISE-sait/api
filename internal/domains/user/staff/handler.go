package staffs

// Step 6: Handle optional staff insertion if provided.
// var staffInfo *auth.StaffInfo
// role := "Athlete" // Default role.

// if createUserOptionalInfoBody.StaffInfo != nil {
// 	// Map staff info request to DB parameters.
// 	var dbCreateStaffParams db.CreateStaffParams
// 	if err := copier.Copy(&dbCreateStaffParams, createUserOptionalInfoBody.StaffInfo); err != nil {
// 		utils.RespondWithError(w, utils.CreateHTTPError("Error copying staff data", http.StatusInternalServerError))
// 		return
// 	}
// 	dbCreateStaffParams.Email = createUserOptionalInfoBody.Email

// 	// Insert staff info.
// 	if err := c.StaffRepository.CreateStaff(r.Context(), &dbCreateStaffParams); err != nil {
// 		utils.RespondWithError(w, err)
// 		return
// 	}

// 	// Populate staff info for JWT.
// 	staffInfo = &auth.StaffInfo{
// 		Role:     string(dbCreateStaffParams.Role),
// 		IsActive: dbCreateStaffParams.IsActive,
// 	}
// 	role = string(dbCreateStaffParams.Role)
// }
