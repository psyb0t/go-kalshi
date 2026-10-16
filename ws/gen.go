// Package kalshiws is a Go client for the Kalshi market data WebSocket API,
// generated with gen-asyncapi from the AsyncAPI document Kalshi publishes.
package kalshiws

//go:generate bash ../scripts/fetch-spec.sh asyncapi.yaml
//go:generate bash ../scripts/match-client-spec.sh asyncapi.yaml
//go:generate go tool gen-asyncapi --config=config.yaml ../api/asyncapi.client.yaml
