package pricer

import "github.com/abergasov/market_timer/internal/entities"

type Observer interface {
	Subscribe(price float64) (chan entities.GasRates, error)
	Unsubscribe(price float64)
}
