# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.1.3, 2026-10-17

- Documentation only. The agent skill explains how to look up every endpoint, parameter and model with `go doc` or on pkg.go.dev, and declares its license.
- The skill no longer lists the API credentials under OpenClaw's `requires.env`, which kept OpenClaw from loading it until they were set. It requires only `go`; the credentials are still documented as what your program reads.
- The README points to the full API reference, and the examples in the README and the skill were checked against the current code.
- The setup reference lists the `HandleMarketLifecycleV2`, `HandleCfbenchmarksValue5Hz` and `HandleCfbenchmarks5HzIndexList` handlers and the error for unsupported key types.

## v0.1.2, 2026-10-17

- Documentation only. The README and `AGENTS.md` describe the daily regeneration workflow: what it checks, how it fixes failing checks, how it picks the next version and which paths a fix may not touch.

## v0.1.1, 2026-10-16

- Added the `go-kalshi` agent skill. The README's new Agent integrations section has the install commands for Claude Code, Codex and ClawHub.

## v0.1.0, 2026-10-16

- Added `kalshi` (repo root, `github.com/psyb0t/go-kalshi`), a REST client generated with oapi-codegen from Kalshi's OpenAPI document (spec version 3.34.0).
- Added `kalshiws`, a market data WebSocket client generated with gen-asyncapi from Kalshi's AsyncAPI document (spec version 2.0.0): typed commands, a `Handler` per push and response type, strict decoding, and a gorilla/websocket transport through `gorillaws.Dial`.
- Added `kalshiperpsws`, the same for Kalshi's perps WebSocket document (spec version 2.0.0).
- Added `kalshiauth`, which signs REST requests and the WebSocket handshake with an RSA-PSS or Ed25519 API key.
- `make generate` downloads the current specs from docs.kalshi.com and regenerates the three packages. The specs are not committed.
