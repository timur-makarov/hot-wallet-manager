package app

import (
	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/services"
)

type Services struct {
	WalletService *services.WalletService
}

func GetServices(repos *Repositories) *Services {
	return &Services{
		WalletService: services.NewWalletService(repos.Wallets),
	}
}
