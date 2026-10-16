package kalshiperpsws_test

import (
	"testing"

	"github.com/psyb0t/gen-asyncapi/asyncapirt"
	kalshiperpsws "github.com/psyb0t/go-kalshi/perpsws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// From the order_group_updates example in Kalshi's perps AsyncAPI document.
const orderGroupFrame = `{
	"type": "order_group_updates",
	"sending_ts_ms": 1700000000246,
	"sid": 21,
	"seq": 7,
	"msg": {
		"event_type": "limit_updated",
		"order_group_id": "og_123",
		"contracts_limit_fp": "150.00",
		"ts_ms": 1700000000123
	}
}`

func TestDecodeReceivedMessageOrderGroupUpdate(t *testing.T) {
	t.Parallel()

	message, err := kalshiperpsws.DecodeReceivedMessage([]byte(orderGroupFrame))
	require.NoError(t, err)
	assert.Equal(t, kalshiperpsws.MessageOrderGroupUpdates, message.Name)

	update, ok := message.Payload.(kalshiperpsws.OrderGroupUpdatesPayload)
	require.True(t, ok)
	assert.Equal(t, "og_123", update.Msg.OrderGroupID)
}

func TestDecodeReceivedMessageRejectsAnUnlistedType(t *testing.T) {
	t.Parallel()

	_, err := kalshiperpsws.DecodeReceivedMessage(
		[]byte(`{"type":"unlisted","sid":1,"msg":{}}`),
	)
	require.ErrorIs(t, err, asyncapirt.ErrUnknownMessage)
}
