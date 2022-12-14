package apps

import (
	"context"
	"math/big"

	"github.com/0xNSHuman/dapp-tools/client"
	"github.com/ethereum/go-ethereum/core/types"
)

type BlockExplorer struct {
	client *client.Client
}

func NewBlockExplorer(rpcEndpoint string) (*BlockExplorer, error) {
	client, err := client.NewClient(rpcEndpoint)
	if err != nil {
		return nil, err
	}

	return &BlockExplorer{client: client}, nil
}

func (be *BlockExplorer) BlockNumber() (*big.Int, error) {
	header, err := be.blockHeader()
	if err != nil {
		return nil, err
	}

	return header.Number, nil
}

func (be *BlockExplorer) blockHeader() (*types.Header, error) {
	header, err := be.client.EthClient.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	return header, nil
}
