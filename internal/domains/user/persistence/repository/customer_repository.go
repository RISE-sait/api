package user

import (
	"api/internal/di"
	db "api/internal/domains/user/persistence/sqlc/generated"
	userValues "api/internal/domains/user/values"
	errLib "api/internal/libs/errors"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CustomerRepository struct {
	Queries *db.Queries
	Db      *sql.DB
}

func NewCustomerRepository(container *di.Container) *CustomerRepository {
	return &CustomerRepository{
		Queries: container.Queries.UserDb,
		Db:      container.DB,
	}
}

func (r *CustomerRepository) WithTx(tx *sql.Tx) *CustomerRepository {
	return &CustomerRepository{
		Queries: r.Queries.WithTx(tx),
		Db:      r.Db,
	}
}

func (r *CustomerRepository) GetCustomers(ctx context.Context, limit, offset int32, filters userValues.CustomerFilterParams) ([]userValues.ReadValue, *errLib.CommonError) {

	params := db.GetCustomersParams{
		ParentID: uuid.NullUUID{
			UUID:  filters.ParentID,
			Valid: filters.ParentID != uuid.Nil,
		},
		Search: sql.NullString{
			String: filters.Search,
			Valid:  filters.Search != "",
		},
		Offset: offset,
		Limit:  limit,
	}

	// Set optional filter params
	if filters.MembershipPlanID != nil {
		params.MembershipPlanID = uuid.NullUUID{UUID: *filters.MembershipPlanID, Valid: true}
	}
	if filters.MembershipStatus != nil {
		params.MembershipStatus = sql.NullString{String: *filters.MembershipStatus, Valid: true}
	}
	if filters.HasMembership != nil {
		params.HasMembership = sql.NullBool{Bool: *filters.HasMembership, Valid: true}
	}
	if filters.HasCredits != nil {
		params.HasCredits = sql.NullBool{Bool: *filters.HasCredits, Valid: true}
	}
	if filters.MinCredits != nil {
		params.MinCredits = sql.NullInt32{Int32: *filters.MinCredits, Valid: true}
	}
	if filters.MaxCredits != nil {
		params.MaxCredits = sql.NullInt32{Int32: *filters.MaxCredits, Valid: true}
	}

	dbCustomers, err := r.Queries.GetCustomers(ctx, params)

	if err != nil {
		log.Printf("Error getting dbCustomers: %s", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	customers := make([]userValues.ReadValue, len(dbCustomers))

	for i, dbCustomer := range dbCustomers {
		customer := userValues.ReadValue{
			ID:          dbCustomer.ID,
			DOB:         dbCustomer.Dob,
			FirstName:   dbCustomer.FirstName,
			LastName:    dbCustomer.LastName,
			CountryCode: dbCustomer.CountryAlpha2Code,
			CreatedAt:   dbCustomer.CreatedAt,
			UpdatedAt:   dbCustomer.UpdatedAt,
			IsArchived:  dbCustomer.IsArchived,
		}

		if dbCustomer.DeletedAt.Valid {
			customer.DeletedAt = &dbCustomer.DeletedAt.Time
		}

		if dbCustomer.ScheduledDeletionAt.Valid {
			customer.ScheduledDeletionAt = &dbCustomer.ScheduledDeletionAt.Time
		}

		if dbCustomer.HubspotID.Valid {
			customer.HubspotID = &dbCustomer.HubspotID.String
		}

		if dbCustomer.Phone.Valid {
			customer.Phone = &dbCustomer.Phone.String
		}

		if dbCustomer.Email.Valid {
			customer.Email = &dbCustomer.Email.String
		}

		if dbCustomer.Notes.Valid {
			customer.Notes = &dbCustomer.Notes.String
		}

		if dbCustomer.EmergencyContactName.Valid {
			customer.EmergencyContactName = &dbCustomer.EmergencyContactName.String
		}

		if dbCustomer.EmergencyContactPhone.Valid {
			customer.EmergencyContactPhone = &dbCustomer.EmergencyContactPhone.String
		}

		if dbCustomer.EmergencyContactRelationship.Valid {
			customer.EmergencyContactRelationship = &dbCustomer.EmergencyContactRelationship.String
		}

		if dbCustomer.LastMobileLoginAt.Valid {
			customer.LastMobileLoginAt = &dbCustomer.LastMobileLoginAt.Time
		}

		if dbCustomer.PendingEmail.Valid {
			customer.PendingEmail = &dbCustomer.PendingEmail.String
		}

		if dbCustomer.MembershipName.Valid && dbCustomer.MembershipPlanName.Valid && dbCustomer.MembershipStartDate.Valid && dbCustomer.MembershipPlanID.Valid {

			customer.MembershipInfo = &userValues.MembershipReadValue{
				MembershipPlanID:      dbCustomer.MembershipPlanID.UUID,
				MembershipPlanName:    dbCustomer.MembershipPlanName.String,
				MembershipName:        dbCustomer.MembershipName.String,
				MembershipStartDate:   dbCustomer.MembershipStartDate.Time,
				MembershipRenewalDate: dbCustomer.MembershipPlanRenewalDate.Time,
				Status:                string(dbCustomer.MembershipStatus.MembershipMembershipStatus),
			}
		}

		if dbCustomer.Rebounds.Valid && dbCustomer.Wins.Valid && dbCustomer.Points.Valid && dbCustomer.Steals.Valid && dbCustomer.Assists.Valid && dbCustomer.Losses.Valid {
			customer.AthleteInfo = &userValues.AthleteReadValue{
				Wins:     dbCustomer.Wins.Int32,
				Losses:   dbCustomer.Losses.Int32,
				Points:   dbCustomer.Points.Int32,
				Steals:   dbCustomer.Steals.Int32,
				Assists:  dbCustomer.Assists.Int32,
				Rebounds: dbCustomer.Rebounds.Int32,
				PhotoURL: func(n sql.NullString) *string {
					if n.Valid {
						return &n.String
					}
					return nil
				}(dbCustomer.PhotoUrl),
			}
		}

		customers[i] = customer
	}

	return customers, nil
}

func (r *CustomerRepository) GetCustomer(ctx context.Context, id uuid.UUID, email string) (userValues.ReadValue, *errLib.CommonError) {

	dbCustomer, err := r.Queries.GetCustomer(ctx, db.GetCustomerParams{
		ID: uuid.NullUUID{
			UUID:  id,
			Valid: id != uuid.Nil,
		},
		Email: sql.NullString{
			String: email,
			Valid:  email != "",
		},
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userValues.ReadValue{}, errLib.New("Customer not found", http.StatusNotFound)
		}
		log.Printf("Error getting dbCustomer: %s", err)
		return userValues.ReadValue{}, errLib.New("internal error", http.StatusInternalServerError)
	}

	customer := userValues.ReadValue{
		ID:          dbCustomer.ID,
		DOB:         dbCustomer.Dob,
		FirstName:   dbCustomer.FirstName,
		LastName:    dbCustomer.LastName,
		CountryCode: dbCustomer.CountryAlpha2Code,
		CreatedAt:   dbCustomer.CreatedAt,
		UpdatedAt:   dbCustomer.UpdatedAt,
		IsArchived:  dbCustomer.IsArchived,
	}

	if dbCustomer.DeletedAt.Valid {
		customer.DeletedAt = &dbCustomer.DeletedAt.Time
	}

	if dbCustomer.ScheduledDeletionAt.Valid {
		customer.ScheduledDeletionAt = &dbCustomer.ScheduledDeletionAt.Time
	}

	if dbCustomer.HubspotID.Valid {
		customer.HubspotID = &dbCustomer.HubspotID.String
	}

	if dbCustomer.Phone.Valid {
		customer.Phone = &dbCustomer.Phone.String
	}

	if dbCustomer.Email.Valid {
		customer.Email = &dbCustomer.Email.String
	}

	if dbCustomer.Notes.Valid {
		customer.Notes = &dbCustomer.Notes.String
	}

	if dbCustomer.EmergencyContactName.Valid {
		customer.EmergencyContactName = &dbCustomer.EmergencyContactName.String
	}

	if dbCustomer.EmergencyContactPhone.Valid {
		customer.EmergencyContactPhone = &dbCustomer.EmergencyContactPhone.String
	}

	if dbCustomer.EmergencyContactRelationship.Valid {
		customer.EmergencyContactRelationship = &dbCustomer.EmergencyContactRelationship.String
	}

	if dbCustomer.LastMobileLoginAt.Valid {
		customer.LastMobileLoginAt = &dbCustomer.LastMobileLoginAt.Time
	}

	if dbCustomer.PendingEmail.Valid {
		customer.PendingEmail = &dbCustomer.PendingEmail.String
	}

	if dbCustomer.MembershipName.Valid && dbCustomer.MembershipPlanName.Valid && dbCustomer.MembershipStartDate.Valid && dbCustomer.MembershipPlanID.Valid {

		customer.MembershipInfo = &userValues.MembershipReadValue{
			MembershipPlanID:      dbCustomer.MembershipPlanID.UUID,
			MembershipPlanName:    dbCustomer.MembershipPlanName.String,
			MembershipName:        dbCustomer.MembershipName.String,
			MembershipStartDate:   dbCustomer.MembershipStartDate.Time,
			MembershipRenewalDate: dbCustomer.MembershipPlanRenewalDate.Time,
			Status:                string(dbCustomer.MembershipStatus.MembershipMembershipStatus),
		}
	}

	if dbCustomer.Rebounds.Valid && dbCustomer.Wins.Valid && dbCustomer.Points.Valid && dbCustomer.Steals.Valid && dbCustomer.Assists.Valid && dbCustomer.Losses.Valid {
		customer.AthleteInfo = &userValues.AthleteReadValue{
			Wins:     dbCustomer.Wins.Int32,
			Losses:   dbCustomer.Losses.Int32,
			Points:   dbCustomer.Points.Int32,
			Steals:   dbCustomer.Steals.Int32,
			Assists:  dbCustomer.Assists.Int32,
			Rebounds: dbCustomer.Rebounds.Int32,
			PhotoURL: func(n sql.NullString) *string {
				if n.Valid {
					return &n.String
				}
				return nil
			}(dbCustomer.PhotoUrl),
		}
	}

	return customer, nil
}

func (r *CustomerRepository) UpdateAthleteTeam(ctx context.Context, athleteID, teamID uuid.UUID) *errLib.CommonError {

	args := db.UpdateAthleteTeamParams{
		AthleteID: athleteID,
		TeamID: uuid.NullUUID{
			UUID:  teamID,
			Valid: teamID != uuid.Nil,
		},
	}

	updatedRows, err := r.Queries.UpdateAthleteTeam(ctx, args)

	if err != nil {

		var pqErr *pq.Error

		if errors.As(err, &pqErr) {
			if pqErr.Constraint == "fk_team" {
				return errLib.New("Team not found", http.StatusNotFound)
			}
		}

		log.Printf("Unhandled error when updating athlete's team: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Athlete not found", http.StatusNotFound)
	}

	return nil
}

func (r *CustomerRepository) UpdateStats(ctx context.Context, valuesToUpdate userValues.StatsUpdateValue) *errLib.CommonError {

	var args db.UpdateAthleteStatsParams

	if valuesToUpdate.Wins != nil {
		args.Wins = sql.NullInt32{
			Int32: *valuesToUpdate.Wins,
			Valid: true,
		}
	}

	if valuesToUpdate.Losses != nil {
		args.Losses = sql.NullInt32{
			Int32: *valuesToUpdate.Losses,
			Valid: true,
		}
	}

	if valuesToUpdate.Points != nil {
		args.Points = sql.NullInt32{
			Int32: *valuesToUpdate.Points,
			Valid: true,
		}
	}

	if valuesToUpdate.Steals != nil {
		args.Steals = sql.NullInt32{
			Int32: *valuesToUpdate.Steals,
			Valid: true,
		}
	}

	if valuesToUpdate.Assists != nil {
		args.Assists = sql.NullInt32{
			Int32: *valuesToUpdate.Assists,
			Valid: true,
		}
	}

	if valuesToUpdate.Rebounds != nil {
		args.Rebounds = sql.NullInt32{
			Int32: *valuesToUpdate.Rebounds,
			Valid: true,
		}
	}

	updatedRows, err := r.Queries.UpdateAthleteStats(ctx, args)

	if err != nil {

		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Person with the associated ID not found", http.StatusNotFound)
	}

	return nil
}

func (r *CustomerRepository) UpdateAthleteProfile(ctx context.Context, valuesToUpdate userValues.AthleteProfileUpdateValue) *errLib.CommonError {

	var photoURL sql.NullString
	if valuesToUpdate.PhotoURL != nil {
		photoURL = sql.NullString{String: *valuesToUpdate.PhotoURL, Valid: true}
	}

	args := db.UpdateAthleteProfileParams{
		ID:       valuesToUpdate.ID,
		PhotoUrl: photoURL,
	}

	updatedRows, err := r.Queries.UpdateAthleteProfile(ctx, args)

	if err != nil {
		log.Printf("Unhandled error: %v", err)
		return errLib.New("Internal server error", http.StatusInternalServerError)
	}

	if updatedRows == 0 {
		return errLib.New("Athlete with the associated ID not found", http.StatusNotFound)
	}

	return nil
}
func (r *CustomerRepository) ListAthletes(ctx context.Context, limit, offset int32) ([]userValues.AthleteReadValue, *errLib.CommonError) {
	dbAthletes, err := r.Queries.GetAthletes(ctx, db.GetAthletesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Printf("Error getting athletes: %s", err)
		return nil, errLib.New("Internal server error", http.StatusInternalServerError)
	}

	athletes := make([]userValues.AthleteReadValue, len(dbAthletes))
	for i, a := range dbAthletes {
		athletes[i] = userValues.AthleteReadValue{
			ID:        a.ID,
			FirstName: a.FirstName,
			LastName:  a.LastName,
			Points:    a.Points,
			Wins:      a.Wins,
			Losses:    a.Losses,
			Assists:   a.Assists,
			Rebounds:  a.Rebounds,
			Steals:    a.Steals,
			PhotoURL:  ToStringPtr(a.PhotoUrl),
			TeamID: func(n uuid.NullUUID) *uuid.UUID {
				if n.Valid {
					return &n.UUID
				}
				return nil
			}(a.TeamID),
		}
	}

	return athletes, nil
}
func (r *CustomerRepository) CountCustomers(ctx context.Context, filters userValues.CustomerFilterParams) (int64, *errLib.CommonError) {
	params := db.CountCustomersParams{
		ParentID: uuid.NullUUID{
			UUID:  filters.ParentID,
			Valid: filters.ParentID != uuid.Nil,
		},
		Search: sql.NullString{
			String: filters.Search,
			Valid:  filters.Search != "",
		},
	}

	// Set optional filter params
	if filters.MembershipPlanID != nil {
		params.MembershipPlanID = uuid.NullUUID{UUID: *filters.MembershipPlanID, Valid: true}
	}
	if filters.MembershipStatus != nil {
		params.MembershipStatus = sql.NullString{String: *filters.MembershipStatus, Valid: true}
	}
	if filters.HasMembership != nil {
		params.HasMembership = sql.NullBool{Bool: *filters.HasMembership, Valid: true}
	}
	if filters.HasCredits != nil {
		params.HasCredits = sql.NullBool{Bool: *filters.HasCredits, Valid: true}
	}
	if filters.MinCredits != nil {
		params.MinCredits = sql.NullInt32{Int32: *filters.MinCredits, Valid: true}
	}
	if filters.MaxCredits != nil {
		params.MaxCredits = sql.NullInt32{Int32: *filters.MaxCredits, Valid: true}
	}

	count, err := r.Queries.CountCustomers(ctx, params)
	if err != nil {
		return 0, errLib.New("Failed to count customers: "+err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (r *CustomerRepository) CountActiveMembers(ctx context.Context) (int64, *errLib.CommonError) {
	count, err := r.Queries.CountActiveMembers(ctx)
	if err != nil {
		return 0, errLib.New("Failed to count active members: "+err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (r *CustomerRepository) GetActiveMembershipInfo(ctx context.Context, customerID uuid.UUID) (*userValues.MembershipPlansReadValue, *errLib.CommonError) {
	row, err := r.Queries.GetActiveMembershipInfo(ctx, customerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errLib.New("active membership not found", http.StatusNotFound)
		}
		log.Printf("error fetching active membership info: %v", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	info := userValues.MembershipPlansReadValue{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		StartDate:  row.StartDate,
		Status:     string(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		PhotoURL: func(n sql.NullString) *string {
			if n.Valid {
				return &n.String
			}
			return nil
		}(row.PhotoUrl),
		MembershipID:       row.MembershipID,
		MembershipPlanID:   row.MembershipPlanID,
		MembershipName:     row.MembershipName,
		MembershipPlanName: row.MembershipPlanName,
	}

	if row.RenewalDate.Valid {
		info.RenewalDate = &row.RenewalDate.Time
	}

	return &info, nil
}
func (r *CustomerRepository) ListMembershipHistory(ctx context.Context, customerID uuid.UUID) ([]userValues.MembershipHistoryValue, *errLib.CommonError) {
	dbRows, err := r.Queries.ListMembershipHistory(ctx, customerID)
	if err != nil {
		log.Printf("error querying membership history: %v", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}

	results := make([]userValues.MembershipHistoryValue, len(dbRows))
	for i, row := range dbRows {
		var renewal *time.Time
		if row.RenewalDate.Valid {
			renewal = &row.RenewalDate.Time
		}

		results[i] = userValues.MembershipHistoryValue{
			ID:                    row.ID,
			CustomerID:            row.CustomerID,
			StartDate:             row.StartDate,
			RenewalDate:           renewal,
			Status:                string(row.Status),
			CreatedAt:             row.CreatedAt,
			UpdatedAt:             row.UpdatedAt,
			MembershipID:          row.MembershipID,
			MembershipName:        row.MembershipName,
			MembershipDescription: row.MembershipDescription,
			MembershipBenefits:    row.MembershipBenefits,
			MembershipPlanID:      row.MembershipPlanID,
			MembershipPlanName:    row.MembershipPlanName,
			UnitAmount: func(n sql.NullInt32) int {
				if n.Valid {
					return int(n.Int32)
				}
				return 0
			}(row.UnitAmount),
			Currency: func(n sql.NullString) string {
				if n.Valid {
					return n.String
				}
				return ""
			}(row.Currency),
			Interval: func(n sql.NullString) string {
				if n.Valid {
					return n.String
				}
				return ""
			}(row.Interval),
			StripePriceID: row.StripePriceID,
		}
	}

	return results, nil
}

func (r *CustomerRepository) ArchiveCustomer(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	affected, err := r.Queries.ArchiveCustomer(ctx, id)
	if err != nil {
		log.Printf("error archiving customer: %v", err)
		return errLib.New("internal error", http.StatusInternalServerError)
	}
	if affected == 0 {
		return errLib.New("customer not found", http.StatusNotFound)
	}
	return nil
}

func (r *CustomerRepository) UnarchiveCustomer(ctx context.Context, id uuid.UUID) *errLib.CommonError {
	affected, err := r.Queries.UnarchiveCustomer(ctx, id)
	if err != nil {
		log.Printf("error unarchiving customer: %v", err)
		return errLib.New("internal error", http.StatusInternalServerError)
	}
	if affected == 0 {
		return errLib.New("customer not found", http.StatusNotFound)
	}
	return nil
}

func (r *CustomerRepository) ListArchivedCustomers(ctx context.Context, limit, offset int32) ([]userValues.ReadValue, *errLib.CommonError) {
	rows, err := r.Queries.ListArchivedCustomers(ctx, db.ListArchivedCustomersParams{Offset: offset, Limit: limit})
	if err != nil {
		log.Printf("error listing archived customers: %v", err)
		return nil, errLib.New("internal error", http.StatusInternalServerError)
	}
	customers := make([]userValues.ReadValue, len(rows))
	for i, dbCustomer := range rows {
		customers[i] = userValues.ReadValue{
			ID:          dbCustomer.ID,
			DOB:         dbCustomer.Dob,
			FirstName:   dbCustomer.FirstName,
			LastName:    dbCustomer.LastName,
			CountryCode: dbCustomer.CountryAlpha2Code,
			CreatedAt:   dbCustomer.CreatedAt,
			UpdatedAt:   dbCustomer.UpdatedAt,
			IsArchived:  dbCustomer.IsArchived,
		}

		if dbCustomer.HubspotID.Valid {
			customers[i].HubspotID = &dbCustomer.HubspotID.String
		}
		if dbCustomer.Phone.Valid {
			customers[i].Phone = &dbCustomer.Phone.String
		}
		if dbCustomer.Email.Valid {
			customers[i].Email = &dbCustomer.Email.String
		}
	}
	return customers, nil
}

// DeleteCustomerAccountCompletely performs a complete account deletion including all related data
func (r *CustomerRepository) DeleteCustomerAccountCompletely(ctx context.Context, customerID uuid.UUID) *errLib.CommonError {
	log.Printf("Starting complete account deletion for customer: %s", customerID)

	// Start a transaction for atomic deletion
	tx, err := r.Db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction for customer deletion: %v", err)
		return errLib.New("Failed to start deletion process", http.StatusInternalServerError)
	}
	defer tx.Rollback()

	txRepo := r.WithTx(tx)

	// 1. Delete customer memberships
	_, err = txRepo.Queries.DeleteCustomerMemberships(ctx, customerID)
	if err != nil {
		log.Printf("Failed to delete customer memberships: %v", err)
		return errLib.New("Failed to delete customer memberships", http.StatusInternalServerError)
	}
	log.Printf("Deleted customer memberships for: %s", customerID)

	// 2. Delete program enrollments
	_, err = txRepo.Queries.DeleteCustomerEnrollments(ctx, customerID)
	if err != nil {
		log.Printf("Failed to delete customer program enrollments: %v", err)
		return errLib.New("Failed to delete customer enrollments", http.StatusInternalServerError)
	}
	log.Printf("Deleted program enrollments for: %s", customerID)

	// 3. Delete event enrollments
	_, err = txRepo.Queries.DeleteCustomerEventEnrollments(ctx, customerID)
	if err != nil {
		log.Printf("Failed to delete customer event enrollments: %v", err)
		return errLib.New("Failed to delete event enrollments", http.StatusInternalServerError)
	}
	log.Printf("Deleted event enrollments for: %s", customerID)

	// 4. Delete athlete data if exists
	_, err = txRepo.Queries.DeleteAthleteData(ctx, customerID)
	if err != nil {
		log.Printf("Failed to delete athlete data (may not exist): %v", err)
		// Don't fail the entire process if athlete data doesn't exist
	} else {
		log.Printf("Deleted athlete data for: %s", customerID)
	}

	// 5. Delete staff data if exists (to avoid foreign key constraint violation)
	_, err = tx.ExecContext(ctx, "DELETE FROM staff.staff WHERE id = $1", customerID)
	if err != nil {
		log.Printf("Failed to delete staff data (may not exist): %v", err)
		// Don't fail the entire process if staff data doesn't exist
	} else {
		log.Printf("Deleted staff data for: %s", customerID)
	}

	// 6. Finally delete the user account (this will cascade to credits and other ON DELETE CASCADE tables)
	affected, err := txRepo.Queries.DeleteCustomerAccount(ctx, customerID)
	if err != nil {
		log.Printf("Failed to delete customer account: %v", err)
		return errLib.New("Failed to delete customer account", http.StatusInternalServerError)
	}
	if affected == 0 {
		log.Printf("Customer not found during deletion: %s", customerID)
		return errLib.New("Customer not found", http.StatusNotFound)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Failed to commit customer deletion transaction: %v", err)
		return errLib.New("Failed to complete account deletion", http.StatusInternalServerError)
	}

	log.Printf("Successfully completed account deletion for customer: %s", customerID)
	return nil
}

// UpdateCustomerNotes updates the notes for a customer
func (r *CustomerRepository) UpdateCustomerNotes(ctx context.Context, updateValue userValues.NotesUpdateValue) (int64, *errLib.CommonError) {
	var notes sql.NullString
	if updateValue.Notes != nil {
		notes = sql.NullString{String: *updateValue.Notes, Valid: true}
	} else {
		notes = sql.NullString{Valid: false}
	}

	rowsAffected, err := r.Queries.UpdateCustomerNotes(ctx, db.UpdateCustomerNotesParams{
		CustomerID: updateValue.CustomerID,
		Notes:      notes,
	})

	if err != nil {
		log.Printf("Failed to update customer notes for customer %s: %v", updateValue.CustomerID, err)
		return 0, errLib.New("Failed to update customer notes", http.StatusInternalServerError)
	}

	return rowsAffected, nil
}
