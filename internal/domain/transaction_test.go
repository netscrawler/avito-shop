package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTransaction(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		senderName   string
		receiverName string
		amount       uint64
		txType       TransactionType
		timestamp    time.Time
		want         *Transaction
	}{
		{
			name:         "создание транзакции перевода",
			senderName:   "sender",
			receiverName: "receiver",
			amount:       100,
			txType:       TransactionTypeTransfer,
			timestamp:    now,
			want: &Transaction{
				SenderName:   "sender",
				ReceiverName: "receiver",
				Amount:       100,
				Type:         TransactionTypeTransfer,
				Timestamp:    now,
			},
		},
		{
			name:         "создание транзакции покупки",
			senderName:   "buyer",
			receiverName: "SHOP",
			amount:       500,
			txType:       TransactionTypePurchase,
			timestamp:    now,
			want: &Transaction{
				SenderName:   "buyer",
				ReceiverName: "SHOP",
				Amount:       500,
				Type:         TransactionTypePurchase,
				Timestamp:    now,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTransaction(tt.senderName, tt.receiverName, tt.amount, tt.txType, tt.timestamp)
			assert.Equal(t, tt.want, got)
		})
	}
}
