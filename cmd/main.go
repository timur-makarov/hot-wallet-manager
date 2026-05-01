package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/timur-makarov/hot-wallet-manager/internal/app"
	"github.com/timur-makarov/hot-wallet-manager/internal/config"
	"github.com/timur-makarov/hot-wallet-manager/internal/server"
)

func main() {
	configPath := os.Getenv("WALLET_CONFIG")
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := config.InitConfiguration(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	application, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	httpServer := server.InitServer(cfg, application)
	log.Printf("wallet service listening on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}
