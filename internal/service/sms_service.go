package service

import (
	"encoding/json"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/model/app"
	"go-ton-pass-telegram-bot/internal/model/sms"
	"net/http"
	"net/url"
)

const (
	baseURL = "https://api.sms-activate.org/stubs/handler_api.php"
)

type SMSService interface {
	GetListOfAllServices() ([]sms.Service, error)
}

type smsService struct {
	container container.Container
}

func NewSMSService(container container.Container) SMSService {
	return &smsService{
		container: container,
	}
}

func (s *smsService) GetListOfAllServices() ([]sms.Service, error) {
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

func (s *smsService) prepareRequest(smsAction app.SMSAction, queryParams url.Values) (*http.Request, error) {
	apiKey := s.container.GetConfig().SMSKey()
	urlPath, err := url.Parse(baseURL)
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
	return req, nil
}
