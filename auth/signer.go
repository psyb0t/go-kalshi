// Package kalshiauth signs Kalshi API requests. Kalshi authenticates every
// private REST call and the WebSocket handshake with three headers: the API
// key ID, a millisecond timestamp, and a base64 signature of
// timestamp + method + path made with the key. RSA keys sign with RSA-PSS and
// SHA-256, Ed25519 keys sign the text as it is.
package kalshiauth

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
	"maps"
	"net/http"
	"strconv"
	"time"

	"github.com/psyb0t/ctxerrors"
	"github.com/psyb0t/ctxerrors/commerr"
)

// Header names Kalshi reads the credentials from.
const (
	HeaderAccessKey       = "KALSHI-ACCESS-KEY"
	HeaderAccessTimestamp = "KALSHI-ACCESS-TIMESTAMP"
	HeaderAccessSignature = "KALSHI-ACCESS-SIGNATURE"
)

// WebSocketPath is the path signed for the market data WebSocket handshake.
const WebSocketPath = "/trade-api/ws/v2"

const (
	pemTypePKCS1 = "RSA PRIVATE KEY"
	pemTypePKCS8 = "PRIVATE KEY"
)

// Signer adds Kalshi's authentication headers to requests.
type Signer struct {
	keyID string
	key   crypto.Signer
	now   func() time.Time
}

// NewSigner returns a Signer for the API key keyID. key must be an RSA or
// Ed25519 private key; anything else fails with commerr.ErrValidationFailed.
func NewSigner(keyID string, key crypto.Signer) (*Signer, error) {
	if keyID == "" {
		return nil, ctxerrors.Wrap(commerr.ErrRequiredFieldNotSet, "key ID")
	}

	switch key.(type) {
	case *rsa.PrivateKey, ed25519.PrivateKey:
	default:
		return nil, unsupportedKey(key)
	}

	return &Signer{keyID: keyID, key: key, now: time.Now}, nil
}

// ParsePrivateKeyPEM reads an RSA private key in PKCS #1 form, which is what
// Kalshi hands out, or an RSA or Ed25519 key in PKCS #8 form.
func ParsePrivateKeyPEM(data []byte) (crypto.Signer, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ctxerrors.Wrap(commerr.ErrParseFailed, "no PEM block")
	}

	switch block.Type {
	case pemTypePKCS1:
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, ctxerrors.Wrap(err, "parse PKCS #1 key")
		}

		return key, nil
	case pemTypePKCS8:
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, ctxerrors.Wrap(err, "parse PKCS #8 key")
		}

		signer, ok := key.(crypto.Signer)
		if !ok {
			return nil, unsupportedKey(key)
		}

		return signer, nil
	default:
		return nil, ctxerrors.Wrapf(
			commerr.ErrParseFailed,
			"unsupported PEM block %q",
			block.Type,
		)
	}
}

// Headers returns the three authentication headers for a request with the
// given method and path. path is the URL path without the query string, for
// example /trade-api/v2/portfolio/balance or WebSocketPath.
func (s *Signer) Headers(method, path string) (http.Header, error) {
	timestamp := strconv.FormatInt(s.now().UnixMilli(), 10)

	signature, err := s.sign(timestamp + method + path)
	if err != nil {
		return nil, err
	}

	header := make(http.Header)
	header.Set(HeaderAccessKey, s.keyID)
	header.Set(HeaderAccessTimestamp, timestamp)
	header.Set(HeaderAccessSignature, signature)

	return header, nil
}

// RequestEditor signs req in place. Its signature matches the generated
// kalshi.RequestEditorFn, so it plugs into kalshi.WithRequestEditorFn.
func (s *Signer) RequestEditor(_ context.Context, req *http.Request) error {
	header, err := s.Headers(req.Method, req.URL.Path)
	if err != nil {
		return err
	}

	maps.Copy(req.Header, header)

	return nil
}

func (s *Signer) sign(text string) (string, error) {
	var (
		signature []byte
		err       error
	)

	switch key := s.key.(type) {
	case *rsa.PrivateKey:
		digest := sha256.Sum256([]byte(text))
		signature, err = rsa.SignPSS(
			rand.Reader,
			key,
			crypto.SHA256,
			digest[:],
			&rsa.PSSOptions{
				SaltLength: rsa.PSSSaltLengthEqualsHash,
				Hash:       crypto.SHA256,
			},
		)
	case ed25519.PrivateKey:
		signature = ed25519.Sign(key, []byte(text))
	}

	if err != nil {
		return "", ctxerrors.Wrap(err, "sign request")
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func unsupportedKey(key any) error {
	return ctxerrors.Wrapf(
		commerr.ErrValidationFailed,
		"unsupported key type %T",
		key,
	)
}
