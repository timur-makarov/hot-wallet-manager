package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/timur-makarov/hot-wallet-manager/api"
	"github.com/timur-makarov/hot-wallet-manager/internal/app"
	"github.com/timur-makarov/hot-wallet-manager/internal/config"
	"github.com/timur-makarov/hot-wallet-manager/internal/transport"
)

func InitServer(cfg *config.Config, a *app.App) *http.Server {
	handler := RegisterEndpoints(a)

	return &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func RegisterEndpoints(a *app.App) http.Handler {
	strictHandler := api.NewStrictHandlerWithOptions(
		a.Handlers.WalletHandler,
		nil,
		api.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  writeRequestError,
			ResponseErrorHandlerFunc: writeResponseError,
		},
	)

	return api.HandlerWithOptions(strictHandler, api.StdHTTPServerOptions{
		BaseURL: "/api/v1",
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeError(w, transport.WrapBadRequest(err, "invalid request"))
		},
	})
}

func writeRequestError(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, transport.WrapBadRequest(err, "invalid request"))
}

func writeResponseError(w http.ResponseWriter, _ *http.Request, err error) {
	writeError(w, transport.WrapInternal(err, "internal server error"))
}

func writeError(w http.ResponseWriter, err error) {
	apiErr := transport.FromError(err)
	response := api.ErrorResponse{}
	response.Error.Message = apiErr.Message

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	_ = json.NewEncoder(w).Encode(response)
}
