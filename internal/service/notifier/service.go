package notifier

import (
	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/service/pricer"
)

// Service keep info about price subscriptions
// and notify consumers about price changes
type Service struct {
	priceObservers map[string]*pricer.Service
	priceSubs      map[string]map[float64][]chan entities.GasRates
}

func NewService() *Service {
	return &Service{}
}
