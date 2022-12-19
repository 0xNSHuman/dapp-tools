package common

type ErrorDomain uint

const (
	ErrorDomainUndefined = iota * 100_000
	ErrorDomainMeta
	ErrorDomainClient
	ErrorDomainWallet
)

type MetaError uint

const (
	Unknown MetaError = ErrorDomainMeta + iota
	NotSupported
)

func (e MetaError) Error() string {
	switch e {
	case NotSupported:
		return "Not supported"
	default:
		return "Unknown"
	}
}
