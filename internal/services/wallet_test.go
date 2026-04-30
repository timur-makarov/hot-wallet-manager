package services

import (
	"strings"
	"testing"

	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/app/repositories"
)

const testMnemonic = "test test test test test test test test test test test junk"

func testService(t *testing.T) *WalletService {
	t.Helper()

	repo, err := repositories.NewWalletCoreRepository("ethereum", testMnemonic)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(repo.Close)

	return NewWalletService(map[string]*repositories.WalletCoreRepository{
		"ethereum": repo,
	})
}

func TestCreateAddress(t *testing.T) {
	service := testService(t)

	address, err := service.CreateAddress(DerivationRequest{
		Gate:         "ethereum",
		Account:      0,
		Change:       0,
		AddressIndex: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	if address != "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266" {
		t.Fatalf("unexpected address: %s", address)
	}
}

func TestValidateAddress(t *testing.T) {
	service := testService(t)

	valid, err := service.ValidateAddress("ethereum", "0xF39Fd6e51aad88F6F4ce6aB8827279cffFb92266")
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("expected address to be valid")
	}

	valid, err = service.ValidateAddress("ethereum", "not-an-address")
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Fatal("expected address to be invalid")
	}
}

func TestSignTransaction(t *testing.T) {
	service := testService(t)

	signed, err := service.SignTransaction(SignTxRequest{
		DerivationRequest: DerivationRequest{
			Gate:         "ethereum",
			Account:      0,
			Change:       0,
			AddressIndex: 0,
		},
		TxParams: TxParams{
			To:                      "0x000000000000000000000000000000000000dEaD",
			ValueWei:                "1",
			Data:                    "0x",
			Nonce:                   0,
			ChainID:                 11155111,
			GasLimit:                21000,
			MaxFeePerGasWei:         "30000000000",
			MaxPriorityFeePerGasWei: "1500000000",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(signed.TxHash, "0x") || len(signed.TxHash) != 66 {
		t.Fatalf("unexpected tx hash: %s", signed.TxHash)
	}
	if !strings.HasPrefix(signed.SignedTx, "0x02") {
		t.Fatalf("expected signed EIP-1559 transaction, got: %s", signed.SignedTx)
	}
}
