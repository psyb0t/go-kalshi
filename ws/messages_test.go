package kalshiws_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/psyb0t/ctxerrors/commerr"
	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	kalshiws "github.com/psyb0t/go-kalshi/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The frames in testdata follow the examples in Kalshi's AsyncAPI document.
// Their tickers carry a dot on purpose: the published ticker pattern rejects
// them, and scripts/match-client-spec.sh drops it.
const (
	tickerFile     = "ticker.json"
	snapshotFile   = "orderbook_snapshot.json"
	fillFile       = "fill.json"
	rfqFile        = "rfq_created.json"
	errorFile      = "error.json"
	subscribedFile = "subscribed.json"

	fedTicker   = "FED-23DEC-T3.00"
	subscribeID = 1
)

func readFrame(t *testing.T, name string) []byte {
	t.Helper()

	frame, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)

	return frame
}

func TestDecodeReceivedMessagePicksTheType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		file     string
		wantName string
	}{
		{tickerFile, kalshiws.MessageTicker},
		{snapshotFile, kalshiws.MessageOrderbookSnapshot},
		{fillFile, kalshiws.MessageFill},
		{rfqFile, kalshiws.MessageRfqCreated},
		{errorFile, kalshiws.MessageErrorResponse},
		{subscribedFile, kalshiws.MessageSubscribedResponse},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			t.Parallel()

			message, err := kalshiws.DecodeReceivedMessage(
				readFrame(t, tc.file),
			)
			require.NoError(t, err)
			assert.Equal(t, tc.wantName, message.Name)
		})
	}
}

func TestTickerDecodesTheFields(t *testing.T) {
	t.Parallel()

	message, err := kalshiws.DecodeReceivedMessage(readFrame(t, tickerFile))
	require.NoError(t, err)

	ticker, ok := message.Payload.(kalshiws.TickerPayload)
	require.True(t, ok)
	assert.Equal(t, int64(11), ticker.Sid)
	assert.Equal(t, fedTicker, ticker.Msg.MarketTicker)
	assert.Equal(t, "0.4800", ticker.Msg.PriceDollars)
	assert.Equal(t, int64(16948), ticker.Msg.DollarVolume)
}

// The rfq_created frame leaves out the optional mve_selected_legs list, which
// the schema bounds with minItems: 1. An absent list must still decode.
func TestRfqCreatedDecodesWithoutOptionalLegs(t *testing.T) {
	t.Parallel()

	message, err := kalshiws.DecodeReceivedMessage(readFrame(t, rfqFile))
	require.NoError(t, err)

	rfq, ok := message.Payload.(kalshiws.RfqCreatedPayload)
	require.True(t, ok)
	assert.Equal(t, fedTicker, rfq.Msg.MarketTicker)
	assert.Nil(t, rfq.Msg.MveSelectedLegs)
}

func TestOrderbookSnapshotDecodesThePriceLevels(t *testing.T) {
	t.Parallel()

	message, err := kalshiws.DecodeReceivedMessage(readFrame(t, snapshotFile))
	require.NoError(t, err)

	snapshot, ok := message.Payload.(kalshiws.OrderbookSnapshotPayload)
	require.True(t, ok)
	require.Len(t, snapshot.Msg.YesDollarsFp, 2)
	assert.Equal(
		t,
		[]string{"0.0800", "300.00"},
		snapshot.Msg.YesDollarsFp[0],
	)
}

func TestErrorResponseDecodesTheCode(t *testing.T) {
	t.Parallel()

	message, err := kalshiws.DecodeReceivedMessage(readFrame(t, errorFile))
	require.NoError(t, err)

	response, ok := message.Payload.(kalshiws.ErrorResponsePayload)
	require.True(t, ok)
	assert.Equal(
		t,
		kalshiws.ErrorResponsePayloadMsgCode8,
		response.Msg.Code,
	)
}

func TestDecodeReceivedMessageRejectsUnknownFrames(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		frame   string
		wantErr error
	}{
		{
			name:    "not JSON",
			frame:   "{",
			wantErr: commerr.ErrCouldNotDecode,
		},
		{
			name:    "unlisted type",
			frame:   `{"type":"unlisted","sid":1,"msg":{}}`,
			wantErr: asyncapirt.ErrUnknownMessage,
		},
		{
			name:    "ticker missing its msg",
			frame:   `{"type":"ticker","sid":1}`,
			wantErr: asyncapirt.ErrUnknownMessage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := kalshiws.DecodeReceivedMessage([]byte(tc.frame))
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

type recordingTransport struct {
	sent [][]byte
}

func (r *recordingTransport) WriteMessage(
	_ context.Context,
	data []byte,
) error {
	r.sent = append(r.sent, data)

	return nil
}

func (r *recordingTransport) ReadMessage(context.Context) ([]byte, error) {
	return nil, context.Canceled
}

func (r *recordingTransport) Close() error {
	return nil
}

func TestClientSendSubscribeWritesTheCommand(t *testing.T) {
	t.Parallel()

	transport := &recordingTransport{}
	client := kalshiws.NewClient(transport)

	err := client.SendSubscribe(
		context.Background(),
		kalshiws.SubscribeCommandPayload{
			ID: subscribeID,
			Params: kalshiws.SubscribeCommandPayloadParams{
				Channels: []kalshiws.SubscribeCommandPayloadParamsChannelsItem{
					kalshiws.SubscribeCommandPayloadParamsChannelsItemTicker,
				},
				MarketTickers: []kalshiws.MarketTicker{fedTicker},
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, transport.sent, 1)

	var sent map[string]any
	require.NoError(t, json.Unmarshal(transport.sent[0], &sent))
	assert.Equal(t, "subscribe", sent["cmd"])
	assert.InDelta(t, subscribeID, sent["id"], 0)
	assert.Equal(t, map[string]any{
		"channels":       []any{"ticker"},
		"market_tickers": []any{fedTicker},
	}, sent["params"])
}
