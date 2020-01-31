package encryption

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

// EncrypterDecrypter encrypts and decrypts
type EncrypterDecrypter interface {
	Encrypt(input string) (string, error)
	Decrypt(encryptedText string) (string, error)
}

type client struct {
	vc    *api.Client
	keyName string
}

func (c *client) Encrypt(input string) (string, error) {
	path := fmt.Sprintf("transit/encrypt/%s", c.keyName)
	data := map[string]interface{} {
		"plaintext" : base64.StdEncoding.EncodeToString([]byte(input)),
	}

	secret, err := c.vc.Logical().Write(path, data)
	if err != nil {
		return "", errors.Wrap(err, "can't encrypt data")
	}

	if cipherText, ok := secret.Data["ciphertext"].(string); ok && cipherText != "" {
		return cipherText, nil
	}
	return "", errors.New("encryption failed. ciphertext is empty or not a string")
}

func (c *client) Decrypt(encryptedText string) (string, error) {
	path := fmt.Sprintf("transit/decrypt/%s", c.keyName)
	data := map[string]interface{} {
		"ciphertext" : encryptedText,
	}

	secret, err := c.vc.Logical().Write(path, data)
	if err != nil {
		return "", errors.Wrap(err, "can't decrypt data")
	}

	if plainText, ok := secret.Data["plaintext"].(string); ok && plainText != "" {
		decodeString, err := base64.StdEncoding.DecodeString(plainText)
		if err != nil {
			return "", err
		}
		return string(decodeString), nil
	}
	return "", errors.New("decryption failed. plaintext is empty or not a string")
}

// NewVaultClient creates a Vault client to interact with transit engine in order to encrypt/decrypt data
func NewVaultClient(addr, token, keyName string) (EncrypterDecrypter, error) {
	conf := api.DefaultConfig()
	conf.Address = addr

	apic, err := api.NewClient(conf)
	if err != nil {
		return nil, errors.Wrap(err, "can't initialize vault client")
	}
	apic.SetToken(token)

	c := client{
		vc:    apic,
		keyName: keyName,
	}

	return &c, nil
}
