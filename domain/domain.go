package domain

type OrderRequest struct {
	UserId string `json:"user_id,omitempty"`
	Amount uint   `json:"number,omitempty"`
}
