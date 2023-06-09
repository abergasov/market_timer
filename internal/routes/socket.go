package routes

import (
	"fmt"
	"strconv"

	"github.com/abergasov/market_timer/internal/entities"
	"github.com/gofiber/websocket/v2"
)

func (s *Server) handleTransactions(c *websocket.Conn) {
	chainID := c.Params("chain")
	if !entities.ValidateChain(chainID) {
		s.log.Error("error validate chain", fmt.Errorf("invalid chain: %s", chainID))
		return
	}
	percentage, err := strconv.ParseFloat(c.Params("percentage"), 64)
	if err != nil {
		s.log.Error("error parse percentage", err)
		return
	}

	source, id, err := s.service.NewSubscribe(chainID, percentage)
	if err != nil {
		s.log.Error("error subscribe", err)
		return
	}
	defer s.service.UnSubscribe(chainID, percentage, id)
	for ex := range source {
		if err = c.WriteJSON(ex); err != nil {
			s.log.Error("error write to socket", err)
			break
		}
	}
}
