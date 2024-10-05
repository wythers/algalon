package utils

import (
	"errors"
)

var (
	ErrSuspended      = errors.New("algalon: the current service has been suspended")
	ErrEOF            = errors.New("algalon: at EOF")
	ErrAppKeyInvalid  = errors.New("algalon: invalid algalon app key")
	ErrClosed         = errors.New("algalon: the service has been closed")
	ErrGenerateWallet = errors.New("algalon: generate wallet failed")
	ErrTimeOut        = errors.New("algalon: timeout")
)

type Writer interface {
	Write(record Record) error
}

type AppInfo struct {
	Threshold string `json:"threshold"`
	// Minimum transaction amount
	Mta string `json:"mta"`

	Address    []string `json:"address"`
	TronApiKey []string `json:"tron_api_key"`
}
