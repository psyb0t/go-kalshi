package kalshi_test

import (
	"context"
	"os"

	"github.com/psyb0t/go-kalshi"
	kalshiauth "github.com/psyb0t/go-kalshi/auth"
)

// readmeREST is the README's REST example with main renamed. It is compiled,
// not run, so the README can't drift from the generated API.
var _ = readmeREST

func readmeREST() {
	pemKey := []byte(os.Getenv("KALSHI_PRIVATE_KEY"))

	key, err := kalshiauth.ParsePrivateKeyPEM(pemKey)
	if err != nil {
		panic(err)
	}

	signer, err := kalshiauth.NewSigner(os.Getenv("KALSHI_API_KEY_ID"), key)
	if err != nil {
		panic(err)
	}

	client, err := kalshi.NewClientWithResponses(
		"https://external-api.kalshi.com/trade-api/v2",
		kalshi.WithRequestEditorFn(signer.RequestEditor),
	)
	if err != nil {
		panic(err)
	}

	res, err := client.GetBalanceWithResponse(
		context.Background(),
		&kalshi.GetBalanceParams{},
	)
	if err != nil {
		panic(err)
	}

	_ = res.JSON200
}
