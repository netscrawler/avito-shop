package model

// Item представляет предмет в инвентаре пользователя.
type Item struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

// BuyMerchRequest представляет запрос на покупку мерча
type BuyMerchRequest struct {
	MerchID uint64 `json:"merch_id"`
	Amount  uint64 `json:"amount"`
}
