package responses

// BalanceResponse is the success payload for GET /users/:user_id/balance
type BalanceResponse struct {
	Balance      float64 `json:"balance"`
	TotalDebits  float64 `json:"total_debits"`
	TotalCredits float64 `json:"total_credits"`
}


