package user

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	dbIdentity "api/internal/domains/identity/persistence/sqlc/generated"
	values "api/internal/domains/identity/values"
	errLib "api/internal/libs/errors"

	"github.com/google/uuid"
)

func (r *UsersRepository) GetIsActualParentChild(ctx context.Context, childID, parentID uuid.UUID) (bool, *errLib.CommonError) {
	isConnected, err := r.IdentityQueries.GetIsActualParentChild(ctx, dbIdentity.GetIsActualParentChildParams{
		ParentID: uuid.NullUUID{
			UUID:  parentID,
			Valid: true,
		},
		ChildID: childID,
	})
	if err != nil {
		log.Printf("Error verifying parent-child relationship: %v", err.Error())
		return false, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	return isConnected, nil
}

func (r *UsersRepository) GetUserInfo(ctx context.Context, email string, id uuid.UUID) (values.UserReadInfo, *errLib.CommonError) {
	if err := validateUserLookupParams(email, id); err != nil {
		return values.UserReadInfo{}, err
	}

	user, err := r.getUserFromDB(ctx, email, id)
	if err != nil {
		return values.UserReadInfo{}, err
	}

	response := r.buildUserResponse(ctx, user)

	return response, nil
}

func validateUserLookupParams(email string, id uuid.UUID) *errLib.CommonError {
	if email != "" && id != uuid.Nil {
		return errLib.New("specify either email or ID, not both", http.StatusBadRequest)
	}
	if email == "" && id == uuid.Nil {
		return errLib.New("either email or ID must be provided", http.StatusBadRequest)
	}
	return nil
}

func (r *UsersRepository) getUserFromDB(ctx context.Context, email string, id uuid.UUID) (dbIdentity.GetUserByIdOrEmailRow, *errLib.CommonError) {
	var params dbIdentity.GetUserByIdOrEmailParams

	switch {
	case email != "":
		params.Email = sql.NullString{String: email, Valid: true}
	case id != uuid.Nil:
		params.ID = uuid.NullUUID{UUID: id, Valid: true}
	}

	user, err := r.IdentityQueries.GetUserByIdOrEmail(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbIdentity.GetUserByIdOrEmailRow{}, errLib.New("user not found", http.StatusNotFound)
		}
		log.Printf("database error when getting user: %v", err)
		return dbIdentity.GetUserByIdOrEmailRow{}, errLib.New("internal server error when getting user", http.StatusInternalServerError)
	}
	return user, nil
}

func (r *UsersRepository) buildUserResponse(ctx context.Context, user dbIdentity.GetUserByIdOrEmailRow) values.UserReadInfo {
	response := values.UserReadInfo{
		ID:          user.ID,
		DOB:         user.Dob,
		CountryCode: user.CountryAlpha2Code,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
	}

	if user.Email.Valid {
		response.Email = &user.Email.String
	}
	if user.HubspotID.Valid {
		response.HubspotID = &user.HubspotID.String
	}
	if user.Phone.Valid {
		response.Phone = &user.Phone.String
	}
	if user.Gender.Valid {
		response.Gender = &user.Gender.String
	}

	addMembershipInfo(&response, user)
	addAthleteInfo(&response, user)
	setUserRole(ctx, r, &response, user)
	addPhotoURL(ctx, r, &response, user)

	return response
}

func addMembershipInfo(response *values.UserReadInfo, user dbIdentity.GetUserByIdOrEmailRow) {
	if user.MembershipName.Valid && user.MembershipPlanName.Valid {
		info := &values.MembershipReadInfo{
			MembershipName:        user.MembershipName.String,
			MembershipDescription: user.MembershipDescription.String,
			MembershipBenefits:    user.MembershipBenefits.String,
			PlanName:              user.MembershipPlanName.String,
			StartDate:             user.MembershipPlanStartDate.Time,
		}
		if user.MembershipPlanRenewalDate.Valid {
			info.RenewalDate = &user.MembershipPlanRenewalDate.Time
		}
		response.MembershipInfo = info
	}
}

func addAthleteInfo(response *values.UserReadInfo, user dbIdentity.GetUserByIdOrEmailRow) {
	if user.Wins.Valid {
		athleteInfo := &values.AthleteInfo{
			Wins:     user.Wins.Int32,
			Losses:   user.Losses.Int32,
			Points:   user.Points.Int32,
			Steals:   user.Steals.Int32,
			Assists:  user.Assists.Int32,
			Rebounds: user.Rebounds.Int32,
		}

		if user.TeamID.Valid {
			athleteInfo.TeamID = &user.TeamID.UUID
		}

		if user.TeamLogoUrl.Valid {
			athleteInfo.TeamLogoURL = &user.TeamLogoUrl.String
		}

		response.AthleteInfo = athleteInfo
	}
}

func setUserRole(ctx context.Context, r *UsersRepository, response *values.UserReadInfo, user dbIdentity.GetUserByIdOrEmailRow) {
	// Staff role takes priority
	if staffInfo, err := r.IdentityQueries.GetStaffById(ctx, user.ID); err == nil {
		response.Role = staffInfo.RoleName
		response.IsActiveStaff = &staffInfo.IsActive
		return
	}

	// All customers are athletes (including those with parent_id)
	if user.Wins.Valid {
		response.Role = "athlete"
		return
	}

	// Fallback to account_type for safety
	if user.AccountType.Valid {
		acctType := strings.ToLower(user.AccountType.String)
		// Map legacy types to athlete
		if acctType == "child" || acctType == "parent" {
			acctType = "athlete"
		}
		response.Role = acctType
	}
}

func addPhotoURL(ctx context.Context, r *UsersRepository, response *values.UserReadInfo, user dbIdentity.GetUserByIdOrEmailRow) {
	// Try to get photo URL from staff table first (for staff members)
	if staffInfo, err := r.IdentityQueries.GetStaffById(ctx, user.ID); err == nil && staffInfo.PhotoUrl.Valid {
		response.PhotoURL = &staffInfo.PhotoUrl.String
		return
	}

	// For athletes, the photo URL is already included in the main query
	if user.AthletePhotoUrl.Valid {
		response.PhotoURL = &user.AthletePhotoUrl.String
	}
}
