# go-kalshi

[![Go Reference](https://pkg.go.dev/badge/github.com/psyb0t/go-kalshi.svg)](https://pkg.go.dev/github.com/psyb0t/go-kalshi)
[![CI](https://github.com/psyb0t/go-kalshi/actions/workflows/pipeline.yml/badge.svg?branch=main)](https://github.com/psyb0t/go-kalshi/actions/workflows/pipeline.yml)
[![version](https://raw.githubusercontent.com/psyb0t/go-kalshi/badges/version.svg)](https://github.com/psyb0t/go-kalshi/tags)
[![license](https://raw.githubusercontent.com/psyb0t/go-kalshi/badges/license.svg)](LICENSE)
[![imported by](https://raw.githubusercontent.com/psyb0t/go-kalshi/badges/importers.svg)](https://github.com/psyb0t/go-kalshi/blob/badges/importers.md)

Unofficial Go client for [Kalshi](https://docs.kalshi.com), generated from the API specs Kalshi publishes. I'm not affiliated with Kalshi.

| Directory | Package | API | Generator |
|---|---|---|---|
| `/` (repo root) | `kalshi` | Trade REST API, `https://external-api.kalshi.com/trade-api/v2` | [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) |
| `ws/` | `kalshiws` | Market data WebSocket, `wss://external-api-ws.kalshi.com/trade-api/ws/v2` | [gen-asyncapi](https://github.com/psyb0t/gen-asyncapi) |
| `perpsws/` | `kalshiperpsws` | Perps WebSocket, `wss://external-api-margin-ws.kalshi.com/trade-api/ws/v2/margin` | [gen-asyncapi](https://github.com/psyb0t/gen-asyncapi) |
| `auth/` | `kalshiauth` | Request signing for all of the above | hand-written |

The REST client lives in the repo root; the WebSocket clients and the signer are subpackages. `client.gen.go` in the root and everything in `ws/` and `perpsws/` except `gen.go`, the generator configs and the tests is generated. Don't edit it by hand, regenerate it.

## Contents

- [Install](#install)
- [Auth](#auth)
- [REST](#rest)
- [WebSocket](#websocket)
- [Regenerating](#regenerating)
- [Agent integrations](#agent-integrations)
- [Development](#development)
- [License](#license)

## Install

```sh
go get github.com/psyb0t/go-kalshi
```

## Auth

Kalshi signs requests instead of sending a bearer token. Every private call carries `KALSHI-ACCESS-KEY` (the API key ID), `KALSHI-ACCESS-TIMESTAMP` (milliseconds) and `KALSHI-ACCESS-SIGNATURE`, a base64 RSA-PSS SHA-256 signature of timestamp + method + path, with the path taken without its query string. Ed25519 keys work too.

`kalshiauth.ParsePrivateKeyPEM` reads the key file Kalshi gives you. `kalshiauth.NewSigner` wraps it: `RequestEditor` signs REST requests, and `Headers` returns the headers for the WebSocket handshake.

## REST

Every endpoint in the spec has a typed method on `ClientWithResponses`. Public endpoints such as `GetMarketsWithResponse` work without the signer.

```go
package main

import (
	"context"
	"os"

	"github.com/psyb0t/go-kalshi"
	"github.com/psyb0t/go-kalshi/auth"
)

func main() {
	pemKey := []byte(os.Getenv("KALSHI_PRIVATE_KEY"))

	key, err := kalshiauth.ParsePrivateKeyPEM(pemKey)
	if err != nil {
		panic(err)
	}

	signer, err := kalshiauth.NewSigner(os.Getenv("KALSHI_API_KEY_ID"), key)
	if err != nil {
		panic(err)
	}

	client, err := kalshi.NewClientWithResponses(
		"https://external-api.kalshi.com/trade-api/v2",
		kalshi.WithRequestEditorFn(signer.RequestEditor),
	)
	if err != nil {
		panic(err)
	}

	res, err := client.GetBalanceWithResponse(
		context.Background(),
		&kalshi.GetBalanceParams{},
	)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
```

The demo environment lives at `https://external-api.demo.kalshi.co/trade-api/v2`.

The full API, with the spec's descriptions as doc comments, is in the [Go reference](https://pkg.go.dev/github.com/psyb0t/go-kalshi) and in `go doc github.com/psyb0t/go-kalshi` (add `/ws`, `/perpsws` or `/auth` for the subpackages).

## WebSocket

`kalshiws` sends the subscribe, unsubscribe, list and update commands, and decodes the orderbook, ticker, trade, fill, position, lifecycle, RFQ, order and index pushes plus the command responses. Every message is checked against Kalshi's schema both ways. `kalshiperpsws` does the same for the perps channels.

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/psyb0t/gen-asyncapi/asyncapirt/gorillaws"
	"github.com/psyb0t/go-kalshi/auth"
	"github.com/psyb0t/go-kalshi/ws"
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

func main() {
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
```

Pushes are told apart by their `type` field. A frame with a type the spec doesn't list fails with `asyncapirt.ErrUnknownMessage` and goes to `HandleUnknownMessage`. Kalshi's heartbeat is a WebSocket ping control frame, which gorilla/websocket answers on its own, so the `ping` and `pong` messages in the spec never reach the handler.

## Regenerating

Kalshi's specs are not committed. `make generate` downloads the current OpenAPI and AsyncAPI documents into the gitignored `api/` directory and regenerates all three packages:

```sh
make generate
```

Each package's `gen.go` declares its own steps, so `make generate` is just `go generate ./...`:

- The root `gen.go` downloads `openapi.yaml` and runs oapi-codegen with `config.yaml` and `overlay.yaml`. The overlay turns one optional `time.Time` field from `omitempty` into `omitzero`.
- `ws/gen.go` and `perpsws/gen.go` download `asyncapi.yaml` and `perps_asyncapi.yaml`, run `scripts/match-client-spec.sh`, and run gen-asyncapi. Kalshi's AsyncAPI documents describe the server, so the match step swaps every operation's `send` and `receive` to get the client side. It also drops the market ticker pattern `^[A-Z0-9-]+$`, which rejects real tickers such as `FED-23DEC-T3.00`.

`.github/workflows/regen.yml` does the same every day in CI. When the new spec changes the client, it runs `make lint`, `make test` and `make sec`. If a check fails, Claude edits the code and the checks run again, up to three times, and a review step reads the result. A passing change is committed with a CHANGELOG entry and tagged as the next release: a patch version for fixes and additions, a minor version for breaking changes.

## Agent integrations

The [skill](.agents/skills/go-kalshi) teaches a coding agent how to use go-kalshi when it writes Go. It works in any agent that reads `.agents/skills/`, and installs natively in the clients below.

### Claude Code

```bash
claude plugin marketplace add psyb0t/agents
claude plugin install go-kalshi@psyb0t
```

### Codex

```bash
codex plugin marketplace add psyb0t/agents
codex plugin add go-kalshi@psyb0t
```

Codex also picks the skill up with no install in any checkout of this repo.

### OpenClaw

The skill is published to ClawHub on every release:

```bash
openclaw skills install @psyb0t/go-kalshi
```

## Development

Every command runs in the `Dockerfile.dev` image. The host needs only Docker and Make.

```sh
make help      # list targets
make dep       # go mod tidy + vendor
make lint      # shellcheck, shfmt, go fix, golangci-lint
make test      # go test -race
make sec       # gitleaks, govulncheck, semgrep
make generate  # download the specs and regenerate
```

## License

[MIT](LICENSE). The generated code comes from Kalshi's published API specifications. Kalshi labels its AsyncAPI documents proprietary, so the specifications themselves are not redistributed here.
