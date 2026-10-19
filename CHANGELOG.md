# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.2.0, 2026-10-19

- `IntraExchangeInstanceTransfer.ClientTransferId` field added. It holds the client-supplied transfer ID from when the transfer was created, omitted if none was provided.
- `IntraExchangeInstanceTransferResponse2.JSON409` field and `GetJSON409` method added, and `ParseIntraExchangeInstanceTransferResponse2` now decodes HTTP 409 into them as a `*ConflictError`. This is the response when a transfer request reuses a `ClientTransferId` that the same source exchange instance and shard already accepted: no new transfer is created, and the existing `transfer_id` is the one already assigned.
- `perpsws` adds a corporate actions channel: the `CorporateActionsAddressTemplate` constant, the `SubscribeCommandPayloadParamsChannelsItemCorporateActions` channel value, the `MessageCorporateActionsDelta` message name, and the payload types `MarginCorporateActionsDeltaPayload`, `MarginCorporateActionsDeltaPayloadMsg`, `MarginCorporateAction`, `MarginCorporateActionType` (values `MarginCorporateActionTypeStockSplit`, `MarginCorporateActionTypeFinalSettlement`, `MarginCorporateActionTypeTickerChange`), `MarginCorporateActionSplitTerms`, `MarginCorporateActionFinalSettlementTerms` and `MarginCorporateActionTickerChangeTerms`. It streams stock splits, final settlements and ticker changes applied to margin markets.
- Breaking: `perpsws.Handler` now requires a `HandleCorporateActionsDelta` method. Any type that implements `Handler` directly, without embedding `UnimplementedHandler`, no longer satisfies the interface and must add this method. `UnimplementedHandler` gets a no-op `HandleCorporateActionsDelta`, so code that embeds it is unaffected.
- Doc comments only, no behavior change: `Milestone.Details` and the `GetLiveDatas`, `GetLiveDataByMilestone`, `GetMilestones` and `GetMilestone` methods now explain that a milestone's own `details.status` lags the live game state, and that `live_data.details.status` from the live data endpoints is the more current source. The `IntraExchangeInstanceTransferRequest.ClientTransferId` comment documents the 409 retry behavior above. The `LastUpdateReason` comment and the WS `CreatedTime`, `LastUpdateTime` and `ExpirationTime` comments are reworded; the latter three note the timestamps are truncated to milliseconds.

## v0.1.4, 2026-10-17

- `perpsws` adds the exported `MarketID` type, an alias for `string`. `OkResponsePayloadMsg.MarketIds` now has type `[]MarketID` instead of `[]string`; code written against `[]string` still compiles and runs the same.
- The `Subaccount` parameter comment on `GetBalanceParams`, `CancelOrderV2Params`, `AmendOrderV2Params`, `DecreaseOrderV2Params`, `DeleteOrderGroupParams`, `UpdateOrderGroupLimitParams`, `ResetOrderGroupParams`, `TriggerOrderGroupParams`, `GetOrderQueuePositionsParams`, `GetPositionsParams` and `GetHistoricalPositionsParams` now explains that subaccount-restricted API keys default to their locked subaccount and must omit the parameter or pass that same subaccount.
- In `perpsws`, the `ID` field comment on the command and response payloads (for example `SubscribeCommandPayload`, `OkResponsePayload`, `ErrorResponsePayload`) now explains that the client generates the ID, that IDs should count up from 1, and that an ID of 0 is treated the same as no ID.
- The `Sid` field comment across the `perpsws` payloads now says it identifies the channel, and the `Seq` field comment now says to check it to confirm no messages were missed.
- `UnsubscribeCommandPayloadParams.Sids`, `OkResponsePayloadMsg.MarketTickers`, `ListSubscriptionsResponsePayload.Msg` and `ListSubscriptionsResponsePayloadMsgItem.Channel` each get a new doc comment describing the field.

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
