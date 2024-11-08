package app

type CryptoBotMethod string

const (
	CreateInvoiceCryptoBotMethod CryptoBotMethod = "createInvoice"
	ExchangeRateCryptoBotMethod  CryptoBotMethod = "getExchangeRates"
	DeleteInvoiceCryptoBotMethod CryptoBotMethod = "deleteInvoice"
)
