package worker

import (
	"context"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"sort"
	"time"
)

type SMSActivate interface {
	GetOrderedServices() ([]sms.Service, error)
	GetPriceForService(serviceCode string) ([]sms.PriceForService, error)
	GetCountries() ([]sms.Country, error)
}

type smsActivate struct {
	container    container.Container
	smsService   service.SMSService
	cache        service.Cache
	serviceOrder []string
}

func NewSMSActivate(container container.Container, smsService service.SMSService, cache service.Cache) SMSActivate {
	return &smsActivate{
		container:    container,
		smsService:   smsService,
		cache:        cache,
		serviceOrder: make([]string, 0),
	}
}

func (s *smsActivate) GetOrderedServices() ([]sms.Service, error) {
	log := s.container.GetLogger()
	popularServiceCodes, err := s.smsService.GetPopularServiceCodeList()
	if err != nil {
		return nil, err
	}
	preferredServiceCodesOrder := s.container.GetPreferredServiceCodesOrder()
	for _, item := range preferredServiceCodesOrder {
		idx := utils.FirstIndexOf(popularServiceCodes, item)
		if idx == -1 {
			continue
		}
		popularServiceCodes = append(popularServiceCodes[:idx], popularServiceCodes[idx+1:]...)
	}
	popularServiceCodes = append(preferredServiceCodesOrder, popularServiceCodes...)
	log.Debug("got popular services", logger.F("popularServiceCodes", popularServiceCodes))
	allServices, err := s.smsService.GetServices()
	if err != nil {
		return nil, err
	}
	sort.Slice(allServices, func(i, j int) bool {
		lhsService := allServices[i]
		rhsService := allServices[j]
		lhsPopularServiceCodeIndex := utils.FirstIndexOf(popularServiceCodes, lhsService.Code)
		rhsPopularServiceCodeIndex := utils.FirstIndexOf(popularServiceCodes, rhsService.Code)
		if lhsPopularServiceCodeIndex == -1 && rhsPopularServiceCodeIndex == -1 {
			return false
		}
		if lhsPopularServiceCodeIndex == -1 {
			return false
		}
		if rhsPopularServiceCodeIndex == -1 {
			return true
		}
		return lhsPopularServiceCodeIndex < rhsPopularServiceCodeIndex
	})
	return allServices, nil
}

func (s *smsActivate) GetPriceForService(serviceCode string) ([]sms.PriceForService, error) {
	preferredCountryCodesOrder := s.container.GetPreferredCountryCodesOrder()
	countries, err := s.GetCountries()
	if err != nil {
		return nil, err
	}
	servicePrices, err := s.smsService.GetServicePrices(serviceCode)
	if err != nil {
		return nil, err
	}
	servicePrices = utils.Filter(servicePrices, func(servicePrice sms.PriceForService) bool {
		return servicePrice.MinPrice > 0
	})
	countryMap := make(map[int64]sms.Country)
	for _, country := range countries {
		countryMap[country.ID] = country
	}
	sort.Slice(servicePrices, func(i, j int) bool {
		lhsServicePrice := servicePrices[i]
		rhsServicePrice := servicePrices[j]
		lhsCountry := countryMap[lhsServicePrice.CountryCode]
		rhsCountry := countryMap[rhsServicePrice.CountryCode]

		lhsIndexInServiceCodesOrder := utils.FirstIndexOf(preferredCountryCodesOrder, lhsCountry.Title)
		rhsIndexInServiceCodesOrder := utils.FirstIndexOf(preferredCountryCodesOrder, rhsCountry.Title)
		if lhsIndexInServiceCodesOrder != -1 && rhsIndexInServiceCodesOrder == -1 {
			return true
		} else if lhsIndexInServiceCodesOrder != -1 && rhsIndexInServiceCodesOrder != -1 {
			return lhsIndexInServiceCodesOrder < rhsIndexInServiceCodesOrder
		} else if lhsIndexInServiceCodesOrder == -1 && rhsIndexInServiceCodesOrder != -1 {
			return false
		}

		// left preferred sort from sms-activate service
		if lhsServicePrice.RetailPrice == rhsServicePrice.RetailPrice {
			return lhsServicePrice.CountryCode < rhsServicePrice.CountryCode
		}

		return lhsServicePrice.RetailPrice < rhsServicePrice.RetailPrice
	})
	return servicePrices, nil
}

func (s *smsActivate) GetCountries() ([]sms.Country, error) {
	go func() {
		ctx := context.Background()
		_ = s.UpToCountries(ctx)
	}()
	ctx := context.Background()
	cacheResponse, err := s.cache.GetSMSCountries(ctx)
	if err == nil {
		return cacheResponse.Result, nil
	}
	countries, err := s.smsService.GetCountries()
	if err != nil {
		return nil, err
	}
	return countries, nil
}

func (s *smsActivate) UpToCountries(ctx context.Context) error {
	response, err := s.cache.GetSMSCountries(ctx)
	if err != nil {
		return err
	}
	var timeFetched time.Time
	if response != nil {
		timeFetched = response.TimeFetched
	} else {
		timeFetched = time.Now()
	}
	if timeFetched.Add(24*time.Hour).Compare(time.Now()) <= 0 {
		return nil
	}
	countries, err := s.smsService.GetCountries()
	return s.cache.SaveSMSCountries(ctx, countries)
}
