package app

import (
	"fmt"

	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/app/repositories"
	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/config"
)

// Repositories aggregates per-gate WalletCoreRepository instances.
// One repository per configured gate, keyed by gate name.
type Repositories struct {
	Wallets map[string]*repositories.WalletCoreRepository
}

// GetRepositories instantiates one wallet-core repository per configured
// gate. Mnemonic validity is checked here (via wallet-core), so a bad
// mnemonic in config aborts startup with a clear error.
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

// Close releases every underlying TWHDWallet. Idempotent.
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
