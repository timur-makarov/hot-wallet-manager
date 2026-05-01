package app

import (
	"fmt"

	"github.com/timur-makarov/hot-wallet-manager/internal/config"
	"github.com/timur-makarov/hot-wallet-manager/internal/repositories"
)

type Repositories struct {
	Wallets map[string]*repositories.WalletCoreRepository
}

func GetRepositories(cfg *config.Config) (*Repositories, error) {
	wallets := make(map[string]*repositories.WalletCoreRepository, len(cfg.Gates))
	for _, gate := range cfg.Gates {
		repo, err := repositories.NewWalletCoreRepository(gate.Name, gate.Mnemonic)
		if err != nil {
			closeAll(wallets)
			return nil, fmt.Errorf("failed to load gate %q: %w", gate.Name, err)
		}
		wallets[gate.Name] = repo
	}
	return &Repositories{Wallets: wallets}, nil
}

func (r *Repositories) Close() {
	if r == nil {
		return
	}
	closeAll(r.Wallets)
}

func closeAll(wallets map[string]*repositories.WalletCoreRepository) {
	for _, w := range wallets {
		w.Close()
	}
}
