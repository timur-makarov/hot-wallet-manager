package app

import (
	"log/slog"

	"github.com/timur-makarov/hot-wallet-manager/internal/config"
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
