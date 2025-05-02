package internal

type KafkaMessage struct {
	TransactionID string  `json:"transaction_id"`
	AccountID     int64   `json:"account_id"`
	AmountChange  float64 `json:"amount_change"`
}
