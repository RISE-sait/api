package purchase

//type SquareWebhookEventDto struct {
//	Type string             `json:"type"`
//	Data SquareEventDataDto `json:"data"`
//}

//type SquareEventDataDto struct {
//	ID     string          `json:"id"`
//	Object json.RawMessage `json:"object"`
//}

type SquarePaymentDto struct {
	ID          string         `json:"id"`
	Status      string         `json:"status"`
	AmountMoney SquareMoneyDto `json:"amount_money"`
	OrderID     string         `json:"order_id"`
	ReceiptNo   string         `json:"receipt_number"`
}

type SquareMoneyDto struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}
