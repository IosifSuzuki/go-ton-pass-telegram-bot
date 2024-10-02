package app

type SMSAction string

const (
	GetServicesListSMSAction  SMSAction = "getServicesList"
	GetCountriesListSMSAction           = "getCountries"
	GetNumberSMSAction                  = "getNumberV2"
	GetPricesSMSAction                  = "getPrices"
)
