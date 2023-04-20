package routes_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/abergasov/market_timer/internal/entities"
	"github.com/abergasov/market_timer/internal/testhelpers"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

const (
	address = "127.0.0.1:8000"
)

func TestServer_ETH(t *testing.T) {
	testhelpers.SpawnWebServer(t, "config.yaml", testhelpers.GetTestContext(t))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	signalChan := make(chan struct{})
	received := false

	go func() {
		ws, err := websocket.Dial(fmt.Sprintf("ws://%s/ws/eth/99", address), "", fmt.Sprintf("http://%s/", address))
		require.NoError(t, err, "unable to connect to websocket")
		for {
			var msg = make([]byte, 512)
			n, err := ws.Read(msg)
			require.NoError(t, err, "unable to read from websocket")
			var gr entities.GasRates
			require.NoError(t, json.Unmarshal(msg[:n], &gr))
			received = true
			signalChan <- struct{}{}
			break
		}
	}()

	select {
	case <-ctx.Done():
		t.Fatal("timeout")
	case <-signalChan:
		break
	}
	require.Truef(t, received, "no message received")
}
