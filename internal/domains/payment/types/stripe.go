package stripe

type PaymentFrequency string

const (
	Day      PaymentFrequency = "day"
	Week     PaymentFrequency = "week"
	Biweekly PaymentFrequency = "biweekly"
	Month    PaymentFrequency = "month"
)

func IsPaymentFrequencyValid(frequency PaymentFrequency) bool {
	switch frequency {
	case Day, Week, Biweekly, Month:
		return true
	default:
		return false
	}
}
