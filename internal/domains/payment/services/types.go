package payment

type RecurringPaymentInterval string

const (
	IntervalDay   RecurringPaymentInterval = "day"
	IntervalWeek  RecurringPaymentInterval = "week"
	IntervalMonth RecurringPaymentInterval = "month"
	IntervalYear  RecurringPaymentInterval = "year"
)

func IsIntervalValid(interval RecurringPaymentInterval) bool {
	switch interval {
	case IntervalDay, IntervalWeek, IntervalMonth, IntervalYear:
		return true
	default:
		return false
	}
}
