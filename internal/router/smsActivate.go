package router

import (
	"encoding/json"
	"errors"
	"go-ton-pass-telegram-bot/internal/container"
	"go-ton-pass-telegram-bot/internal/controller/sms"
	smsModel "go-ton-pass-telegram-bot/internal/model/sms"
	"go-ton-pass-telegram-bot/pkg/logger"
	"io"
	"net/http"
)

type SMSActivate struct {
	container  container.Container
	controller sms.SMSActivateController
}

func NewSMSActivateRouter(container container.Container, controller sms.SMSActivateController) *SMSActivate {
	return &SMSActivate{
		container:  container,
		controller: controller,
	}
}

func (s *SMSActivate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		w.WriteHeader(http.StatusInternalServerError)
	}()
	log := s.container.GetLogger()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("can't read body", logger.FError(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debug("receive body", logger.F("json", string(body)))
	var update smsModel.WebhookUpdates
	if err := json.Unmarshal(body, &update); err != nil {
		if err := filterSMSActivateErrors(err); err != nil {
			log.Fatal("fail to decode", logger.FError(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	err = s.controller.Serve(&update)
	if err != nil {
		log.Fatal("controller has failed", logger.FError(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func filterSMSActivateErrors(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
