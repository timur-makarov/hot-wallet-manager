package handlers

import (
	"context"

	"github.com/timur-makarov/sheepy-tt-go-wallet/api"
	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/services"
	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/transport"
)

type WalletHandler struct {
	Service *services.WalletService
}

func (h WalletHandler) CreateAddress(
	_ context.Context,
	request api.CreateAddressRequestObject,
) (api.CreateAddressResponseObject, error) {
	req, err := mapDerivationInput(request.Body)
	if err != nil {
		return createAddressError(err), nil
	}

	address, err := h.Service.CreateAddress(req)
	if err != nil {
		return createAddressError(err), nil
	}

	return api.CreateAddress200JSONResponse{Address: address}, nil
}

func (h WalletHandler) ValidateAddress(
	_ context.Context,
	request api.ValidateAddressRequestObject,
) (api.ValidateAddressResponseObject, error) {
	if request.Body == nil {
		return validateAddressError(transport.BadRequest("request body is required")), nil
	}

	valid, err := h.Service.ValidateAddress(request.Body.Gate, request.Body.Address)
	if err != nil {
		return validateAddressError(err), nil
	}

	return api.ValidateAddress200JSONResponse{Valid: valid}, nil
}

func (h WalletHandler) SignTransaction(
	_ context.Context,
	request api.SignTransactionRequestObject,
) (api.SignTransactionResponseObject, error) {
	req, err := mapTxInput(request.Body)
	if err != nil {
		return signTransactionError(err), nil
	}

	tx, err := h.Service.SignTransaction(req)
	if err != nil {
		return signTransactionError(err), nil
	}

	return api.SignTransaction200JSONResponse{
		SignedTx: tx.SignedTx,
		TxHash:   tx.TxHash,
	}, nil
}

func mapDerivationInput(input *api.DerivationInput) (services.DerivationRequest, error) {
	if input == nil {
		return services.DerivationRequest{}, transport.BadRequest("request body is required")
	}

	return services.DerivationRequest{
		Gate:         input.Gate,
		Account:      input.Account,
		Change:       input.Change,
		AddressIndex: input.AddressIndex,
	}, nil
}

func mapTxInput(input *api.TxInput) (services.SignTxRequest, error) {
	if input == nil {
		return services.SignTxRequest{}, transport.BadRequest("request body is required")
	}

	derivation, err := mapDerivationInput(&api.DerivationInput{
		Account:      input.Account,
		AddressIndex: input.AddressIndex,
		Change:       input.Change,
		Gate:         input.Gate,
	})
	if err != nil {
		return services.SignTxRequest{}, err
	}

	return services.SignTxRequest{
		DerivationRequest: derivation,
		TxParams: services.TxParams{
			To:                      input.TxParams.To,
			ValueWei:                input.TxParams.ValueWei,
			Data:                    input.TxParams.Data,
			Nonce:                   input.TxParams.Nonce,
			ChainID:                 input.TxParams.ChainId,
			GasLimit:                input.TxParams.GasLimit,
			MaxFeePerGasWei:         input.TxParams.MaxFeePerGasWei,
			MaxPriorityFeePerGasWei: input.TxParams.MaxPriorityFeePerGasWei,
		},
	}, nil
}

func createAddressError(err error) api.CreateAddressdefaultJSONResponse {
	code, response := apiError(err)
	return api.CreateAddressdefaultJSONResponse{StatusCode: code, Body: response}
}

func validateAddressError(err error) api.ValidateAddressdefaultJSONResponse {
	code, response := apiError(err)
	return api.ValidateAddressdefaultJSONResponse{StatusCode: code, Body: response}
}

func signTransactionError(err error) api.SignTransactiondefaultJSONResponse {
	code, response := apiError(err)
	return api.SignTransactiondefaultJSONResponse{StatusCode: code, Body: response}
}

func apiError(err error) (int, api.ErrorResponse) {
	apiErr := transport.FromError(err)

	response := api.ErrorResponse{}
	response.Error.Message = apiErr.Message

	return apiErr.Code, response
}
