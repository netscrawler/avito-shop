package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddItem(t *testing.T) {
	user := &User{
		Username: "testuser",
		Inventory: []UserInventory{
			{Type: "item1", Quantity: 1},
		},
	}

	tests := []struct {
		name          string
		itemType      string
		quantity      int
		wantInventory []UserInventory
	}{
		{
			name:     "добавление нового предмета",
			itemType: "item2",
			quantity: 1,
			wantInventory: []UserInventory{
				{Type: "item1", Quantity: 1},
				{Type: "item2", Quantity: 1},
			},
		},
		{
			name:     "добавление существующего предмета",
			itemType: "item1",
			quantity: 2,
			wantInventory: []UserInventory{
				{Type: "item1", Quantity: 3},
				{Type: "item2", Quantity: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user.AddItem(tt.itemType, tt.quantity)
			assert.Equal(t, tt.wantInventory, user.Inventory)
		})
	}
}

func TestSubtractCoins(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		amount    uint64
		wantCoins uint64
		wantErr   error
	}{
		{
			name: "успешное вычитание монет",
			user: &User{
				Username: "testuser",
				Coins:    1000,
			},
			amount:    500,
			wantCoins: 500,
			wantErr:   nil,
		},
		{
			name: "недостаточно монет",
			user: &User{
				Username: "testuser",
				Coins:    100,
			},
			amount:    500,
			wantCoins: 100,
			wantErr:   fmt.Errorf("недостаточно монет: имеется %d, требуется %d", 100, 500),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.SubtractCoins(tt.amount)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCoins, tt.user.Coins)
		})
	}
}

func TestHasEnoughCoins(t *testing.T) {
	tests := []struct {
		name   string
		user   *User
		amount uint64
		want   bool
	}{
		{
			name: "достаточно монет",
			user: &User{
				Username: "testuser",
				Coins:    1000,
			},
			amount: 500,
			want:   true,
		},
		{
			name: "недостаточно монет",
			user: &User{
				Username: "testuser",
				Coins:    100,
			},
			amount: 500,
			want:   false,
		},
		{
			name: "ровно столько монет",
			user: &User{
				Username: "testuser",
				Coins:    500,
			},
			amount: 500,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.user.HasEnoughCoins(tt.amount)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAddCoins(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		amount    uint64
		wantCoins uint64
		wantErr   error
	}{
		{
			name: "успешное добавление монет",
			user: &User{
				Username: "testuser",
				Coins:    1000,
			},
			amount:    500,
			wantCoins: 1500,
			wantErr:   nil,
		},
		{
			name: "переполнение баланса",
			user: &User{
				Username: "testuser",
				Coins:    ^uint64(0) - 100,
			},
			amount:    200,
			wantCoins: ^uint64(0) - 100,
			wantErr:   fmt.Errorf("переполнение баланса: текущий баланс %d, добавляемая сумма %d", ^uint64(0)-100, 200),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.AddCoins(tt.amount)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCoins, tt.user.Coins)
		})
	}
}
