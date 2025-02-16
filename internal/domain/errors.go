package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrUserAlreadyExists  = errors.New("пользователь уже существует")
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrInsufficientFunds  = errors.New("недостаточно средств")
	ErrRecipientNotFound  = errors.New("получатель не найден")
	ErrSenderNotFound     = errors.New("отправитель не найден")
	ErrInvalidAmount      = errors.New("неверная сумма перевода")
	ErrTransactionFailed  = errors.New("ошибка выполнения транзакции")
	ErrMerchNotFound      = errors.New("товар не найден")
	ErrEmptyUserHistory   = errors.New("user history is empty")
)
