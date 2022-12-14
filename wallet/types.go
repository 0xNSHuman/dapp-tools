package wallet

import (
	"github.com/0xNSHuman/dapp-tools/common"
)

type WalletError uint

const (
	Unknown WalletError = common.ErrorDomainWallet + iota
	FileSystemAccess
	AccountNotFound
	UnauthorizedAccess
	InvalidPrivateKey
	InvalidSeedPhrase
	SigningFailed
)

func (e WalletError) Error() string {
	switch e {
	case FileSystemAccess:
		return "Can't access the file system"
	case AccountNotFound:
		return "Account not found"
	case UnauthorizedAccess:
		return "Unauthorized access"
	case InvalidPrivateKey:
		return "Invalid private key"
	case InvalidSeedPhrase:
		return "Invalid seed phrase"
	case SigningFailed:
		return "Transaction signing failed"
	default:
		return "Unknown"
	}
}

type ImportMode uint8

const (
	ImportModeSeedPhrase ImportMode = iota
	ImportModePrivateKey
)

type ExportMode uint8

const (
	ExportModeSeedPhrase ExportMode = iota
	ExportModePrivateKey
)
