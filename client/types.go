package client

import (
	"github.com/0xNSHuman/dapp-tools/common"
)

type ClientError uint

const (
	Unknown ClientError = common.ErrorDomainWallet + iota
	BadRPCConnection
	TransactionFailed
)

func (e ClientError) Error() string {
	switch e {
	case BadRPCConnection:
		return "Bad RPC connection"
	case TransactionFailed:
		return "Transaction failed"
	default:
		return "Unknown"
	}
}
