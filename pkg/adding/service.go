package adding

import "fmt"

// Service is the interface providing the AddCustomer method.
//
// Types implementing Service interface are able to orchestrate the encryption and storage of sensitive data
// into a data store
type Service interface {
	AddCustomer(Customer) (int64, error)
}

// Repository is the interface providing the AddCustomer method.
//
// Types implementing Repository interface are able to store customer data into a data store
type Repository interface {
	AddCustomer(Customer) (int64, error)
}

// Encrypter is the interface providing the Encrypt method.
//
// Types implementing Encrypt interface are able to encrypt an input
type Encrypter interface {
	Encrypt(input string) (string, error)
}

type service struct {
	cR Repository
	e  Encrypter
}

// NewService creates an adding service with the necessary dependencies
func NewService(r Repository, e Encrypter) Service {
	return &service{cR: r, e: e}
}

// AddCustomer adds a customer to the database
func (s *service) AddCustomer(c Customer) (int64, error) {
	accEncrypted, err := s.e.Encrypt(string(c.AccountNumber))
	if err != nil {
		return 0, fmt.Errorf("encryption failed: %w", err)
	}

	ec := Customer{
		Name:          c.Name,
		AccountNumber: AccountNumber(accEncrypted),
	}

	return s.cR.AddCustomer(ec)
}
