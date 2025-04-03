// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db_seed

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AuditAuditStatus string

const (
	AuditAuditStatusPENDING   AuditAuditStatus = "PENDING"
	AuditAuditStatusCOMPLETED AuditAuditStatus = "COMPLETED"
	AuditAuditStatusFAILED    AuditAuditStatus = "FAILED"
)

func (e *AuditAuditStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = AuditAuditStatus(s)
	case string:
		*e = AuditAuditStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for AuditAuditStatus: %T", src)
	}
	return nil
}

type NullAuditAuditStatus struct {
	AuditAuditStatus AuditAuditStatus `json:"audit_audit_status"`
	Valid            bool             `json:"valid"` // Valid is true if AuditAuditStatus is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullAuditAuditStatus) Scan(value interface{}) error {
	if value == nil {
		ns.AuditAuditStatus, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.AuditAuditStatus.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullAuditAuditStatus) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.AuditAuditStatus), nil
}

func (e AuditAuditStatus) Valid() bool {
	switch e {
	case AuditAuditStatusPENDING,
		AuditAuditStatusCOMPLETED,
		AuditAuditStatusFAILED:
		return true
	}
	return false
}

func AllAuditAuditStatusValues() []AuditAuditStatus {
	return []AuditAuditStatus{
		AuditAuditStatusPENDING,
		AuditAuditStatusCOMPLETED,
		AuditAuditStatusFAILED,
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
	PaymentFrequencyOnce     PaymentFrequency = "once"
	PaymentFrequencyWeek     PaymentFrequency = "week"
	PaymentFrequencyBiweekly PaymentFrequency = "biweekly"
	PaymentFrequencyMonth    PaymentFrequency = "month"
	PaymentFrequencyDay      PaymentFrequency = "day"
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
		PaymentFrequencyBiweekly,
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
		PaymentFrequencyBiweekly,
		PaymentFrequencyMonth,
		PaymentFrequencyDay,
	}
}

type ProgramProgramLevel string

const (
	ProgramProgramLevelBeginner     ProgramProgramLevel = "beginner"
	ProgramProgramLevelIntermediate ProgramProgramLevel = "intermediate"
	ProgramProgramLevelAdvanced     ProgramProgramLevel = "advanced"
	ProgramProgramLevelAll          ProgramProgramLevel = "all"
)

func (e *ProgramProgramLevel) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ProgramProgramLevel(s)
	case string:
		*e = ProgramProgramLevel(s)
	default:
		return fmt.Errorf("unsupported scan type for ProgramProgramLevel: %T", src)
	}
	return nil
}

type NullProgramProgramLevel struct {
	ProgramProgramLevel ProgramProgramLevel `json:"program_program_level"`
	Valid               bool                `json:"valid"` // Valid is true if ProgramProgramLevel is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullProgramProgramLevel) Scan(value interface{}) error {
	if value == nil {
		ns.ProgramProgramLevel, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ProgramProgramLevel.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullProgramProgramLevel) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ProgramProgramLevel), nil
}

func (e ProgramProgramLevel) Valid() bool {
	switch e {
	case ProgramProgramLevelBeginner,
		ProgramProgramLevelIntermediate,
		ProgramProgramLevelAdvanced,
		ProgramProgramLevelAll:
		return true
	}
	return false
}

func AllProgramProgramLevelValues() []ProgramProgramLevel {
	return []ProgramProgramLevel{
		ProgramProgramLevelBeginner,
		ProgramProgramLevelIntermediate,
		ProgramProgramLevelAdvanced,
		ProgramProgramLevelAll,
	}
}

type ProgramProgramType string

const (
	ProgramProgramTypePractice ProgramProgramType = "practice"
	ProgramProgramTypeCourse   ProgramProgramType = "course"
	ProgramProgramTypeGame     ProgramProgramType = "game"
	ProgramProgramTypeOthers   ProgramProgramType = "others"
)

func (e *ProgramProgramType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = ProgramProgramType(s)
	case string:
		*e = ProgramProgramType(s)
	default:
		return fmt.Errorf("unsupported scan type for ProgramProgramType: %T", src)
	}
	return nil
}

type NullProgramProgramType struct {
	ProgramProgramType ProgramProgramType `json:"program_program_type"`
	Valid              bool               `json:"valid"` // Valid is true if ProgramProgramType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullProgramProgramType) Scan(value interface{}) error {
	if value == nil {
		ns.ProgramProgramType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ProgramProgramType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullProgramProgramType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ProgramProgramType), nil
}

func (e ProgramProgramType) Valid() bool {
	switch e {
	case ProgramProgramTypePractice,
		ProgramProgramTypeCourse,
		ProgramProgramTypeGame,
		ProgramProgramTypeOthers:
		return true
	}
	return false
}

func AllProgramProgramTypeValues() []ProgramProgramType {
	return []ProgramProgramType{
		ProgramProgramTypePractice,
		ProgramProgramTypeCourse,
		ProgramProgramTypeGame,
		ProgramProgramTypeOthers,
	}
}

type AthleticAthlete struct {
	ID        uuid.UUID     `json:"id"`
	Wins      int32         `json:"wins"`
	Losses    int32         `json:"losses"`
	Points    int32         `json:"points"`
	Steals    int32         `json:"steals"`
	Assists   int32         `json:"assists"`
	Rebounds  int32         `json:"rebounds"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	TeamID    uuid.NullUUID `json:"team_id"`
}

type AthleticCoachStat struct {
	ID        uuid.UUID `json:"id"`
	Wins      int32     `json:"wins"`
	Losses    int32     `json:"losses"`
	CoachID   uuid.UUID `json:"coach_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AthleticTeam struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Capacity  int32         `json:"capacity"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	CoachID   uuid.NullUUID `json:"coach_id"`
}

type AuditOutbox struct {
	ID           uuid.UUID        `json:"id"`
	SqlStatement string           `json:"sql_statement"`
	Status       AuditAuditStatus `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
}

type CustomerDiscountUsage struct {
	CustomerID uuid.UUID `json:"customer_id"`
	DiscountID uuid.UUID `json:"discount_id"`
	UsageCount int32     `json:"usage_count"`
	LastUsedAt time.Time `json:"last_used_at"`
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

type EnrollmentFee struct {
	ID           uuid.UUID      `json:"id"`
	ProgramID    uuid.UUID      `json:"program_id"`
	MembershipID uuid.NullUUID  `json:"membership_id"`
	DropInPrice  sql.NullString `json:"drop_in_price"`
	ProgramPrice sql.NullString `json:"program_price"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type EventsAttendance struct {
	ID          uuid.UUID    `json:"id"`
	EventID     uuid.UUID    `json:"event_id"`
	UserID      uuid.UUID    `json:"user_id"`
	CheckInTime sql.NullTime `json:"check_in_time"`
}

type EventsCustomerEnrollment struct {
	ID          uuid.UUID    `json:"id"`
	CustomerID  uuid.UUID    `json:"customer_id"`
	EventID     uuid.UUID    `json:"event_id"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CheckedInAt sql.NullTime `json:"checked_in_at"`
	IsCancelled bool         `json:"is_cancelled"`
}

type EventsEvent struct {
	ID                 uuid.UUID      `json:"id"`
	LocationID         uuid.UUID      `json:"location_id"`
	ProgramID          uuid.NullUUID  `json:"program_id"`
	TeamID             uuid.NullUUID  `json:"team_id"`
	StartAt            time.Time      `json:"start_at"`
	EndAt              time.Time      `json:"end_at"`
	CreatedBy          uuid.UUID      `json:"created_by"`
	UpdatedBy          uuid.UUID      `json:"updated_by"`
	Capacity           sql.NullInt32  `json:"capacity"`
	IsCancelled        bool           `json:"is_cancelled"`
	CancellationReason sql.NullString `json:"cancellation_reason"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
}

type EventsStaff struct {
	EventID uuid.UUID `json:"event_id"`
	StaffID uuid.UUID `json:"staff_id"`
}

type HaircutBarberService struct {
	ID        uuid.UUID `json:"id"`
	BarberID  uuid.UUID `json:"barber_id"`
	ServiceID uuid.UUID `json:"service_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HaircutEvent struct {
	ID            uuid.UUID `json:"id"`
	BeginDateTime time.Time `json:"begin_date_time"`
	EndDateTime   time.Time `json:"end_date_time"`
	CustomerID    uuid.UUID `json:"customer_id"`
	BarberID      uuid.UUID `json:"barber_id"`
	ServiceTypeID uuid.UUID `json:"service_type_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type HaircutHaircutService struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Description   sql.NullString  `json:"description"`
	Price         decimal.Decimal `json:"price"`
	DurationInMin int32           `json:"duration_in_min"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type LocationLocation struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MembershipMembership struct {
	ID          uuid.UUID      `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	Benefits    string         `json:"benefits"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type MembershipMembershipPlan struct {
	ID               uuid.UUID        `json:"id"`
	Name             string           `json:"name"`
	Price            decimal.Decimal  `json:"price"`
	JoiningFee       decimal.Decimal  `json:"joining_fee"`
	AutoRenew        bool             `json:"auto_renew"`
	MembershipID     uuid.UUID        `json:"membership_id"`
	PaymentFrequency PaymentFrequency `json:"payment_frequency"`
	AmtPeriods       sql.NullInt32    `json:"amt_periods"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type ProgramGame struct {
	ID        uuid.UUID `json:"id"`
	WinTeam   uuid.UUID `json:"win_team"`
	LoseTeam  uuid.UUID `json:"lose_team"`
	WinScore  int32     `json:"win_score"`
	LoseScore int32     `json:"lose_score"`
}

type ProgramProgram struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Level       ProgramProgramLevel `json:"level"`
	Type        ProgramProgramType  `json:"type"`
	Capacity    sql.NullInt32       `json:"capacity"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type StaffStaff struct {
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
}

type StaffStaffActivityLog struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Activity   string    `json:"activity"`
	OccurredAt time.Time `json:"occurred_at"`
}

type StaffStaffRole struct {
	ID        uuid.UUID `json:"id"`
	RoleName  string    `json:"role_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UsersCustomerCredit struct {
	CustomerID uuid.UUID `json:"customer_id"`
	Credits    int32     `json:"credits"`
}

type UsersUser struct {
	ID                       uuid.UUID      `json:"id"`
	HubspotID                sql.NullString `json:"hubspot_id"`
	CountryAlpha2Code        string         `json:"country_alpha2_code"`
	Gender                   sql.NullString `json:"gender"`
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
