// Package kalshi is a Go client for the Kalshi Trade REST API, generated with
// oapi-codegen from the OpenAPI document Kalshi publishes.
package kalshi

//go:generate bash scripts/fetch-spec.sh openapi.yaml
//go:generate go tool oapi-codegen --config=config.yaml api/openapi.yaml
