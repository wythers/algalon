package provider

import (
	"github.com/shopspring/decimal"
)

type Providerer interface {
	GenerateWallet() (Wallet, error)

	GetTRXBalance(string, []string) (decimal.Decimal, error)

	SendTRX(Wallet, string, decimal.Decimal, []string) (string, error)

	GetTRC20Balance(string, []string) (decimal.Decimal, error)

	SendTRC20(Wallet, string, decimal.Decimal, []string) (string, error)
}

type Wallet struct {
	PrivKey string `json:"private_key"`

	Address string `json:"address"`
}
