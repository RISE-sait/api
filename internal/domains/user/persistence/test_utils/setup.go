package user

import (
	db "api/internal/domains/user/persistence/sqlc/generated"
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

//func TestGetCustomerTeam(t *testing.T) {
//
//	db, cleanup := dbTestUtils.SetupTestDbQueries(t, "../../../../../../db/migrations")
//
//	identityQ := identityDb.New(db)
//	userQ := userDb.New(db)
//	teamQ := teamDb.New(db)
//	paymentQ := paymentDb.New(db)
//
//	defer cleanup()
//
//	createdCustomer, err := identityQ.CreateUser(context.Background(), identityDb.CreateUserParams{
//		CountryAlpha2Code: "CA",
//		Dob:               time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
//		Phone: sql.NullString{
//			String: "+1514123456337",
//			Valid:  true,
//		},
//		HasMarketingEmailConsent: false,
//		HasSmsConsent:            false,
//		FirstName:                "John",
//		LastName:                 "Doe",
//	})
//
//	require.NoError(t, err)
//
//	err = identityQ.CreateAthlete(context.Background(), createdCustomer.ID)
//
//	require.NoError(t, err)
//
//	createdStaffRole, err := userQ.CreateStaffRole(context.Background(), "coach")
//
//	require.NoError(t, err)
//
//	createdPendingStaff, err := identityQ.CreatePendingStaff(context.Background(), identityDb.CreatePendingStaffParams{
//		CountryAlpha2Code: "CA",
//		Dob:               time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
//		Phone: sql.NullString{
//			String: "+14141234567",
//			Valid:  true,
//		},
//		Email: "klintlee1@gmail.com",
//		Gender: sql.NullString{
//			String: "M",
//			Valid:  true,
//		},
//		RoleName:  createdStaffRole.RoleName,
//		FirstName: "John",
//		LastName:  "Doe",
//	})
//
//	require.NoError(t, err)
//
//	createdStaff, err := identityQ.ApproveStaff(context.Background(), createdPendingStaff.ID)
//	require.NoError(t, err)
//
//	createdTeam, err := teamQ.CreateTeam(context.Background(), teamDb.CreateTeamParams{
//		Name:     "Go Team",
//		Capacity: 20,
//		CoachID: uuid.NullUUID{
//			UUID:  createdStaff.ID,
//			Valid: true,
//		},
//	})
//
//	require.NoError(t, err)
//
//	impactedRows, err := userQ.AddAthleteToTeam(context.Background(), userDb.AddAthleteToTeamParams{
//		CustomerID: createdCustomer.ID,
//		TeamID: uuid.NullUUID{
//			UUID:  createdTeam.ID,
//			Valid: true,
//		},
//	})
//
//	require.NoError(t, err)
//	require.Equal(t, int64(1), impactedRows)
//
//	team, err := paymentQ.GetCustomersTeam(context.Background(), createdCustomer.ID)
//
//	require.NoError(t, err)
//
//	require.True(t, team.ID.Valid)
//	require.Equal(t, createdTeam.ID, team.ID.UUID)
//}

func SetupUsersTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists users;

	create table if not exists users.users
(
    id                          uuid                     default gen_random_uuid() not null
        primary key,
    hubspot_id                  text
        unique,
    country_alpha2_code         char(2)                                            not null,
    gender                      char
        constraint users_gender_check
            check (gender = ANY (ARRAY ['M'::bpchar, 'F'::bpchar])),
    first_name                  varchar(20)                                        not null,
    last_name                   varchar(20)                                        not null,
    age                         integer                                            not null,
    parent_id                   uuid
        references users.users,
    phone                       varchar(25),
    email                       varchar(255)
        unique,
    has_marketing_email_consent boolean                                            not null,
    has_sms_consent             boolean                                            not null,
    created_at                  timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at                  timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table users.users
    owner to postgres;
`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `DELETE FROM users.users;`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}

func SetupStaffsTestDb(t *testing.T, testDb *sql.DB) (*db.Queries, func()) {

	migrationScript := `

create schema if not exists staff;

create table if not exists staff.staff_roles
(
    id         uuid                     default gen_random_uuid() not null
        primary key,
    role_name  text                                               not null
        unique,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null
);

alter table staff.staff_roles
    owner to postgres;

create table if not exists staff.staff
(
    id         uuid                                               not null
        primary key
        references users.users,
    is_active  boolean                  default true              not null,
    created_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    updated_at timestamp with time zone default CURRENT_TIMESTAMP not null,
    role_id    uuid                                               not null
        references staff.staff_roles
);

alter table staff.staff
    owner to postgres;
`

	_, err := testDb.Exec(migrationScript)
	require.NoError(t, err)

	// Return the repo and cleanup function
	repo := db.New(testDb)
	cleanUpScript := `
DELETE FROM staff.staff; 
DELETE  FROM staff.staff_roles;
`

	// Cleanup function to delete data after test
	return repo, func() {
		_, err = testDb.Exec(cleanUpScript)
		require.NoError(t, err)
	}
}
