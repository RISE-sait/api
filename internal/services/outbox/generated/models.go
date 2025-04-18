// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db_outbox

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AuditStatus string

const (
	AuditStatusPENDING   AuditStatus = "PENDING"
	AuditStatusCOMPLETED AuditStatus = "COMPLETED"
	AuditStatusFAILED    AuditStatus = "FAILED"
)

func (e *AuditStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AuditStatus(s)
	case string:
		*e = AuditStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for AuditStatus: %T", src)
	}
	return nil
}

type NullAuditStatus struct {
	AuditStatus AuditStatus `json:"audit_status"`
	Valid       bool        `json:"valid"` // Valid is true if AuditStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAuditStatus) Scan(value interface{}) error {
	if value == nil {
		ns.AuditStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AuditStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAuditStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AuditStatus), nil
}

func (e AuditStatus) Valid() bool {
	switch e {
	case AuditStatusPENDING,
		AuditStatusCOMPLETED,
		AuditStatusFAILED:
		return true
	}
	return false
}

func AllAuditStatusValues() []AuditStatus {
	return []AuditStatus{
		AuditStatusPENDING,
		AuditStatusCOMPLETED,
		AuditStatusFAILED,
	}
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

func (e DayEnum) Valid() bool {
	switch e {
	case DayEnumMONDAY,
		DayEnumTUESDAY,
		DayEnumWEDNESDAY,
		DayEnumTHURSDAY,
		DayEnumFRIDAY,
		DayEnumSATURDAY,
		DayEnumSUNDAY:
		return true
	}
	return false
}

func AllDayEnumValues() []DayEnum {
	return []DayEnum{
		DayEnumMONDAY,
		DayEnumTUESDAY,
		DayEnumWEDNESDAY,
		DayEnumTHURSDAY,
		DayEnumFRIDAY,
		DayEnumSATURDAY,
		DayEnumSUNDAY,
	}
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

func (e MembershipStatus) Valid() bool {
	switch e {
	case MembershipStatusActive,
		MembershipStatusInactive,
		MembershipStatusCanceled,
		MembershipStatusExpired:
		return true
	}
	return false
}

func AllMembershipStatusValues() []MembershipStatus {
	return []MembershipStatus{
		MembershipStatusActive,
		MembershipStatusInactive,
		MembershipStatusCanceled,
		MembershipStatusExpired,
	}
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

func (e PaymentFrequency) Valid() bool {
	switch e {
	case PaymentFrequencyOnce,
		PaymentFrequencyWeek,
		PaymentFrequencyMonth,
		PaymentFrequencyDay:
		return true
	}
	return false
}

func AllPaymentFrequencyValues() []PaymentFrequency {
	return []PaymentFrequency{
		PaymentFrequencyOnce,
		PaymentFrequencyWeek,
		PaymentFrequencyMonth,
		PaymentFrequencyDay,
	}
}

type PracticeLevel string

const (
	PracticeLevelBeginner     PracticeLevel = "beginner"
	PracticeLevelIntermediate PracticeLevel = "intermediate"
	PracticeLevelAdvanced     PracticeLevel = "advanced"
	PracticeLevelAll          PracticeLevel = "all"
)

func (e *PracticeLevel) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = PracticeLevel(s)
	case string:
		*e = PracticeLevel(s)
	default:
		return fmt.Errorf("unsupported scan type for PracticeLevel: %T", src)
	}
	return nil
}

type NullPracticeLevel struct {
	PracticeLevel PracticeLevel `json:"practice_level"`
	Valid         bool          `json:"valid"` // Valid is true if PracticeLevel is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullPracticeLevel) Scan(value interface{}) error {
	if value == nil {
		ns.PracticeLevel, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.PracticeLevel.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullPracticeLevel) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.PracticeLevel), nil
}

func (e PracticeLevel) Valid() bool {
	switch e {
	case PracticeLevelBeginner,
		PracticeLevelIntermediate,
		PracticeLevelAdvanced,
		PracticeLevelAll:
		return true
	}
	return false
}

func AllPracticeLevelValues() []PracticeLevel {
	return []PracticeLevel{
		PracticeLevelBeginner,
		PracticeLevelIntermediate,
		PracticeLevelAdvanced,
		PracticeLevelAll,
	}
}

type AuditOutbox struct {
	ID           uuid.UUID   `json:"id"`
	SqlStatement string      `json:"sql_statement"`
	Status       AuditStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
}

type BarberBarberEvent struct {
	ID            uuid.UUID `json:"id"`
	BeginDateTime time.Time `json:"begin_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	CustomerID    uuid.UUID `json:"customer_id"`
	BarberID      uuid.UUID `json:"barber_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CourseCourse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Capacity    int32     `json:"capacity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CourseMembership struct {
	CourseID        uuid.UUID      `json:"course_id"`
	MembershipID    uuid.UUID      `json:"membership_id"`
	PricePerBooking sql.NullString `json:"price_per_booking"`
	IsEligible      bool           `json:"is_eligible"`
}

type CustomerDiscountUsage struct {
	CustomerID uuid.UUID `json:"customer_id"`
	DiscountID uuid.UUID `json:"discount_id"`
	UsageCount int32     `json:"usage_count"`
	LastUsedAt time.Time `json:"last_used_at"`
}

type CustomerEnrollment struct {
	ID          uuid.UUID    `json:"id"`
	CustomerID  uuid.UUID    `json:"customer_id"`
	EventID     uuid.UUID    `json:"event_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CheckedInAt sql.NullTime `json:"checked_in_at"`
	IsCancelled bool         `json:"is_cancelled"`
}

type CustomerMembershipPlan struct {
	ID               uuid.UUID        `json:"id"`
	CustomerID       uuid.UUID        `json:"customer_id"`
	MembershipPlanID uuid.UUID        `json:"membership_plan_id"`
	StartDate        time.Time        `json:"start_date"`
	RenewalDate      sql.NullTime     `json:"renewal_date"`
	Status           MembershipStatus `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
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
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type DiscountRestrictedMembershipPlan struct {
	DiscountID       uuid.UUID `json:"discount_id"`
	MembershipPlanID uuid.UUID `json:"membership_plan_id"`
	CreatedAt        time.Time `json:"created_at"`
}

type Event struct {
	ID               uuid.UUID     `json:"id"`
	EventStartAt     time.Time     `json:"event_start_at"`
	EventEndAt       time.Time     `json:"event_end_at"`
	SessionStartTime interface{}   `json:"session_start_time"`
	SessionEndTime   interface{}   `json:"session_end_time"`
	Day              DayEnum       `json:"day"`
	PracticeID       uuid.NullUUID `json:"practice_id"`
	CourseID         uuid.NullUUID `json:"course_id"`
	GameID           uuid.NullUUID `json:"game_id"`
	LocationID       uuid.UUID     `json:"location_id"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type EventStaff struct {
	EventID uuid.UUID `json:"event_id"`
	StaffID uuid.UUID `json:"staff_id"`
}

type Game struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type LocationLocation struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Address string    `json:"address"`
}

type MembershipMembership struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	CreatedAt   sql.NullTime   `json:"created_at"`
	UpdatedAt   sql.NullTime   `json:"updated_at"`
}

type MembershipMembershipPlan struct {
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

type Practice struct {
	ID                             uuid.UUID     `json:"id"`
	Name                           string        `json:"name"`
	Description                    string        `json:"description"`
	Level                          PracticeLevel `json:"level"`
	ShouldEmailBookingNotification sql.NullBool  `json:"should_email_booking_notification"`
	Capacity                       int32         `json:"capacity"`
	CreatedAt                      time.Time     `json:"created_at"`
	UpdatedAt                      time.Time     `json:"updated_at"`
}

type PracticeMembership struct {
	PracticeID      uuid.UUID      `json:"practice_id"`
	MembershipID    uuid.UUID      `json:"membership_id"`
	PricePerBooking sql.NullString `json:"price_per_booking"`
	IsEligible      bool           `json:"is_eligible"`
}

type UsersAthlete struct {
	ID        uuid.UUID `json:"id"`
	Wins      int32     `json:"wins"`
	Losses    int32     `json:"losses"`
	Points    int32     `json:"points"`
	Steals    int32     `json:"steals"`
	Assists   int32     `json:"assists"`
	Rebounds  int32     `json:"rebounds"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UsersCustomerCredit struct {
	CustomerID uuid.UUID `json:"customer_id"`
	Credits    int32     `json:"credits"`
}

type UsersStaff struct {
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
}

type UsersStaffActivityLog struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Activity   string    `json:"activity"`
	OccurredAt time.Time `json:"occurred_at"`
}

type UsersStaffRole struct {
	ID       uuid.UUID `json:"id"`
	RoleName string    `json:"role_name"`
}

type UsersUser struct {
	ID                       uuid.UUID      `json:"id"`
	HubspotID                sql.NullString `json:"hubspot_id"`
	CountryAlpha2Code        string         `json:"country_alpha2_code"`
	FirstName                string         `json:"first_name"`
	LastName                 string         `json:"last_name"`
	Age                      int32          `json:"age"`
	ParentID                 uuid.NullUUID  `json:"parent_id"`
	Phone                    sql.NullString `json:"phone"`
	Email                    sql.NullString `json:"email"`
	HasMarketingEmailConsent bool           `json:"has_marketing_email_consent"`
	HasSmsConsent            bool           `json:"has_sms_consent"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
}

type WaiverWaiver struct {
	ID         uuid.UUID `json:"id"`
	WaiverUrl  string    `json:"waiver_url"`
	WaiverName string    `json:"waiver_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WaiverWaiverSigning struct {
	UserID    uuid.UUID `json:"user_id"`
	WaiverID  uuid.UUID `json:"waiver_id"`
	IsSigned  bool      `json:"is_signed"`
	UpdatedAt time.Time `json:"updated_at"`
}
