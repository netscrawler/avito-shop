package domain

import "time"

// TransactionType определяет тип транзакции
type TransactionType string

const (
	// TransactionTypePurchase представляет покупку товара
	TransactionTypePurchase TransactionType = "PURCHASE"
	// TransactionTypeTransfer представляет перевод монет между пользователями
	TransactionTypeTransfer TransactionType = "TRANSFER"
)

// Transaction представляет транзакцию в системе
type Transaction struct {
	SenderName   string          // Имя отправителя
	ReceiverName string          // Имя получателя
	Amount       uint64          // Сумма транзакции
	Type         TransactionType // Тип транзакции
	Timestamp    time.Time       // Время транзакции
}

// NewTransaction создает новую транзакцию
func NewTransaction(senderName, receiverName string, amount uint64, typeTransaction TransactionType, timestamp time.Time) *Transaction {
	return &Transaction{
		SenderName:   senderName,
		ReceiverName: receiverName,
		Amount:       amount,
		Type:         typeTransaction,
		Timestamp:    timestamp,
	}
}
