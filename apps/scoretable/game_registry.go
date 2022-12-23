package scoretable

import (
	"context"
	"math/big"

	"github.com/0xNSHuman/dapp-tools/client"
	"github.com/0xNSHuman/dapp-tools/utils"
	"github.com/0xNSHuman/dapp-tools/wallet"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type GameRegistry struct {
	address common.Address
	abi     abi.ABI
	client  *client.Client
	wallet  *wallet.WalletKeeper
}

func NewGameRegistry(
	address common.Address,
	abi abi.ABI,
	client *client.Client,
	walletUI wallet.WalletUI,
) (*GameRegistry, error) {
	wallet, err := wallet.NewWalletKeeper(walletUI, false)
	if err != nil {
		return nil, err
	}

	return &GameRegistry{
		address: address,
		abi:     abi,
		client:  client,
		wallet:  wallet,
	}, nil
}

func (r *GameRegistry) GameCount() (*big.Int, error) {
	callData, err := r.abi.Pack("gameCount")
	if err != nil {
		return nil, err
	}

	result, err := r.client.EthClient.CallContract(
		context.Background(),
		ethereum.CallMsg{
			To:   &r.address,
			Data: callData,
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	var output []interface{}
	output, err = r.abi.Unpack("gameCount", result)
	if err != nil {
		return nil, err
	}

	return output[0].(*big.Int), nil
}

func (r *GameRegistry) RegisterGame(name string, meta [32]byte) (*string, error) {
	pubkey, err := r.wallet.PublicKey(0)
	if err != nil {
		return nil, err
	}

	from := common.HexToAddress(pubkey)
	to := r.address
	value := &big.Int{}

	callData, err := r.abi.Pack("registerGame", name, meta)
	if err != nil {
		return nil, err
	}

	tx, err := utils.EncodeTransaction(r.client, from, to, value, callData, 1.15)
	if err != nil {
		return nil, err
	}

	chainId, err := r.client.ChainID()
	if err != nil {
		return nil, err
	}

	signedTx, err := r.wallet.SignTransaction(chainId, tx, common.Address{}, false)
	if err != nil {
		return nil, err
	}

	txHash, err := r.client.SendTransaction(signedTx)
	if err != nil {
		return nil, err
	}

	return txHash, nil
}
