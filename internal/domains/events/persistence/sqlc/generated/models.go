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

type ClassLevel string

const (
	ClassLevelBeginner     ClassLevel = "beginner"
	ClassLevelIntermediate ClassLevel = "intermediate"
	ClassLevelAdvanced     ClassLevel = "advanced"
	ClassLevelAll          ClassLevel = "all"
)

func (e *ClassLevel) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ClassLevel(s)
	case string:
		*e = ClassLevel(s)
	default:
		return fmt.Errorf("unsupported scan type for ClassLevel: %T", src)
	}
	return nil
}

type NullClassLevel struct {
	ClassLevel ClassLevel `json:"class_level"`
	Valid      bool       `json:"valid"` // Valid is true if ClassLevel is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullClassLevel) Scan(value interface{}) error {
	if value == nil {
		ns.ClassLevel, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ClassLevel.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullClassLevel) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ClassLevel), nil
}

type DayEnum string

const (
	DayEnumMONDAY    DayEnum = "MONDAY"
	DayEnumTUESDAY   DayEnum = "TUESDAY"
	DayEnumWEDNESDAY DayEnum = "WEDNESDAY"
	DayEnumTHURSDAY  DayEnum = "THURSDAY"
	DayEnumFRIDAY    DayEnum = "FRIDAY"
	DayEnumSATURDAY  DayEnum = "SATURDAY"
	DayEnumSUNDAY    DayEnum = "SUNDAY"
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

type MembershipStatus string

const (
	MembershipStatusActive   MembershipStatus = "active"
	MembershipStatusInactive MembershipStatus = "inactive"
	MembershipStatusCanceled MembershipStatus = "canceled"
	MembershipStatusExpired  MembershipStatus = "expired"
)

func (e *MembershipStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MembershipStatus(s)
	case string:
		*e = MembershipStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for MembershipStatus: %T", src)
	}
	return nil
}

type NullMembershipStatus struct {
	MembershipStatus MembershipStatus `json:"membership_status"`
	Valid            bool             `json:"valid"` // Valid is true if MembershipStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMembershipStatus) Scan(value interface{}) error {
	if value == nil {
		ns.MembershipStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MembershipStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMembershipStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MembershipStatus), nil
}

type PaymentFrequency string

const (
	PaymentFrequencyOnce  PaymentFrequency = "once"
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

type Class struct {
	ID                       uuid.UUID      `json:"id"`
	Name                     string         `json:"name"`
	Description              sql.NullString `json:"description"`
	Level                    ClassLevel     `json:"level"`
	EmailBookingNotification sql.NullBool   `json:"email_booking_notification"`
	Capacity                 int32          `json:"capacity"`
	StartDate                time.Time      `json:"start_date"`
	EndDate                  sql.NullTime   `json:"end_date"`
	CreatedAt                sql.NullTime   `json:"created_at"`
	UpdatedAt                sql.NullTime   `json:"updated_at"`
}

type ClassMembership struct {
	ClassID         uuid.UUID      `json:"class_id"`
	MembershipID    uuid.UUID      `json:"membership_id"`
	PricePerBooking sql.NullString `json:"price_per_booking"`
	IsEligible      bool           `json:"is_eligible"`
}

type Course struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
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
	Credits   int32     `json:"credits"`
}

type CustomerDiscountUsage struct {
	CustomerID uuid.UUID    `json:"customer_id"`
	DiscountID uuid.UUID    `json:"discount_id"`
	UsageCount int32        `json:"usage_count"`
	LastUsedAt sql.NullTime `json:"last_used_at"`
}

type CustomerEvent struct {
	ID          uuid.UUID    `json:"id"`
	CustomerID  uuid.UUID    `json:"customer_id"`
	EventID     uuid.UUID    `json:"event_id"`
	CreatedAt   sql.NullTime `json:"created_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
	CheckedInAt sql.NullTime `json:"checked_in_at"`
	IsCancelled sql.NullBool `json:"is_cancelled"`
}

type CustomerMembershipPlan struct {
	ID               uuid.UUID        `json:"id"`
	CustomerID       uuid.UUID        `json:"customer_id"`
	MembershipPlanID uuid.UUID        `json:"membership_plan_id"`
	StartDate        sql.NullTime     `json:"start_date"`
	RenewalDate      sql.NullTime     `json:"renewal_date"`
	Status           MembershipStatus `json:"status"`
	CreatedAt        sql.NullTime     `json:"created_at"`
	UpdatedAt        sql.NullTime     `json:"updated_at"`
}

type Discount struct {
	ID              uuid.UUID      `json:"id"`
	Name            string         `json:"name"`
	Description     sql.NullString `json:"description"`
	DiscountPercent int32          `json:"discount_percent"`
	IsUseUnlimited  bool           `json:"is_use_unlimited"`
	UsePerClient    sql.NullInt32  `json:"use_per_client"`
	IsActive        bool           `json:"is_active"`
	ValidFrom       time.Time      `json:"valid_from"`
	ValidTo         sql.NullTime   `json:"valid_to"`
	CreatedAt       sql.NullTime   `json:"created_at"`
	UpdatedAt       sql.NullTime   `json:"updated_at"`
}

type DiscountRestrictedMembershipPlan struct {
	DiscountID       uuid.UUID    `json:"discount_id"`
	MembershipPlanID uuid.UUID    `json:"membership_plan_id"`
	CreatedAt        sql.NullTime `json:"created_at"`
}

type Event struct {
	ID         uuid.UUID     `json:"id"`
	BeginTime  time.Time     `json:"begin_time"`
	EndTime    time.Time     `json:"end_time"`
	CourseID   uuid.NullUUID `json:"course_id"`
	FacilityID uuid.UUID     `json:"facility_id"`
	CreatedAt  sql.NullTime  `json:"created_at"`
	UpdatedAt  sql.NullTime  `json:"updated_at"`
	Day        DayEnum       `json:"day"`
}

type EventStaff struct {
	EventID uuid.UUID `json:"event_id"`
	StaffID uuid.UUID `json:"staff_id"`
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
	CreatedAt   sql.NullTime   `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
}

type MembershipPlan struct {
	ID               uuid.UUID            `json:"id"`
	Name             string               `json:"name"`
	Price            int32                `json:"price"`
	JoiningFee       sql.NullInt32        `json:"joining_fee"`
	AutoRenew        bool                 `json:"auto_renew"`
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
	FirstName      sql.NullString `json:"first_name"`
	LastName       sql.NullString `json:"last_name"`
	Phone          sql.NullString `json:"phone"`
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
