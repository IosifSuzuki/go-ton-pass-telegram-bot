package stripe_payment

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"go-ton-pass-telegram-bot/internal/utils"
	"go-ton-pass-telegram-bot/pkg/stripe_payment/model"
)

type StripePaymentClient interface {
	CreatePaymentLink(title string, amount float64, currency model.Currency, metadata map[string]string) (*model.CheckoutSession, error)
}

type stripePaymentClient struct {
	secretKey  string
	successURL string
	cancelURL  string
}

func NewStripePaymentClient(secretKey string, successURL, cancelURL string) StripePaymentClient {
	return &stripePaymentClient{
		secretKey:  secretKey,
		successURL: successURL,
		cancelURL:  cancelURL,
	}
}

func (c *stripePaymentClient) CreatePaymentLink(title string, amount float64, currency model.Currency, metadata map[string]string) (*model.CheckoutSession, error) {
	stripe.Key = c.secretKey
	stripeAmount := amount * 100
	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					UnitAmount: stripe.Int64(int64(stripeAmount)),
					Currency:   stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String(title),
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(stripe.CheckoutSessionModePayment),
		SuccessURL: stripe.String(c.successURL),
		CancelURL:  stripe.String(c.cancelURL),
	}

	for key, value := range metadata {
		params.AddMetadata(key, value)
	}
	s, err := session.New(params)
	if err != nil {
		return nil, err
	}

	return &model.CheckoutSession{
		PaymentLink: utils.NewString(s.URL),
		Status:      string(s.Status),
	}, nil
}
