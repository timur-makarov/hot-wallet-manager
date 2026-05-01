package app

import (
	"github.com/timur-makarov/hot-wallet-manager/internal/services"
)

type Services struct {
	WalletService *services.WalletService
}

func GetServices(repos *Repositories) *Services {
	return &Services{
		WalletService: services.NewWalletService(repos.Wallets),
	}
}
