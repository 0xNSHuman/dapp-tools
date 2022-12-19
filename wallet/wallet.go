package wallet

import (
	"math/big"
	"os"
	"path/filepath"

	"github.com/0xNSHuman/dapp-tools/common"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
)

type WalletAPI interface {
	CreateWallet(passphrase string) error
	ImportWallet(mode ImportMode, input []byte, passphrase string) error
	ExportWallet(mode ExportMode, passphrase string) ([]byte, error)
	PublicKey() (string, error)
	SignTransaction() ([]byte, error)
	DeleteWallet(passphrase string) error
}

type WalletUI interface {
	EnterPassphrase() (string, error)
}

type WalletKeeper struct {
	ks *keystore.KeyStore
	am *accounts.Manager
	ui WalletUI
}

func NewWalletKeeper(ui WalletUI) (*WalletKeeper, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, FileSystemAccess
	}

	keystorePath := filepath.Join(userHomeDir, "evm", "wallet", "keystore")

	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: false}, ks)

	return &WalletKeeper{
		ks: ks,
		am: am,
		ui: ui,
	}, nil
}

func (wk *WalletKeeper) CreateWallet(passphrase string) error {
	wk.DeleteWallet(passphrase)

	_, error := wk.ks.NewAccount(passphrase)

	return error
}

func (wk *WalletKeeper) ImportWallet(mode ImportMode, input []byte, passphrase string) error {
	wk.DeleteWallet(passphrase)

	switch mode {
	case ImportModePrivateKey:
		privKey, err := crypto.HexToECDSA(gethcommon.Bytes2Hex(input))
		if err != nil {
			return InvalidPrivateKey
		}

		_, err = wk.ks.ImportECDSA(privKey, passphrase)
		if err != nil {
			return err
		}
	case ImportModeSeedPhrase:
		rootKey, err := bip32.NewMasterKey(input)
		if err != nil {
			return InvalidSeedPhrase
		}

		privKey, err := rootKey.NewChildKey(0)
		if err != nil {
			return err
		}

		return wk.ImportWallet(ImportModePrivateKey, []byte(privKey.String()), passphrase)
	}

	return nil
}

func (wk *WalletKeeper) ExportWallet(mode ExportMode, passphrase string) ([]byte, error) {
	accs := wk.ks.Accounts()

	if len(accs) == 0 {
		return nil, AccountNotFound
	}

	// TODO: Require a pwd change?
	keyJSON, err := wk.ks.Export(accs[0], passphrase, passphrase)
	if err != nil {
		return nil, err
	}

	switch mode {
	case ExportModePrivateKey:
		key, err := keystore.DecryptKey(keyJSON, passphrase)
		if err != nil {
			return nil, err
		}

		return crypto.FromECDSA(key.PrivateKey), nil
	case ExportModeSeedPhrase:
		return nil, common.NotSupported
	}

	panic(common.NotSupported)
}

func (wk *WalletKeeper) PublicKey() (string, error) {
	accs := wk.ks.Accounts()

	if len(accs) == 0 {
		return "", AccountNotFound
	}

	return accs[0].Address.Hex(), nil
}

func (wk *WalletKeeper) SignTransaction(chainId *big.Int, tx *types.Transaction) (*types.Transaction, error) {
	accs := wk.ks.Accounts()

	if len(accs) == 0 {
		return nil, AccountNotFound
	}

	passphrase, err := wk.ui.EnterPassphrase()
	if err != nil {
		return nil, err
	}

	err = wk.ks.Unlock(accs[0], passphrase)
	if err != nil {
		return nil, UnauthorizedAccess
	}

	signedTx, err := wk.ks.SignTx(accs[0], tx, chainId)
	if err != nil {
		return nil, SigningFailed
	}

	wk.ks.Lock(accs[0].Address)

	return signedTx, nil
}

func (wk *WalletKeeper) DeleteWallet(passphrase string) error {
	for _, acc := range wk.ks.Accounts() {
		err := wk.ks.Delete(acc, passphrase)
		if err != nil {
			return err
		}
	}

	return nil
}
