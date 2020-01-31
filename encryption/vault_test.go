package encryption

import (
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/builtin/logical/transit"
	"github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestNewClient(t *testing.T) {
	_, err := NewVaultClient("\n", "token", "keyname")
	assert.Error(t, err, "can't initialize vault client")
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns plain text from encrypted text"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, token, addr := setupVaultServer(t)
			defer server.Close()

			client, err := NewVaultClient(addr, token, "keyname")
			assert.NoError(t, err)

			input := "foo"
			ciphertext, err := client.Encrypt(input)
			assert.NoError(t, err)
			assert.NotEqual(t, input, ciphertext)

			plaintext, err := client.Decrypt(ciphertext)
			assert.NoError(t, err)

			assert.Equal(t, input, plaintext)
		})
	}
}

func setupVaultServer(t *testing.T) (net.Listener, string, string) {
	t.Helper()

	// Create an in-memory, unsealed core with transient backend
	c := &vault.CoreConfig{
		Seal:            nil,
		EnableUI:        false,
		EnableRaw:       false,
		BuiltinRegistry: vault.NewMockBuiltinRegistry(),
		LogicalBackends: map[string]logical.Factory{
			"transit": transit.Factory,
		},
	}
	core, _, rootToken := vault.TestCoreUnsealedWithConfig(t, c)
	server, addr := http.TestServer(t, core)

	// Create a vc that talks to the server, initially authenticating with
	// the root token.
	client, err := createClient(addr, rootToken)
	if err != nil {
		t.Fatal(err)
	}

	err = enableTransit(client)
	if err != nil {
		t.Fatal(err)
	}

	token, err := createClientToken(client)
	if err != nil {
		t.Fatal(err)
	}

	return server, token, addr
}

func enableTransit(client *api.Client) error {
	return client.Sys().Mount("transit", &api.MountInput{
		Type: "transit",
	})
}

func createClient(addr, token string) (*api.Client, error) {
	conf := api.DefaultConfig()
	conf.Address = addr

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}
	client.SetToken(token)

	return client, nil
}

func createClientToken(c *api.Client) (string, error) {
	req := &api.TokenCreateRequest{
		Period: "5s",
	}
	rsp, err := c.Auth().Token().Create(req)
	if err != nil {
		return "", err
	}
	return rsp.Auth.ClientToken, nil
}
