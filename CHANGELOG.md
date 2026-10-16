# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.0, 2026-10-16

- Added `kalshi` (repo root, `github.com/psyb0t/go-kalshi`), a REST client generated with oapi-codegen from Kalshi's OpenAPI document (spec version 3.34.0).
- Added `kalshiws`, a market data WebSocket client generated with gen-asyncapi from Kalshi's AsyncAPI document (spec version 2.0.0): typed commands, a `Handler` per push and response type, strict decoding, and a gorilla/websocket transport through `gorillaws.Dial`.
- Added `kalshiperpsws`, the same for Kalshi's perps WebSocket document (spec version 2.0.0).
- Added `kalshiauth`, which signs REST requests and the WebSocket handshake with an RSA-PSS or Ed25519 API key.
- `make generate` downloads the current specs from docs.kalshi.com and regenerates the three packages. The specs are not committed.
