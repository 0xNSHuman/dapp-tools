package client

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client struct {
	EthClient *ethclient.Client
}

func (c *Client) ChainID() (*big.Int, error) {
	return c.EthClient.ChainID(context.Background())
}

func NewClient(rpcEndpoint string) (*Client, error) {
	rpcClient, err := rpc.Dial(rpcEndpoint)
	if err != nil {
		return nil, BadRPCConnection
	}

	ethClient := ethclient.NewClient(rpcClient)

	return &Client{EthClient: ethClient}, nil
}

func (c *Client) CreateCallMessage(
	from common.Address,
	to common.Address,
	value *big.Int,
	data []byte,
) (*ethereum.CallMsg, error) {
	// TODO: Handle client errors while constructing a call

	msg := &ethereum.CallMsg{
		From:      from,
		To:        &to,
		Gas:       0,
		GasPrice:  &big.Int{},
		GasFeeCap: &big.Int{},
		GasTipCap: &big.Int{},
		Value:     value,
		Data:      data,
	}

	return msg, nil
}

func (c *Client) CreateTransaction(msg ethereum.CallMsg) (*types.Transaction, error) {
	// TODO: Handle client errors while constructing a transaction

	chainId, _ := c.EthClient.ChainID(context.Background())
	nonce, _ := c.EthClient.PendingNonceAt(context.Background(), msg.From)
	gasPrice, _ := c.EthClient.SuggestGasPrice(context.Background())
	gasTip, _ := c.EthClient.SuggestGasTipCap(context.Background())
	gasLimit, _ := c.EthClient.EstimateGas(context.Background(), msg)
	toAddress := msg.To
	value := msg.Value

	txData := &types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     nonce,
		GasTipCap: gasTip,
		GasFeeCap: gasPrice,
		Gas:       gasLimit,
		To:        toAddress,
		Value:     value,
		Data:      msg.Data,
	}

	tx := types.NewTx(txData)
	return tx, nil
}

func (c *Client) SendTransaction(tx *types.Transaction) (*string, error) {
	err := c.EthClient.SendTransaction(context.Background(), tx)
	if err != nil {
		return nil, err
	}

	receipt, err := bind.WaitMined(context.Background(), c.EthClient, tx)
	if err != nil {
		return nil, err
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, TransactionFailed
	}

	txHash := receipt.TxHash.Hex()

	return &txHash, nil
}
