package service

import (
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/internal/utils"
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
	GetServicePrices(code string) ([]sms.PriceForService, error)
	GetPopularServiceCodeList() ([]string, error)
	RequestNumber(serviceCode string, countryNumber int64, maxPrice float64) (*sms.RequestedNumber, error)
	GetStatus(activationID int64) (app.SMSActivationState, error)
	CancelActivation(activationID int64) error
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

func (s *smsService) GetPopularServiceCodeList() ([]string, error) {
	var response map[string]any
	urlValues := url.Values{}
	req, err := s.prepareRequest(app.GetTopCountriesByServiceAction, urlValues)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	popularServices := make([]string, 0, len(response))
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		t, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := t.(string)
		if ok {
			popularServices = append(popularServices, key)
		}
	}
	return popularServices, nil
}

func (s *smsService) GetServicePrices(serviceCode string) ([]sms.PriceForService, error) {
	urlValues := url.Values{}
	urlValues.Set("service", serviceCode)
	urlValues.Set("freePrice", "true")
	req, err := s.prepareRequest(app.GetTopCountriesByServiceAction, urlValues)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]sms.PriceForService
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	priceForServices := make([]sms.PriceForService, 0)
	for _, priceForService := range result {
		priceForServices = append(priceForServices, priceForService)
	}
	return priceForServices, nil
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

func (s *smsService) GetStatus(activationID int64) (app.SMSActivationState, error) {
	log := s.container.GetLogger()
	urlValues := url.Values{}
	urlValues.Set("id", strconv.Itoa(int(activationID)))
	req, err := s.prepareRequest(app.GetActivationStatus, urlValues)
	if err != nil {
		return app.UnknownSMSActivateState, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return app.UnknownSMSActivateState, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	text := string(body)
	log.Debug("receive from sms activate response", logger.F("text", text))
	if utils.Equal(text, string(app.CancelSMSActivateState)) {
		return app.CancelSMSActivateState, nil
	} else if utils.Equal(text, string(app.DoneSMSActivateState)) {
		return app.DoneSMSActivateState, nil
	}
	return app.PendingSMSActivateState, nil
}

func (s *smsService) CancelActivation(activationID int64) error {
	log := s.container.GetLogger()
	urlValues := url.Values{}
	urlValues.Set("id", strconv.Itoa(int(activationID)))
	urlValues.Set("status", "8")
	req, err := s.prepareRequest(app.SetActivationStatus, urlValues)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	text := string(body)
	log.Debug("cancel activation", logger.F("response", text))
	return nil
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
	if err := json.Unmarshal(body, &errorResponse); err == nil {
		if err := sms.DecodeError(errorResponse.Message); err != nil {
			return *err, errorResponse.Info
		}
	}
	if err := sms.DecodeError(string(body)); err != nil {
		return *err, nil
	}
	return app.UnknownError, errorResponse.Info
}
