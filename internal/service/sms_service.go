package service

import (
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/pkg/logger"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	baseSMSURL = "https://api.sms-activate.org/stubs/handler_api.php"
)

type SMSService interface {
	GetServices() ([]sms.Service, error)
	GetCountries() ([]sms.Country, error)
	GetServicePrices(code string) ([]sms.ServicePrice, error)
	RequestNumber(serviceCode string, countryNumber int64, maxPrice float64) (*sms.RequestedNumber, error)
}

type smsService struct {
	container     container.Container
	servicePrices map[string][]sms.ServicePrice
}

func NewSMSService(container container.Container) SMSService {
	return &smsService{
		container:     container,
		servicePrices: make(map[string][]sms.ServicePrice),
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

func (s *smsService) GetServicePrices(serviceCode string) ([]sms.ServicePrice, error) {
	servicePrice, err := s.getServicePricesFromCache(serviceCode)
	if err != nil && err != app.NilError {
		return nil, err
	} else if err == nil && servicePrice != nil {
		return servicePrice, nil
	}
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
	s.servicePrices[serviceCode] = servicePrices
	return servicePrices, nil
}

func (s *smsService) RequestNumber(serviceCode string, countryNumber int64, maxPrice float64) (*sms.RequestedNumber, error) {
	log := s.container.GetLogger()
	urlValues := url.Values{}
	urlValues.Set("service", serviceCode)
	urlValues.Set("country", strconv.FormatInt(countryNumber, 10))
	urlValues.Set("useCashBack", "true")
	maxPriceInText := strconv.FormatFloat(maxPrice, 'f', 2, 64)
	urlValues.Set("maxPrice", maxPriceInText)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debug("get response from RequestNumber endpoint", logger.F("response", string(body)))
	requestedNumber := sms.RequestedNumber{}
	err = json.Unmarshal(body, &requestedNumber)
	if err == nil && requestedNumber.ActivationID != "" {
		return &requestedNumber, nil
	}
	err, errInfo := s.handleRequestNumberError(body)
	if strings.EqualFold(err.Error(), sms.WrongMaxPriceErrorName) && errInfo != nil {
		correctedPrice := errInfo["min"].(float64)
		if math.Abs(correctedPrice-maxPrice) > 0.1 {
			return s.RequestNumber(serviceCode, countryNumber, correctedPrice)
		}
		return nil, err
	}
	return nil, err
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
	req, err := http.NewRequest(http.MethodGet, urlPath.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	log.Debug("prepare request", logger.F("url", req.URL))
	return req, nil
}

func (s *smsService) handleRequestNumberError(body []byte) (error, map[string]any) {
	var errorResponse sms.ErrorResponse
	if err := json.Unmarshal(body, &errorResponse); err != nil {
		return err, nil
	}
	if err := sms.DecodeError(errorResponse.Message); err != nil {
		return *err, errorResponse.Info
	}
	if err := sms.DecodeError(string(body)); err != nil {
		return *err, nil
	}
	return app.UnknownError, errorResponse.Info
}

func (s *smsService) getServicePricesFromCache(code string) ([]sms.ServicePrice, error) {
	priceServices, ok := s.servicePrices[code]
	if ok {
		return priceServices, nil
	}
	return nil, app.NilError
}
