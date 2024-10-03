package service

import (
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/pkg/logger"
	"net/http"
	"net/url"
	"strconv"
)

const (
	baseSMSURL = "https://api.sms-activate.org/stubs/handler_api.php"
)

type SMSService interface {
	GetServices() ([]sms.Service, error)
	GetCountries() ([]sms.Country, error)
	GetPriceForService(code string) ([]sms.ServicePrice, error)
	RequestNumber(serviceCode string) (*sms.RequestedNumber, error)
}

type smsService struct {
	container container.Container
}

func NewSMSService(container container.Container) SMSService {
	return &smsService{
		container: container,
	}
}

func (s *smsService) GetServices() ([]sms.Service, error) {
	type Response struct {
		Services []sms.Service
	}
	req, err := s.prepareRequest(app.GetServicesListSMSAction, url.Values{})
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	response := Response{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return response.Services, nil
}

func (s *smsService) GetCountries() ([]sms.Country, error) {
	req, err := s.prepareRequest(app.GetCountriesListSMSAction, url.Values{})
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var countries map[string]sms.Country

	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, err
	}
	allCountries := make([]sms.Country, 0, len(countries))
	for _, country := range countries {
		allCountries = append(allCountries, country)
	}
	return allCountries, nil
}

func (s *smsService) GetPriceForService(serviceCode string) ([]sms.ServicePrice, error) {
	urlValues := url.Values{}
	urlValues.Set("service", serviceCode)
	req, err := s.prepareRequest(app.GetPricesSMSAction, urlValues)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]map[string]sms.ServicePrice
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	servicePrices := make([]sms.ServicePrice, 0)
	for countryCode, part := range result {
		servicePrice := part[serviceCode]
		servicePrice.Code = serviceCode
		countryCodeInt, err := strconv.Atoi(countryCode)
		if err != nil {
			continue
		}
		servicePrice.CountryCode = countryCodeInt
		servicePrices = append(servicePrices, servicePrice)
	}
	return servicePrices, nil
}

func (s *smsService) RequestNumber(serviceCode string) (*sms.RequestedNumber, error) {
	urlValues := url.Values{}
	urlValues.Set("service", serviceCode)
	urlValues.Set("maxPrice", "0")
	req, err := s.prepareRequest(app.GetNumberSMSAction, urlValues)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	requestedNumber := sms.RequestedNumber{}
	if err := json.NewDecoder(resp.Body).Decode(&requestedNumber); err != nil {
		//return nil, err
	}
	return &requestedNumber, nil
}

func (s *smsService) prepareRequest(smsAction app.SMSAction, queryParams url.Values) (*http.Request, error) {
	log := s.container.GetLogger()
	apiKey := s.container.GetConfig().SMSKey()
	urlPath, err := url.Parse(baseSMSURL)
	if err != nil {
		return nil, err
	}
	queryParams.Set("api_key", apiKey)
	queryParams.Set("action", string(smsAction))
	urlPath.RawQuery = queryParams.Encode()
	req, err := http.NewRequest("GET", urlPath.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	log.Debug("prepare request", logger.F("url", req.URL))
	return req, nil
}
