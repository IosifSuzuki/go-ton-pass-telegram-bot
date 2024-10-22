package app

type CryptoBotMethod string

const (
	CreateInvoiceCryptoBotMethod CryptoBotMethod = "createInvoice"
	ExchangeRateCryptoBotMethod                  = "getExchangeRates"
)
