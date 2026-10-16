package kalshiauth_test

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/psyb0t/ctxerrors/commerr"
	kalshiauth "github.com/psyb0t/go-kalshi/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKeyID      = "a952bcbe-ec3b-4b5b-b8f9-11dae589608c"
	testRSAKeyBits = 2048
	balancePath    = "/trade-api/v2/portfolio/balance"
	balanceURL     = "https://external-api.kalshi.com" + balancePath
	timestampSlack = time.Minute
	pemTypePKCS1   = "RSA PRIVATE KEY"
	pemTypePKCS8   = "PRIVATE KEY"
)

func newRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, testRSAKeyBits)
	require.NoError(t, err)

	return key
}

func signedText(header http.Header, method, path string) []byte {
	return []byte(header.Get(kalshiauth.HeaderAccessTimestamp) + method + path)
}

func decodeSignature(t *testing.T, header http.Header) []byte {
	t.Helper()

	signature, err := base64.StdEncoding.DecodeString(
		header.Get(kalshiauth.HeaderAccessSignature),
	)
	require.NoError(t, err)

	return signature
}

func verifyRSA(
	t *testing.T,
	key *rsa.PrivateKey,
	header http.Header,
	path string,
) {
	t.Helper()

	digest := sha256.Sum256(signedText(header, http.MethodGet, path))
	err := rsa.VerifyPSS(
		&key.PublicKey,
		crypto.SHA256,
		digest[:],
		decodeSignature(t, header),
		&rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
			Hash:       crypto.SHA256,
		},
	)
	require.NoError(t, err)
}

func TestHeadersSignWithRSAPSS(t *testing.T) {
	t.Parallel()

	key := newRSAKey(t)
	signer, err := kalshiauth.NewSigner(testKeyID, key)
	require.NoError(t, err)

	header, err := signer.Headers(http.MethodGet, balancePath)
	require.NoError(t, err)

	assert.Equal(t, testKeyID, header.Get(kalshiauth.HeaderAccessKey))

	millis, err := strconv.ParseInt(
		header.Get(kalshiauth.HeaderAccessTimestamp),
		10,
		64,
	)
	require.NoError(t, err)
	assert.WithinDuration(
		t,
		time.Now(),
		time.UnixMilli(millis),
		timestampSlack,
	)

	verifyRSA(t, key, header, balancePath)
}

func TestHeadersSignWithEd25519(t *testing.T) {
	t.Parallel()

	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := kalshiauth.NewSigner(testKeyID, private)
	require.NoError(t, err)

	header, err := signer.Headers(http.MethodGet, kalshiauth.WebSocketPath)
	require.NoError(t, err)

	text := signedText(header, http.MethodGet, kalshiauth.WebSocketPath)
	assert.True(t, ed25519.Verify(public, text, decodeSignature(t, header)))
}

func TestRequestEditorSignsThePathWithoutTheQuery(t *testing.T) {
	t.Parallel()

	key := newRSAKey(t)
	signer, err := kalshiauth.NewSigner(testKeyID, key)
	require.NoError(t, err)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		balanceURL+"?subaccount=1",
		nil,
	)
	require.NoError(t, signer.RequestEditor(req.Context(), req))

	verifyRSA(t, key, req.Header, balancePath)
}

func TestNewSignerRejectsBadInput(t *testing.T) {
	t.Parallel()

	_, err := kalshiauth.NewSigner("", newRSAKey(t))
	require.ErrorIs(t, err, commerr.ErrRequiredFieldNotSet)

	_, err = kalshiauth.NewSigner(testKeyID, nil)
	require.ErrorIs(t, err, commerr.ErrValidationFailed)
}

func TestParsePrivateKeyPEM(t *testing.T) {
	t.Parallel()

	rsaKey := newRSAKey(t)
	_, edKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pkcs8RSA, err := x509.MarshalPKCS8PrivateKey(rsaKey)
	require.NoError(t, err)
	pkcs8Ed, err := x509.MarshalPKCS8PrivateKey(edKey)
	require.NoError(t, err)

	testCases := []struct {
		name  string
		block *pem.Block
	}{
		{
			name: "pkcs1 rsa",
			block: &pem.Block{
				Type:  pemTypePKCS1,
				Bytes: x509.MarshalPKCS1PrivateKey(rsaKey),
			},
		},
		{
			name:  "pkcs8 rsa",
			block: &pem.Block{Type: pemTypePKCS8, Bytes: pkcs8RSA},
		},
		{
			name:  "pkcs8 ed25519",
			block: &pem.Block{Type: pemTypePKCS8, Bytes: pkcs8Ed},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			key, err := kalshiauth.ParsePrivateKeyPEM(
				pem.EncodeToMemory(tc.block),
			)
			require.NoError(t, err)

			_, err = kalshiauth.NewSigner(testKeyID, key)
			assert.NoError(t, err)
		})
	}
}

func TestParsePrivateKeyPEMRejectsBadInput(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		data    []byte
		wantErr error
	}{
		{
			name:    "not PEM",
			data:    []byte("not a key"),
			wantErr: commerr.ErrParseFailed,
		},
		{
			name: "certificate block",
			data: pem.EncodeToMemory(
				&pem.Block{Type: "CERTIFICATE", Bytes: []byte{1}},
			),
			wantErr: commerr.ErrParseFailed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := kalshiauth.ParsePrivateKeyPEM(tc.data)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}

	_, err := kalshiauth.ParsePrivateKeyPEM(pem.EncodeToMemory(
		&pem.Block{Type: pemTypePKCS1, Bytes: []byte{1}},
	))
	require.Error(t, err)
}
