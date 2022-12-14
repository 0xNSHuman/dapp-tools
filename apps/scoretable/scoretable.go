package scoretable

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/0xNSHuman/dapp-tools/client"
	"github.com/0xNSHuman/dapp-tools/wallet"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Scoretable struct {
	Registry *GameRegistry
}

func NewScoretable(rpcEndpoint string, walletUI wallet.WalletUI) (*Scoretable, error) {
	client, err := client.NewClient(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	// Hardcoded address on Goerli
	contractAddress := common.HexToAddress("0xaa77116B42baFB2C9b5661FCFe3C79e91D535Eb3")

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	abiPath := filepath.Join(userHomeDir, "evm", "contracts", "abi", "IGameRegistry.json")
	contractABIFile, err := os.ReadFile(abiPath)
	if err != nil {
		return nil, err
	}

	contractABI, err := abi.JSON(bytes.NewReader(contractABIFile))
	if err != nil {
		return nil, err
	}

	registry, err := NewGameRegistry(contractAddress, contractABI, client, walletUI)
	if err != nil {
		return nil, err
	}

	return &Scoretable{Registry: registry}, nil
}
