package app

import "github.com/timur-makarov/hot-wallet-manager/internal/handlers"

type Handlers struct {
	WalletHandler handlers.WalletHandler
}

func GetHandlers(services *Services) *Handlers {
	return &Handlers{
		WalletHandler: handlers.WalletHandler{Service: services.WalletService},
	}
}
