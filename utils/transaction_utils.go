package utils

import (
	"math/big"

	"github.com/0xNSHuman/dapp-tools/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func EncodeTransaction(
	client *client.Client,
	from common.Address,
	to common.Address,
	value *big.Int,
	calldata []byte,
	gasMultiplier float64,
) (*types.Transaction, error) {
	callMessage, err := client.CreateCallMessage(from, to, value, calldata)
	if err != nil {
		return nil, err
	}

	tx, err := client.CreateTransaction(*callMessage, gasMultiplier)
	if err != nil {
		return nil, err
	}

	return tx, err
}
