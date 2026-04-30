package repositories

// #include <TrustWalletCore/TWHDWallet.h>
// #include <TrustWalletCore/TWPrivateKey.h>
// #include <TrustWalletCore/TWPublicKey.h>
// #include <TrustWalletCore/TWPublicKeyType.h>
// #include <TrustWalletCore/TWAnyAddress.h>
// #include <TrustWalletCore/TWAnySigner.h>
// #include <TrustWalletCore/TWCoinType.h>
// #include <TrustWalletCore/TWHash.h>
// #include <TrustWalletCore/TWString.h>
// #include <TrustWalletCore/TWData.h>
import "C"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/timur-makarov/sheepy-tt-go-wallet/internal/types"
	common "github.com/timur-makarov/sheepy-tt-go-wallet/proto/gen/common"
	ethereum "github.com/timur-makarov/sheepy-tt-go-wallet/proto/gen/ethereum"
)

// EthereumTxParams is the canonical, already-validated input for
// signing an EIP-1559 (type-2) Ethereum transaction.
//
// All wei amounts are represented as *big.Int to avoid string parsing here;
// it is the service layer's job to validate ranges and parse JSON.
type EthereumTxParams struct {
	To                      string
	ValueWei                *big.Int
	Data                    []byte
	Nonce                   uint64
	ChainID                 *big.Int
	GasLimit                uint64
	MaxFeePerGasWei         *big.Int
	MaxPriorityFeePerGasWei *big.Int
}

// EthereumSignedTx is the result of an offline EIP-1559 signing.
//
// Encoded already includes the 0x02 envelope byte and is ready to be
// broadcast as the payload of `eth_sendRawTransaction`.
type EthereumSignedTx struct {
	TxHash  string // 0x-prefixed hex of keccak256(Encoded)
	Encoded []byte
}

// WalletCoreRepository is a thin wrapper around a single Trust Wallet Core
// HD wallet handle (derived from a BIP39 mnemonic). One instance per gate.
//
// The wrapped *TWHDWallet is owned by this repository and must be released
// with Close. The repository is intended to be created at process start and
// reused for the lifetime of the service; it is safe for concurrent reads
// because TWHDWallet only exposes pure derivation/signing operations that
// do not mutate internal state.
type WalletCoreRepository struct {
	name   string
	wallet *C.struct_TWHDWallet
}

// NewWalletCoreRepository loads a BIP39 mnemonic into wallet-core and
// returns a repository bound to it.
func NewWalletCoreRepository(name, mnemonic string) (*WalletCoreRepository, error) {
	if name == "" {
		return nil, errors.New("repository name is required")
	}
	if mnemonic == "" {
		return nil, errors.New("mnemonic is required")
	}

	cMn := types.TWStringCreateWithGoString(mnemonic)
	defer C.TWStringDelete(cMn)
	cPass := types.TWStringCreateWithGoString("")
	defer C.TWStringDelete(cPass)

	wallet := C.TWHDWalletCreateWithMnemonic(cMn, cPass)
	if wallet == nil {
		return nil, fmt.Errorf("invalid mnemonic for repository %q", name)
	}

	return &WalletCoreRepository{
		name:   name,
		wallet: wallet,
	}, nil
}

// Close releases the underlying TWHDWallet. Safe to call multiple times.
func (r *WalletCoreRepository) Close() {
	if r == nil || r.wallet == nil {
		return
	}
	C.TWHDWalletDelete(r.wallet)
	r.wallet = nil
}

// Name returns the configured gate name this repository was created for.
func (r *WalletCoreRepository) Name() string { return r.name }

// DeriveEthereumAddress returns the Ethereum address for the BIP44 path
// m/44'/60'/<account>'/<change>/<addressIndex> without exposing the
// intermediate private key.
func (r *WalletCoreRepository) DeriveEthereumAddress(account, change, addressIndex uint32) (string, error) {
	priv := r.derivedPrivateKey(account, change, addressIndex)
	defer C.TWPrivateKeyDelete(priv)

	pub := C.TWPrivateKeyGetPublicKeyByType(priv, C.TWPublicKeyTypeSECP256k1Extended)
	defer C.TWPublicKeyDelete(pub)

	addr := C.TWAnyAddressCreateWithPublicKey(pub, C.TWCoinTypeEthereum)
	if addr == nil {
		return "", errors.New("failed to derive Ethereum address from public key")
	}
	defer C.TWAnyAddressDelete(addr)

	desc := C.TWAnyAddressDescription(addr)
	defer C.TWStringDelete(desc)

	return types.TWStringGoString(unsafe.Pointer(desc)), nil
}

// SignEthereumTransaction signs an EIP-1559 dynamic-fee transaction with
// the private key derived at m/44'/60'/<account>'/<change>/<addressIndex>.
// The private key is materialised in process memory only for the duration
// of the call and zeroed before return.
func (r *WalletCoreRepository) SignEthereumTransaction(
	account, change, addressIndex uint32,
	p EthereumTxParams,
) (*EthereumSignedTx, error) {
	if err := validateTxParams(p); err != nil {
		return nil, err
	}

	priv := r.derivedPrivateKey(account, change, addressIndex)
	defer C.TWPrivateKeyDelete(priv)

	privDataC := C.TWPrivateKeyData(priv)
	defer C.TWDataDelete(privDataC)
	privBytes := types.TWDataGoBytes(unsafe.Pointer(privDataC))
	defer zero(privBytes)

	input := &ethereum.SigningInput{
		ChainId:               bigIntBytes(p.ChainID),
		Nonce:                 uint64Bytes(p.Nonce),
		TxMode:                ethereum.TransactionMode_Enveloped,
		GasLimit:              uint64Bytes(p.GasLimit),
		MaxFeePerGas:          bigIntBytes(p.MaxFeePerGasWei),
		MaxInclusionFeePerGas: bigIntBytes(p.MaxPriorityFeePerGasWei),
		ToAddress:             p.To,
		PrivateKey:            privBytes,
		Transaction: &ethereum.Transaction{
			TransactionOneof: &ethereum.Transaction_Transfer_{
				Transfer: &ethereum.Transaction_Transfer{
					Amount: bigIntBytes(p.ValueWei),
					Data:   p.Data,
				},
			},
		},
	}

	inputBytes, err := proto.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal SigningInput: %w", err)
	}

	cIn := types.TWDataCreateWithGoBytes(inputBytes)
	defer C.TWDataDelete(cIn)

	cOut := C.TWAnySignerSign(cIn, C.TWCoinTypeEthereum)
	defer C.TWDataDelete(cOut)

	output := &ethereum.SigningOutput{}
	if err := proto.Unmarshal(types.TWDataGoBytes(unsafe.Pointer(cOut)), output); err != nil {
		return nil, fmt.Errorf("unmarshal SigningOutput: %w", err)
	}

	if output.GetError() != common.SigningError_OK {
		msg := output.GetErrorMessage()
		if msg == "" {
			msg = output.GetError().String()
		}
		return nil, fmt.Errorf("wallet-core signing failed: %s", msg)
	}

	encoded := output.GetEncoded()
	if len(encoded) == 0 {
		return nil, errors.New("wallet-core returned empty encoded transaction")
	}

	return &EthereumSignedTx{
		TxHash:  "0x" + hex.EncodeToString(keccak256(encoded)),
		Encoded: encoded,
	}, nil
}

// IsValidEthereumAddress checks an address with wallet-core's coin-aware
// validator. Exposed as a package-level helper because it does not depend
// on a particular HD wallet.
func IsValidEthereumAddress(addr string) bool {
	if addr == "" {
		return false
	}
	cStr := types.TWStringCreateWithGoString(addr)
	defer C.TWStringDelete(cStr)
	return bool(C.TWAnyAddressIsValid(cStr, C.TWCoinTypeEthereum))
}

func (r *WalletCoreRepository) derivedPrivateKey(account, change, addressIndex uint32) *C.struct_TWPrivateKey {
	return C.TWHDWalletGetDerivedKey(
		r.wallet,
		C.TWCoinTypeEthereum,
		C.uint32_t(account),
		C.uint32_t(change),
		C.uint32_t(addressIndex),
	)
}

func validateTxParams(p EthereumTxParams) error {
	if !IsValidEthereumAddress(p.To) {
		return errors.New("destination address is invalid")
	}
	if p.ChainID == nil || p.ChainID.Sign() <= 0 {
		return errors.New("chain id must be positive")
	}
	if p.ValueWei == nil || p.ValueWei.Sign() < 0 {
		return errors.New("value cannot be negative")
	}
	if p.GasLimit == 0 {
		return errors.New("gas limit must be positive")
	}
	if p.MaxFeePerGasWei == nil || p.MaxFeePerGasWei.Sign() <= 0 {
		return errors.New("max fee per gas must be positive")
	}
	if p.MaxPriorityFeePerGasWei == nil || p.MaxPriorityFeePerGasWei.Sign() <= 0 {
		return errors.New("max priority fee per gas must be positive")
	}
	if p.MaxPriorityFeePerGasWei.Cmp(p.MaxFeePerGasWei) > 0 {
		return errors.New("max priority fee cannot exceed max fee per gas")
	}
	return nil
}

func keccak256(data []byte) []byte {
	cIn := types.TWDataCreateWithGoBytes(data)
	defer C.TWDataDelete(cIn)
	cOut := C.TWHashKeccak256(cIn)
	defer C.TWDataDelete(cOut)
	return types.TWDataGoBytes(unsafe.Pointer(cOut))
}

// bigIntBytes returns the canonical big-endian byte encoding wallet-core
// expects for uint256 fields. Zero is encoded as nil (empty bytes), which
// the protobuf wire format treats identically to a zero-length payload.
func bigIntBytes(n *big.Int) []byte {
	if n == nil || n.Sign() == 0 {
		return nil
	}
	return n.Bytes()
}

func uint64Bytes(n uint64) []byte {
	if n == 0 {
		return nil
	}
	return new(big.Int).SetUint64(n).Bytes()
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
