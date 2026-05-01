package services

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/timur-makarov/hot-wallet-manager/internal/app/repositories"
	"github.com/timur-makarov/hot-wallet-manager/internal/transport"
)

type WalletService struct {
	wallets map[string]*repositories.WalletCoreRepository
}

type DerivationRequest struct {
	Gate         string `json:"gate"`
	Account      uint32 `json:"account"`
	Change       uint32 `json:"change"`
	AddressIndex uint32 `json:"address_index"`
}

type TxParams struct {
	To                      string `json:"to"`
	ValueWei                string `json:"value_wei"`
	Data                    string `json:"data"`
	Nonce                   uint64 `json:"nonce"`
	ChainID                 uint64 `json:"chain_id"`
	GasLimit                uint64 `json:"gas_limit"`
	MaxFeePerGasWei         string `json:"max_fee_per_gas_wei"`
	MaxPriorityFeePerGasWei string `json:"max_priority_fee_per_gas_wei"`
}

type SignTxRequest struct {
	DerivationRequest
	TxParams TxParams `json:"tx_params"`
}

type SignedTx struct {
	TxHash   string `json:"tx_hash"`
	SignedTx string `json:"signed_tx"`
}

func NewWalletService(wallets map[string]*repositories.WalletCoreRepository) *WalletService {
	return &WalletService{wallets: wallets}
}

func (s *WalletService) CreateAddress(req DerivationRequest) (string, error) {
	repo, err := s.repoForGate(req.Gate)
	if err != nil {
		return "", err
	}
	if err := validateDerivation(req); err != nil {
		return "", err
	}

	address, err := repo.DeriveEthereumAddress(req.Account, req.Change, req.AddressIndex)
	if err != nil {
		return "", transport.WrapBadRequest(err, "failed to derive address")
	}
	return address, nil
}

func (s *WalletService) ValidateAddress(gate string, address string) (bool, error) {
	if _, err := s.repoForGate(gate); err != nil {
		return false, err
	}
	return repositories.IsValidEthereumAddress(address), nil
}

func (s *WalletService) SignTransaction(req SignTxRequest) (*SignedTx, error) {
	repo, err := s.repoForGate(req.Gate)
	if err != nil {
		return nil, err
	}
	if err := validateDerivation(req.DerivationRequest); err != nil {
		return nil, err
	}

	params, err := normalizeTxParams(req.TxParams)
	if err != nil {
		return nil, err
	}

	signed, err := repo.SignEthereumTransaction(
		req.Account,
		req.Change,
		req.AddressIndex,
		repositories.EthereumTxParams{
			To:                      req.TxParams.To,
			ValueWei:                params.value,
			Data:                    params.data,
			Nonce:                   req.TxParams.Nonce,
			ChainID:                 params.chainID,
			GasLimit:                req.TxParams.GasLimit,
			MaxFeePerGasWei:         params.maxFeePerGas,
			MaxPriorityFeePerGasWei: params.maxPriorityFeePerGas,
		},
	)
	if err != nil {
		return nil, transport.WrapBadRequest(err, "failed to sign transaction")
	}

	return &SignedTx{
		TxHash:   signed.TxHash,
		SignedTx: "0x" + hex.EncodeToString(signed.Encoded),
	}, nil
}

func (s *WalletService) repoForGate(gate string) (*repositories.WalletCoreRepository, error) {
	repo, ok := s.wallets[gate]
	if !ok {
		return nil, transport.NotFound("gate %q is not configured", gate)
	}
	return repo, nil
}

func validateDerivation(req DerivationRequest) error {
	if req.Gate == "" {
		return transport.BadRequest("gate is required")
	}
	if req.Change > 1 {
		return transport.BadRequest("change must be 0 or 1")
	}
	return nil
}

type normalizedTxParams struct {
	value                *big.Int
	data                 []byte
	chainID              *big.Int
	maxFeePerGas         *big.Int
	maxPriorityFeePerGas *big.Int
}

func normalizeTxParams(params TxParams) (*normalizedTxParams, error) {
	if !repositories.IsValidEthereumAddress(params.To) {
		return nil, transport.BadRequest("tx_params.to must be a valid Ethereum address")
	}

	value, err := parseDecimalBigInt("tx_params.value_wei", params.ValueWei, false)
	if err != nil {
		return nil, err
	}
	maxFeePerGas, err := parseDecimalBigInt("tx_params.max_fee_per_gas_wei", params.MaxFeePerGasWei, true)
	if err != nil {
		return nil, err
	}
	maxPriorityFeePerGas, err := parseDecimalBigInt("tx_params.max_priority_fee_per_gas_wei", params.MaxPriorityFeePerGasWei, true)
	if err != nil {
		return nil, err
	}
	if maxPriorityFeePerGas.Cmp(maxFeePerGas) > 0 {
		return nil, transport.BadRequest("tx_params.max_priority_fee_per_gas_wei cannot exceed max_fee_per_gas_wei")
	}

	data, err := parseHexData(params.Data)
	if err != nil {
		return nil, err
	}

	return &normalizedTxParams{
		value:                value,
		data:                 data,
		chainID:              big.NewInt(0).SetUint64(params.ChainID),
		maxFeePerGas:         maxFeePerGas,
		maxPriorityFeePerGas: maxPriorityFeePerGas,
	}, nil
}

func parseDecimalBigInt(field string, raw string, requirePositive bool) (*big.Int, error) {
	if raw == "" {
		return nil, transport.BadRequest("%s is required", field)
	}
	if strings.HasPrefix(raw, "-") || strings.ContainsAny(raw, ".eE") {
		return nil, transport.BadRequest("%s must be a base-10 integer string", field)
	}

	value, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, transport.BadRequest("%s must be a base-10 integer string", field)
	}
	if value.Sign() < 0 || (requirePositive && value.Sign() == 0) {
		return nil, transport.BadRequest("%s must be positive", field)
	}

	return value, nil
}

func parseHexData(raw string) ([]byte, error) {
	if raw == "" {
		return nil, transport.BadRequest("tx_params.data is required")
	}
	if !strings.HasPrefix(raw, "0x") {
		return nil, transport.BadRequest("tx_params.data must start with 0x")
	}

	payload := strings.TrimPrefix(raw, "0x")
	if payload == "" {
		return []byte{}, nil
	}
	if len(payload)%2 != 0 {
		return nil, transport.BadRequest("tx_params.data must have an even number of hex characters")
	}

	data, err := hex.DecodeString(payload)
	if err != nil {
		return nil, transport.BadRequest("tx_params.data must be valid hex")
	}
	return data, nil
}
