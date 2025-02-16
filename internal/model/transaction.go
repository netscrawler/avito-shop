package model

// SendCoinRequest используется для отправки монет другому пользователю.
type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount uint64 `json:"amount"`
}
