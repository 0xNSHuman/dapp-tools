package wallet

import (
	"fmt"
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

type WalletUI interface {
	EnterPassphrase() (string, error)
}

type WalletKeeper struct {
	ks *keystore.KeyStore
	am *accounts.Manager
	ui WalletUI
}

func NewWalletKeeper(ui WalletUI, autoUnlock bool) (*WalletKeeper, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, FileSystemAccess
	}

	keystorePath := filepath.Join(userHomeDir, "evm", "wallet", "keystore")

	ks := keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
	am := accounts.NewManager(&accounts.Config{InsecureUnlockAllowed: autoUnlock}, ks)

	return &WalletKeeper{
		ks: ks,
		am: am,
		ui: ui,
	}, nil
}

func (wk *WalletKeeper) CreateWallet(passphrase string) error {
	_, error := wk.ks.NewAccount(passphrase)

	return error
}

func (wk *WalletKeeper) ImportWallet(mode ImportMode, input []byte, passphrase string) error {
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

func (wk *WalletKeeper) ExportWallet(index int, mode ExportMode, passphrase string) ([]byte, error) {
	accs := wk.ks.Accounts()
	if len(accs) <= index {
		return nil, AccountNotFound
	}

	// TODO: Require a pwd change?
	keyJSON, err := wk.ks.Export(accs[index], passphrase, passphrase)
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

func (wk *WalletKeeper) NumberOfAccounts() int {
	return len(wk.ks.Accounts())
}

func (wk *WalletKeeper) PublicKey(index int) (string, error) {
	accs := wk.ks.Accounts()
	if len(accs) <= index {
		return "", AccountNotFound
	}

	return accs[index].Address.Hex(), nil
}

func (wk *WalletKeeper) Unlock(index int, passphrase string) error {
	accs := wk.ks.Accounts()
	if len(accs) <= index {
		return AccountNotFound
	}

	return wk.ks.TimedUnlock(accs[index], passphrase, 0)
}

func (wk *WalletKeeper) SignTransaction(
	chainId *big.Int,
	tx *types.Transaction,
	signer gethcommon.Address,
	autosign bool,
) (*types.Transaction, error) {
	var signerAcc accounts.Account = accounts.Account{}

	for _, acc := range wk.ks.Accounts() {
		if acc.Address == signer {
			signerAcc = acc
		}
	}
	if signerAcc == (accounts.Account{}) {
		return nil, AccountNotFound
	}

	if !autosign {
		passphrase, err := wk.ui.EnterPassphrase()
		if err != nil {
			return nil, err
		}

		err = wk.ks.Unlock(signerAcc, passphrase)
		if err != nil {
			return nil, UnauthorizedAccess
		}
	}

	fmt.Println("Signing with address:", signerAcc.Address.Hex())
	fmt.Println()

	signedTx, err := wk.ks.SignTx(signerAcc, tx, chainId)
	if err != nil {
		fmt.Println(err)
		return nil, SigningFailed
	}

	if !autosign {
		wk.ks.Lock((signerAcc).Address)
	}

	return signedTx, nil
}

func (wk *WalletKeeper) DeleteWallet(index int, passphrase string) error {
	accs := wk.ks.Accounts()
	if len(accs) <= index {
		return AccountNotFound
	}

	err := wk.ks.Delete(accs[index], passphrase)
	if err != nil {
		return err
	}

	return nil
}
