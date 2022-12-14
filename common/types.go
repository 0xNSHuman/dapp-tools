package common

type ErrorDomain uint

const (
	ErrorDomainUndefined = iota * 100_000
	ErrorDomainClient
	ErrorDomainWallet
)
