// Package kalshiperpsws is a Go client for the Kalshi perps WebSocket API,
// generated with gen-asyncapi from the AsyncAPI document Kalshi publishes.
package kalshiperpsws

//go:generate bash ../scripts/fetch-spec.sh perps_asyncapi.yaml
//go:generate bash ../scripts/match-client-spec.sh perps_asyncapi.yaml
//go:generate go tool gen-asyncapi --config=config.yaml ../api/perps_asyncapi.client.yaml
