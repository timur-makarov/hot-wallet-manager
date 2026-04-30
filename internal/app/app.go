package app

import (
	"log/slog"

	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/config"
)

type App struct {
	Handlers     *Handlers
	services     *Services
	repositories *Repositories

	Logger *slog.Logger
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		Logger: slog.Default(),
	}

	repos, err := GetRepositories(cfg)
	if err != nil {
		return nil, err
	}
	app.repositories = repos

	app.services = GetServices(repos)
	app.Handlers = GetHandlers(app.services)

	return app, nil
}

// Close releases external resources (currently the per-gate wallet-core
// HD wallet handles). Safe to call multiple times.
func (a *App) Close() {
	if a == nil {
		return
	}
	a.repositories.Close()
}
