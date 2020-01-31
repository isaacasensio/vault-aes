package getting

import (
	"encoding/base64"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type FailingDecrypter struct {
}

func (f FailingDecrypter) Decrypt(input string) (string, error) {
	return "", errors.New("something went wrong")
}

type Base64Decrypter struct {
}

func (b Base64Decrypter) Decrypt(input string) (string, error) {
	decodeString, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(decodeString), nil
}

type FailingCustomerRepo struct {
}

func (f FailingCustomerRepo) GetCustomer(int64) (*Customer, error) {
	return nil, errors.New("getting customer failed")
}

type CustomerRepo struct {
	id int64
}

func (c *CustomerRepo) GetCustomer(id int64) (*Customer, error) {
	c.id = id
	return &Customer{
		ID:            id,
		Name:          "Peter Parker",
		AccountNumber: AccountNumber(base64.StdEncoding.EncodeToString([]byte("12345"))),
	}, nil
}

func TestService_GetCustomerReturnsErrorWhenDecryptionFails(t *testing.T) {
	s := NewService(&CustomerRepo{}, FailingDecrypter{})

	c, err := s.GetCustomer(1)

	assert.EqualError(t, err, "decryption failed: something went wrong")
	assert.Nil(t, c)
}

func TestService_GetCustomerReturnsErrorOnDatabaseAccessError(t *testing.T) {
	s := NewService(FailingCustomerRepo{}, Base64Decrypter{})

	c, err := s.GetCustomer(1)

	assert.EqualError(t, err, "fetching customer 1 failed: getting customer failed")
	assert.Nil(t, c)
}

func TestService_GetCustomerReturnsIdOnSuccessfulSave(t *testing.T) {
	repo := CustomerRepo{}
	s := NewService(&repo, Base64Decrypter{})

	c, err := s.GetCustomer(1)

	require.NoError(t, err)
	assert.Equal(t, repo.id, c.ID)
	assert.Equal(t, c.Name, "Peter Parker")
	assert.Equal(t, string(c.AccountNumber),"12345")
}
