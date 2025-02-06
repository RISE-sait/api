// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DayEnum string

const (
	DayEnumM     DayEnum = "M"
	DayEnumTues  DayEnum = "Tues"
	DayEnumW     DayEnum = "W"
	DayEnumThurs DayEnum = "Thurs"
	DayEnumF     DayEnum = "F"
	DayEnumSat   DayEnum = "Sat"
	DayEnumSun   DayEnum = "Sun"
)

func (e *DayEnum) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = DayEnum(s)
	case string:
		*e = DayEnum(s)
	default:
		return fmt.Errorf("unsupported scan type for DayEnum: %T", src)
	}
	return nil
}

type NullDayEnum struct {
	DayEnum DayEnum `json:"day_enum"`
	Valid   bool    `json:"valid"` // Valid is true if DayEnum is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullDayEnum) Scan(value interface{}) error {
	if value == nil {
		ns.DayEnum, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.DayEnum.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullDayEnum) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.DayEnum), nil
}

type PaymentFrequency string

const (
	PaymentFrequencyWeek  PaymentFrequency = "week"
	PaymentFrequencyMonth PaymentFrequency = "month"
	PaymentFrequencyDay   PaymentFrequency = "day"
)

func (e *PaymentFrequency) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = PaymentFrequency(s)
	case string:
		*e = PaymentFrequency(s)
	default:
		return fmt.Errorf("unsupported scan type for PaymentFrequency: %T", src)
	}
	return nil
}

type NullPaymentFrequency struct {
	PaymentFrequency PaymentFrequency `json:"payment_frequency"`
	Valid            bool             `json:"valid"` // Valid is true if PaymentFrequency is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullPaymentFrequency) Scan(value interface{}) error {
	if value == nil {
		ns.PaymentFrequency, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.PaymentFrequency.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullPaymentFrequency) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.PaymentFrequency), nil
}

type Course struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	CreatedAt   sql.NullTime   `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
}

type CourseMembership struct {
	CourseID        uuid.UUID      `json:"course_id"`
	MembershipID    uuid.UUID      `json:"membership_id"`
	PricePerBooking sql.NullString `json:"price_per_booking"`
	IsEligible      bool           `json:"is_eligible"`
}

type Customer struct {
	UserID    uuid.UUID `json:"user_id"`
	HubspotID int64     `json:"hubspot_id"`
}

type Facility struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Location       string    `json:"location"`
	FacilityTypeID uuid.UUID `json:"facility_type_id"`
}

type FacilityType struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Membership struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	CreatedAt   sql.NullTime   `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
}

type MembershipPlan struct {
	ID               uuid.UUID            `json:"id"`
	Name             string               `json:"name"`
	Price            int64                `json:"price"`
	MembershipID     uuid.UUID            `json:"membership_id"`
	PaymentFrequency NullPaymentFrequency `json:"payment_frequency"`
	AmtPeriods       sql.NullInt32        `json:"amt_periods"`
	CreatedAt        sql.NullTime         `json:"created_at"`
	UpdatedAt        sql.NullTime         `json:"updated_at"`
}

type PendingAccountsWaiverSigning struct {
	UserID    uuid.UUID `json:"user_id"`
	WaiverID  uuid.UUID `json:"waiver_id"`
	IsSigned  bool      `json:"is_signed"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PendingChildAccount struct {
	ID          uuid.UUID      `json:"id"`
	ParentEmail string         `json:"parent_email"`
	UserEmail   string         `json:"user_email"`
	Password    sql.NullString `json:"password"`
	CreatedAt   time.Time      `json:"created_at"`
}

type Schedule struct {
	ID         uuid.UUID     `json:"id"`
	BeginTime  time.Time     `json:"begin_time"`
	EndTime    time.Time     `json:"end_time"`
	CourseID   uuid.NullUUID `json:"course_id"`
	FacilityID uuid.UUID     `json:"facility_id"`
	CreatedAt  sql.NullTime  `json:"created_at"`
	UpdatedAt  sql.NullTime  `json:"updated_at"`
	Day        DayEnum       `json:"day"`
}

type Staff struct {
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
}

type StaffActivityLog struct {
	ID         uuid.UUID    `json:"id"`
	UserID     uuid.UUID    `json:"user_id"`
	Activity   string       `json:"activity"`
	OccurredAt sql.NullTime `json:"occurred_at"`
}

type StaffRole struct {
	ID       uuid.UUID `json:"id"`
	RoleName string    `json:"role_name"`
}

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type UserOptionalInfo struct {
	ID             uuid.UUID      `json:"id"`
	Name           sql.NullString `json:"name"`
	HashedPassword sql.NullString `json:"hashed_password"`
}

type Waiver struct {
	ID        uuid.UUID `json:"id"`
	WaiverUrl string    `json:"waiver_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WaiverSigning struct {
	UserID    uuid.UUID `json:"user_id"`
	WaiverID  uuid.UUID `json:"waiver_id"`
	IsSigned  bool      `json:"is_signed"`
	UpdatedAt time.Time `json:"updated_at"`
}
