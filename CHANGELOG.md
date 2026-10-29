# Changelog

All notable changes per release. Versions follow [semver](https://semver.org).

## v0.6.0, 2026-10-29

- The REST client adds position exit trigger (take-profit and stop-loss bracket) endpoints: `GetEventContractExitTriggers` (`GET /portfolio/positions/{ticker}/exit_trigger`), `SetEventContractExitTrigger` (`PUT`) and `DeleteEventContractExitTrigger` (`DELETE`). `SetEventContractExitTrigger` also has a `WithBody` variant.
- The REST client adds `CreateOrderWithExitTrigger` and `CreateOrderWithExitTriggerWithBody` (`POST /portfolio/events/orders/with_exit_trigger`). Each of the five methods has a `WithResponse` variant on `ClientWithResponses`, plus matching `New...Request` and `Parse...Response` functions.
- New REST types: `EventContractExitTrigger`, `AttachedEventContractExitTrigger`, `EventContractExitSlippage`, `SetEventContractExitTriggerRequest`, `GetEventContractExitTriggersParams`, `SetEventContractExitTriggerParams`, `DeleteEventContractExitTriggerParams`, and the enums `EventContractExitTriggerKind`, `EventContractExitTriggerSide`, `EventContractExitTriggerStatus`, `EventContractExitTriggerTriggeredLeg` and `EventContractExitSlippageUnit` (`bps` or `cents`).
- Breaking: `ClientInterface` gains the six exit trigger methods above and `ClientWithResponsesInterface` gains the six matching `WithResponse` methods. Code that implements either interface, for example a mock, must add them. `Client` and `ClientWithResponses` already have them.
- `ws` and `perpsws` add an optional `UpdateIntervalMs` field to `SubscribeCommandPayloadParams`, with the type `SubscribeCommandPayloadParamsUpdateIntervalMs` and values `0` (real time, the default), `100` and `1000` for conflated updates. Set it on the first `subscribe`. A repeated `subscribe` with a different interval returns error code 11.
- `ws` and `perpsws` add the error code `ErrorResponsePayloadMsgCode29`.

## v0.5.2, 2026-10-28

- Doc comment only, no behavior change: the `ID` field comment on `perpsws.MarginCorporateAction` now says "Corporate action id" instead of "Corporate action notice id".

## v0.5.1, 2026-10-28

- Updated Semgrep in the development image to 1.182.1 and refreshed its locked dependencies. certifi 2026.10.23 drops the emSign root CAs whose intermediate keys were compromised (GHSA-6q48-6c7w-9gw2).
- Dependabot no longer opens pull requests for `mcp` or the OpenTelemetry packages in `requirements-semgrep.txt`. Semgrep pins their versions itself, so bumping one of them on its own cannot install.
- Updated kin-openapi, which the code generator uses, from 0.142.0 to 0.149.0. The newer version fixes several panics on malformed specs reported in GitHub security advisories.
- The security scan merges its two gitleaks runs into one SARIF run. GitHub code scanning rejects a report with two runs from the same tool, so no scan results had been uploaded.

## v0.5.0, 2026-10-27

- The REST client adds the credit history endpoint `GET /portfolio/credit_history`: `GetCreditHistory` and `GetCreditHistoryWithResponse`, plus `NewGetCreditHistoryRequest` and `ParseGetCreditHistoryResponse2`. It returns the member's credits, newest first. `min_ts` is inclusive and `max_ts` is exclusive, and both filter on credit creation time.
- New REST types: `CreditHistory` (fields `AmountDollars`, `CreatedTime`, `CreditId`, `Reason`, `Status`, `Type`), `CreditHistoryStatus` (values `CreditHistoryStatusApplied`, `CreditHistoryStatusFailed`, `CreditHistoryStatusPending` and `CreditHistoryStatusUnknown`), `GetCreditHistoryParams`, `GetCreditHistoryResponse` and `GetCreditHistoryResponse2`.
- Breaking: `ClientInterface` gains `GetCreditHistory` and `ClientWithResponsesInterface` gains `GetCreditHistoryWithResponse`. Code that implements either interface, for example a mock, must add them. `Client` and `ClientWithResponses` already have them.

## v0.4.1, 2026-10-27

- The development image is based on Debian (`golang:1.26.10-trixie`) instead of Alpine. Alpine keeps only the latest build of each package, so the exact pins broke the image build whenever a package was updated. System packages now come from the snapshot.debian.org date the base image was built from, which keeps them fixed for a given base image digest. jq 1.8.2, shellcheck 0.11.0 and shfmt 3.13.1 keep their versions, and Semgrep runs in its own virtual environment.

## v0.4.0, 2026-10-26

- Breaking: `ws.BatchCommunicationsAddressTemplate` is removed. The spec no longer has a separate `batch_communications` channel. RFQ batch events now come on the `communications` channel, so code that subscribes with this constant must use `CommunicationsAddressTemplate`.
- Breaking: `perpsws.Handler` now requires a `HandleMarketLifecycleV2` method. Types that implement `Handler` without embedding `UnimplementedHandler` must add it. `UnimplementedHandler` has a no-op version.
- `perpsws` adds a market lifecycle V2 channel: `MarketLifecycleV2AddressTemplate`, `SubscribeCommandPayloadParamsChannelsItemMarketLifecycleV2`, `MessageMarketLifecycleV2`, and the payload types `MarginMarketLifecycleV2Payload`, `MarginMarketLifecycleV2PayloadMsg` and `MarginMarketLifecycleV2PayloadMsgEventType` (values `MarginMarketLifecycleV2PayloadMsgEventTypeActivated` and `MarginMarketLifecycleV2PayloadMsgEventTypeDeactivated`).
- `RFQ` gains an optional `RfqBatchId` field. It holds the batch an RFQ is a child of. A child can be quoted like a single RFQ, and its best quote counts toward the batch's.
- `ws.RfqCreatedPayloadMsg` and `ws.RfqDeletedPayloadMsg` gain an optional `RfqBatchID` field.
- Doc comments: the `ws` `UserFilter` comment now says `self` also filters RFQ batch events. The `QuoteExecuted` batch ID comment and the creator ID comments describe quotes on child RFQs. The `CreateRFQBatch` and `CreateRFQBatchWithBody` docs say makers see each child as an ordinary RFQ.
- The REST spec moves to version 3.35.0.

## v0.3.2, 2026-10-25

- `CreateRFQBatchChild` has two new optional fields: `ContractsFp` (`*FixedPointCount`) and `TargetCostDollars` (`FixedPointDollars`). Each child of an RFQ batch can now carry its own size. Every child must still be sized the same way, by contracts or by target cost.
- `CreateRFQBatchRequest.TargetCostDollars` now acts as a default for children that set no size of their own. The `TargetCostExcludesFees` comment now says it is valid when the children are sized by target cost.
- `CreateRFQBatch` and `CreateRFQBatchWithBody` docs now say nothing trades unless every child can fill in full, and that the batch closes partially executed in the rare case that filling one child cancels a maker order another child needed.
- The transfer method docs now say subaccount-restricted API keys can only transfer between `event_contract` exchange indices, with `source_subaccount` and `destination_subaccount` both set to the key's subaccount.

## v0.3.1, 2026-10-23

- Doc comments only, no behavior change and no exported API change.
- RFQ batch docs: `CreateRFQBatchChild.RestRemainder` now says it must be false because a batch executes as one fill-or-kill. The `SubtraderId` comment notes FCM members cannot create RFQ batches yet.
- RFQ batch docs: the `CloseReason` and `RFQBatchQuoteStatus` comments now say `partially_executed` happens only in the rare case that filling one child cancelled a maker order another child needed. The `ws` batch execution payload comments and the RFQ batch channel notes say the children execute together, every child trades or none does.
- `GetBalance` and its variants now document that an API key restricted to a subaccount gets only that subaccount's balances across exchange indices.
- The `EventTicker` parameter comment now says the maximum is 100 instead of 10, and that events with no markets return no orders.
- Reworded the `CreateRFQBatch` and `ConfirmRFQBatchQuote` method comments, and the `ws` notification comments for RFQ batch execution.

## v0.3.0, 2026-10-21

- The REST client adds RFQ batch endpoints: `CreateRFQBatch`, `GetRFQBatches`, `GetRFQBatch`, `DeleteRFQBatch`, `AcceptRFQBatch`, `GetRFQBatchBestQuotes`, `GetRFQBatchQuotes`, `CreateRFQBatchQuote`, `DeleteRFQBatchQuote` and `ConfirmRFQBatchQuote`. Each has a `WithBody` variant where it takes a body and a `WithResponse` variant on `ClientWithResponses`, plus matching `New...Request` and `Parse...Response2` functions.
- New REST types: `RFQBatch`, `RFQBatchStatus`, `RFQBatchQuote`, `RFQBatchQuoteStatus`, `RFQBatchQuoteEntry`, `RFQBatchBestQuote`, `RFQBatchBestQuotes`, `RFQBatchAcceptLimit`, `CreateRFQBatchRequest`, `CreateRFQBatchChild`, `CreateRFQBatchQuoteRequest`, `AcceptRFQBatchRequest`, `GetRFQBatchesParams`, `RfqBatchIdPath` and `BatchQuoteIdPath`.
- Breaking: `ClientInterface` and `ClientWithResponsesInterface` gain the RFQ batch methods above. Code that implements either interface, for example a mock, must add them. `Client` and `ClientWithResponses` already have them.
- `ws` adds seven RFQ batch channel messages: `MessageRfqBatchCreated`, `MessageRfqBatchDeleted`, `MessageRfqBatchExecuted`, `MessageRfqBatchBestQuotes`, `MessageBatchQuoteCreated`, `MessageBatchQuoteAccepted` and `MessageBatchQuoteExecuted`, with payload types such as `RfqBatchCreatedPayload` and `BatchQuoteAcceptedPayload`, and the `BatchCommunicationsAddressTemplate` constant.
- Breaking: `ws.Handler` now requires `HandleRfqBatchCreated`, `HandleRfqBatchDeleted`, `HandleRfqBatchExecuted`, `HandleRfqBatchBestQuotes`, `HandleBatchQuoteCreated`, `HandleBatchQuoteAccepted` and `HandleBatchQuoteExecuted`. Types that implement `Handler` without embedding `UnimplementedHandler` must add them. `UnimplementedHandler` has no-op versions.
- `ws.OrderbookSnapshotPayloadMsg` and `ws.OrderbookDeltaPayloadMsg` gain a `Clock` field. `ws.QuoteExecutedPayloadMsg` gains `RfqBatchID`.
- The minimum Go version moves from 1.26.9 to 1.26.10 in `go.mod`, the dev image and the setup reference, to pick up the `crypto/x509` fix for GO-2026-6809.
- The `orderbook_snapshot.json` test fixture now includes `clock`, which the updated spec requires.

## v0.2.1, 2026-10-20

- Doc comments only, no behavior change: the `pyth_value` channel docs on `PythValueAddressTemplate`, `MessagePythValue` and `Handler.HandlePythValue` now say the channel also streams Treasury yields. The channel's table lists the new symbols `InterestRate.US2Y`, `InterestRate.US5Y`, `InterestRate.US10Y` and `InterestRate.US30Y`, which are included in `["all"]` subscriptions or can be requested individually.
- The `PythValuePayloadMsg.ValueUsd` comment now says that for Treasury yield symbols the value is a percent, so `"4.25000000"` means 4.25%. The field name and type are unchanged.
- The `TickerPayloadMsg.Time` comment notes the timestamp is truncated to milliseconds.

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
