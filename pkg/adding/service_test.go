package adding

import (
	"encoding/base64"
	"errors"
	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
	"testing"
)

type FailingEncrypter struct {
}

type Base64Encrypter struct {
}

type FailingCustomerRepo struct {
}

type CustomerRepo struct {
	Customer Customer
}

func (c *CustomerRepo) AddCustomer(cs Customer) (int64, error) {
	c.Customer = Customer{
		Name:          cs.Name,
		AccountNumber: cs.AccountNumber,
	}
	return 1, nil
}

func (c FailingCustomerRepo) AddCustomer(Customer) (int64, error) {
	return 0, errors.New("saving customer failed")
}

func (f FailingEncrypter) Encrypt(input string) (string, error) {
	return "", errors.New("something went wrong")
}

func (f Base64Encrypter) Encrypt(input string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(input)), nil
}

func TestService_AddCustomerReturnsErrorWhenEncryptionFails(t *testing.T) {
	s := NewService(&CustomerRepo{}, FailingEncrypter{})

	id, err := s.AddCustomer(Customer{
		Name:          "Test",
		AccountNumber: "12345",
	})

	assert.Error(t, err, "encryption failed: something went wrong")
	assert.Equal(t, int64(0), id)
}

func TestService_AddCustomerReturnsErrorOnDatabaseAccessError(t *testing.T) {
	s := NewService(FailingCustomerRepo{}, Base64Encrypter{})

	id, err := s.AddCustomer(Customer{
		Name:          "Test",
		AccountNumber: "12345",
	})

	assert.Error(t, err, "saving customer failed")
	assert.Equal(t, int64(0), id)
}

func TestService_AddCustomerReturnsIdOnSuccessfulSave(t *testing.T) {
	repo := CustomerRepo{}
	s := NewService(&repo, Base64Encrypter{})

	customer := Customer{
		Name:          "Test",
		AccountNumber: "12345",
	}
	id, err := s.AddCustomer(customer)

	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	assert.Equal(t, string(repo.Customer.AccountNumber), base64.StdEncoding.EncodeToString([]byte(string(customer.AccountNumber))))
	assert.Equal(t, repo.Customer.Name, customer.Name)
}
