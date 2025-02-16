package model

type CoinHistory struct {
	Received []ReceivedTransaction `json:"received"`
	Sent     []SentTransaction     `json:"sent"`
}

type ReceivedTransaction struct {
	FromUser string `json:"fromUser"`
	Amount   uint64 `json:"amount"`
}

type SentTransaction struct {
	ToUser string `json:"toUser"`
	Amount uint64 `json:"amount"`
}
