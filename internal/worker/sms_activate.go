package worker

import (
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/service"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/logger"
	"sort"
)

type SMSActivate interface {
	GetOrderedServices() ([]sms.Service, error)
	GetPriceForService(serviceCode string) ([]sms.PriceForService, error)
}

type smsActivate struct {
	container  container.Container
	smsService service.SMSService
	cache      service.Cache
}

func NewSMSActivate(container container.Container, smsService service.SMSService, cache service.Cache) SMSActivate {
	return &smsActivate{
		container:  container,
		smsService: smsService,
		cache:      cache,
	}
}

func (s *smsActivate) GetOrderedServices() ([]sms.Service, error) {
	log := s.container.GetLogger()
	popularServiceCodes, err := s.smsService.GetPopularServiceCodeList()
	if err != nil {
		return nil, err
	}
	log.Debug("got popular services", logger.F("popularServiceCodes", popularServiceCodes))
	allServices, err := s.smsService.GetServices()
	if err != nil {
		return nil, err
	}
	log.Debug("got allServices", logger.F("allServices", allServices))
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
	servicePrices, err := s.smsService.GetServicePrices(serviceCode)
	if err != nil {
		return nil, err
	}
	servicePrices = utils.Filter(servicePrices, func(servicePrice sms.PriceForService) bool {
		return servicePrice.MinPrice > 0
	})
	sort.Slice(servicePrices, func(i, j int) bool {
		return servicePrices[i].MinPrice < servicePrices[j].MinPrice
	})
	return servicePrices, nil
}
