package kalshi_test

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/psyb0t/go-kalshi"
	kalshiauth "github.com/psyb0t/go-kalshi/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKeyID      = "a952bcbe-ec3b-4b5b-b8f9-11dae589608c"
	testRSAKeyBits = 2048
	apiBasePath    = "/trade-api/v2"
	balancePath    = apiBasePath + "/portfolio/balance"
	testSubaccount = 2
	balanceBody    = `{
		"balance": 12345,
		"balance_dollars": "123.45",
		"portfolio_value": 500,
		"updated_ts": 1733047200
	}`
)

type capturedRequest struct {
	method string
	path   string
	query  string
	header http.Header
}

func newBalanceServer(
	t *testing.T,
) (*httptest.Server, <-chan capturedRequest) {
	t.Helper()

	captured := make(chan capturedRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			captured <- capturedRequest{
				method: r.Method,
				path:   r.URL.Path,
				query:  r.URL.RawQuery,
				header: r.Header.Clone(),
			}

			w.Header().Set("Content-Type", "application/json")

			_, err := w.Write([]byte(balanceBody))
			assert.NoError(t, err)
		},
	))
	t.Cleanup(server.Close)

	return server, captured
}

func TestGetBalanceSendsASignedRequest(t *testing.T) {
	t.Parallel()

	key, err := rsa.GenerateKey(rand.Reader, testRSAKeyBits)
	require.NoError(t, err)

	signer, err := kalshiauth.NewSigner(testKeyID, key)
	require.NoError(t, err)

	server, captured := newBalanceServer(t)

	client, err := kalshi.NewClientWithResponses(
		server.URL+apiBasePath,
		kalshi.WithRequestEditorFn(signer.RequestEditor),
	)
	require.NoError(t, err)

	subaccount := testSubaccount
	res, err := client.GetBalanceWithResponse(
		context.Background(),
		&kalshi.GetBalanceParams{Subaccount: &subaccount},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode())
	require.NotNil(t, res.JSON200)
	assert.Equal(t, int64(12345), res.JSON200.Balance)
	assert.Equal(t, "123.45", res.JSON200.BalanceDollars)

	req := <-captured
	assert.Equal(t, http.MethodGet, req.method)
	assert.Equal(t, balancePath, req.path)
	assert.Equal(t, "subaccount=2", req.query)
	assert.Equal(t, testKeyID, req.header.Get(kalshiauth.HeaderAccessKey))

	signature, err := base64.StdEncoding.DecodeString(
		req.header.Get(kalshiauth.HeaderAccessSignature),
	)
	require.NoError(t, err)

	timestamp := req.header.Get(kalshiauth.HeaderAccessTimestamp)
	digest := sha256.Sum256([]byte(timestamp + http.MethodGet + balancePath))
	assert.NoError(t, rsa.VerifyPSS(
		&key.PublicKey,
		crypto.SHA256,
		digest[:],
		signature,
		nil,
	))
}
