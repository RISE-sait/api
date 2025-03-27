package payment

type Frequency string

const (
	Day      Frequency = "day"
	Week     Frequency = "week"
	Biweekly Frequency = "biweekly"
	Month    Frequency = "month"
)

func IsFrequencyValid(frequency Frequency) bool {
	switch frequency {
	case Day, Week, Biweekly, Month:
		return true
	default:
		return false
	}
}
