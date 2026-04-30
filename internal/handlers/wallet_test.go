package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/timur-makarov/sheepy-tt-go-wallet/api"
	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/app/repositories"
	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/services"
)

const testMnemonic = "test test test test test test test test test test test junk"

func testWalletHandler(t *testing.T) WalletHandler {
	t.Helper()

	repo, err := repositories.NewWalletCoreRepository("ethereum", testMnemonic)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(repo.Close)

	service := services.NewWalletService(map[string]*repositories.WalletCoreRepository{
		"ethereum": repo,
	})

	return WalletHandler{Service: service}
}

func TestWalletHandlerCreateAddress(t *testing.T) {
	handler := testWalletHandler(t)

	response, err := handler.CreateAddress(context.Background(), api.CreateAddressRequestObject{
		Body: &api.CreateAddressJSONRequestBody{
			Gate:         "ethereum",
			Account:      0,
			Change:       0,
			AddressIndex: 0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ok, isOK := response.(api.CreateAddress200JSONResponse)
	if !isOK {
		t.Fatalf("expected success response, got %T", response)
	}
	if ok.Address != "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" {
		t.Fatalf("unexpected address: %s", ok.Address)
	}
}

func TestWalletHandlerValidateAddress(t *testing.T) {
	handler := testWalletHandler(t)

	response, err := handler.ValidateAddress(context.Background(), api.ValidateAddressRequestObject{
		Body: &api.ValidateAddressJSONRequestBody{
			Gate:    "ethereum",
			Address: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ok, isOK := response.(api.ValidateAddress200JSONResponse)
	if !isOK {
		t.Fatalf("expected success response, got %T", response)
	}
	if !ok.Valid {
		t.Fatal("expected address to be valid")
	}
}

func TestWalletHandlerSignTransaction(t *testing.T) {
	handler := testWalletHandler(t)

	response, err := handler.SignTransaction(context.Background(), api.SignTransactionRequestObject{
		Body: &api.SignTransactionJSONRequestBody{
			Gate:         "ethereum",
			Account:      0,
			Change:       0,
			AddressIndex: 0,
			TxParams: api.TxParams{
				To:                      "0x000000000000000000000000000000000000dEaD",
				ValueWei:                "1",
				Data:                    "0x",
				Nonce:                   0,
				ChainId:                 11155111,
				GasLimit:                21000,
				MaxFeePerGasWei:         "30000000000",
				MaxPriorityFeePerGasWei: "1500000000",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ok, isOK := response.(api.SignTransaction200JSONResponse)
	if !isOK {
		t.Fatalf("expected success response, got %T", response)
	}
	if !strings.HasPrefix(ok.TxHash, "0x") || len(ok.TxHash) != 66 {
		t.Fatalf("unexpected tx hash: %s", ok.TxHash)
	}
	if !strings.HasPrefix(ok.SignedTx, "0x02") {
		t.Fatalf("expected signed EIP-1559 transaction, got: %s", ok.SignedTx)
	}
}

func TestWalletHandlerReturnsErrorResponse(t *testing.T) {
	handler := testWalletHandler(t)

	response, err := handler.CreateAddress(context.Background(), api.CreateAddressRequestObject{
		Body: &api.CreateAddressJSONRequestBody{
			Gate:         "unknown",
			Account:      0,
			Change:       0,
			AddressIndex: 0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	defaultResponse, isDefault := response.(api.CreateAddressdefaultJSONResponse)
	if !isDefault {
		t.Fatalf("expected default error response, got %T", response)
	}
	if defaultResponse.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", defaultResponse.StatusCode)
	}
	if defaultResponse.Body.Error.Message == "" {
		t.Fatal("expected error message")
	}
}
