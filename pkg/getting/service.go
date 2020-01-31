package getting

import "fmt"

// Service is the interface providing the GetCustomer method.
//
// Types implementing Service interface are able to orchestrate the retrieval and decryption of sensitive data
// from a data store
type Service interface {
	GetCustomer(int64) (*Customer, error)
}

// Repository is the interface providing the GetCustomer method.
//
// Types implementing Repository interface are able to get customer data from a data store
type Repository interface {
	GetCustomer(int64) (*Customer, error)
}

// Decrypter is the interface providing the Encrypt method.
//
// Types implementing Decrypt interface are able to decrypt an encrypted input
type Decrypter interface {
	Decrypt(input string) (string, error)
}

type service struct {
	cR Repository
	d  Decrypter
}

// NewService creates a getting service with the necessary dependencies
func NewService(r Repository, d Decrypter) Service {
	return &service{cR:r, d: d}
}

// GetCustomer returns a customer from the database
func (s *service) GetCustomer(id int64) (*Customer, error){
	c, err := s.cR.GetCustomer(id)
	if err != nil {
		return nil, fmt.Errorf("fetching customer %d failed: %w", id, err)
	}

	dacc, err := s.d.Decrypt(string(c.AccountNumber))
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return &Customer{
		ID:            c.ID,
		Name:          c.Name,
		AccountNumber: AccountNumber(dacc),
	}, nil
}