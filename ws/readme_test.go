package kalshiws_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/psyb0t/gen-asyncapi/asyncapirt/gorillaws"
	kalshiauth "github.com/psyb0t/go-kalshi/auth"
	kalshiws "github.com/psyb0t/go-kalshi/ws"
)

type handler struct {
	kalshiws.UnimplementedHandler
}

func (handler) HandleTicker(
	ctx context.Context,
	ticker kalshiws.TickerPayload,
) error {
	slog.InfoContext(ctx, "ticker",
		"market", ticker.Msg.MarketTicker,
		"price", ticker.Msg.PriceDollars,
	)

	return nil
}

// readmeWS is the README's WebSocket example with main renamed. It is
// compiled, not run, so the README can't drift from the generated API.
var _ = readmeWS

func readmeWS() {
	ctx := context.Background()
	pemKey := []byte(os.Getenv("KALSHI_PRIVATE_KEY"))

	key, err := kalshiauth.ParsePrivateKeyPEM(pemKey)
	if err != nil {
		panic(err)
	}

	signer, err := kalshiauth.NewSigner(os.Getenv("KALSHI_API_KEY_ID"), key)
	if err != nil {
		panic(err)
	}

	header, err := signer.Headers(http.MethodGet, kalshiauth.WebSocketPath)
	if err != nil {
		panic(err)
	}

	options := make([]gorillaws.Option, 0, len(header))
	for name := range header {
		options = append(options, gorillaws.WithHeader(name, header.Get(name)))
	}

	conn, err := gorillaws.Dial(ctx, kalshiws.ServerProductionURL, options...)
	if err != nil {
		panic(err)
	}

	client := kalshiws.NewClient(conn)

	defer func() {
		if err := client.Close(); err != nil {
			slog.WarnContext(ctx, "close", "error", err)
		}
	}()

	err = client.SendSubscribe(ctx, kalshiws.SubscribeCommandPayload{
		ID: 1,
		Params: kalshiws.SubscribeCommandPayloadParams{
			Channels: []kalshiws.SubscribeCommandPayloadParamsChannelsItem{
				kalshiws.SubscribeCommandPayloadParamsChannelsItemTicker,
			},
			MarketTickers: []kalshiws.MarketTicker{"HIGHNY-22DEC23-B53.5"},
		},
	})
	if err != nil {
		panic(err)
	}

	if err := client.Run(ctx, handler{}); err != nil {
		panic(err)
	}
}
