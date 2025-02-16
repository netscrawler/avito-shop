package model

// InfoResponse представляет информацию о пользователе: баланс, инвентарь и историю транзакций.
type InfoResponse struct {
	Coins       uint64      `json:"coins"`
	Inventory   []Item      `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}
