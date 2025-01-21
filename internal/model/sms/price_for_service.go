package sms

import (
	"go-ton-pass-telegram-bot/internal/utils"
	"math"
)

type PriceForService struct {
	RetailPrice  float64        `json:"retail_price"`
	CountryCode  int64          `json:"country"`
	FreePriceMap map[string]int `json:"freePriceMap"`
	MinPrice     PriceFiled     `json:"price"`
	Count        int            `json:"count"`
}

func (p *PriceForService) MinPriceCount() int {
	minPrice := math.MaxFloat64
	count := 0
	for key, value := range p.FreePriceMap {
		price, _ := utils.ParseFloat64FromText(key)
		if minPrice > price {
			minPrice = price
			count = value
		}
	}
	return count
}
