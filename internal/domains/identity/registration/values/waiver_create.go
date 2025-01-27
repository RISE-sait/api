package values

type WaiverCreate struct {
	Email     string
	WaiverUrl string
	IsSigned  bool
}

func NewWaiverCreate(email string, waiverUrl string, isSigned bool) *WaiverCreate {
	return &WaiverCreate{
		Email:     email,
		WaiverUrl: waiverUrl,
		IsSigned:  isSigned,
	}
}
