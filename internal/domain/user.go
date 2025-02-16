package domain

import (
	"fmt"

	"github.com/netscrawler/avito-shop/internal/model"
)

// User представляет пользователя в системе
type User struct {
	Id        int64           // Идентификатор пользователя
	Username  string          // Имя пользователя
	Password  []byte          // Хэш пароля
	Coins     uint64          // Количество монет
	Inventory []UserInventory // Инвентарь пользователя
}

// UserInventory представляет предмет в инвентаре пользователя
type UserInventory = model.Item

// NewUser создает нового пользователя
func NewUser(username string, password []byte, coins uint64) *User {
	return &User{
		Username: username,
		Password: password,
		Coins:    coins,
	}
}

// AddItem добавляет предмет в инвентарь пользователя
func (u *User) AddItem(item string, quantity int) {
	for i, inv := range u.Inventory {
		if inv.Type == item {
			u.Inventory[i].Quantity += quantity
			return
		}
	}
	u.Inventory = append(u.Inventory, UserInventory{Type: item, Quantity: quantity})
}

// AddCoins добавляет монеты пользователю
func (u *User) AddCoins(amount uint64) error {
	// Проверяем на переполнение
	if amount > (^uint64(0) - u.Coins) {
		return fmt.Errorf("переполнение баланса: текущий баланс %d, добавляемая сумма %d", u.Coins, amount)
	}
	u.Coins += amount
	return nil
}

// SubtractCoins вычитает монеты у пользователя
// Возвращает ошибку, если недостаточно монет
func (u *User) SubtractCoins(amount uint64) error {
	if u.Coins < amount {
		return fmt.Errorf("недостаточно монет: имеется %d, требуется %d", u.Coins, amount)
	}
	u.Coins -= amount
	return nil
}

// HasEnoughCoins проверяет, достаточно ли монет у пользователя
func (u *User) HasEnoughCoins(amount uint64) bool {
	return u.Coins >= amount
}
