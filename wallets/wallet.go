package wallets

import (
	"github.com/shopspring/decimal"
	"github.com/wythers/algalon/provider"
)

type TRCWallet struct {
	// atomic variables
	Sqno uint64

	W     provider.Wallet
	Owner string

	Trx   decimal.Decimal
	Trc20 decimal.Decimal
}
