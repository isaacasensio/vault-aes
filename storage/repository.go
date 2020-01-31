package storage

import (
	"database/sql"
	"github.com/isaacasensio/vault-eas/pkg/adding"
	"github.com/isaacasensio/vault-eas/pkg/getting"
	_ "github.com/lib/pq" //required to load pq driver
)
// CustomerRepository is the interface providing the AddCustomer and GetCustomer methods.
// Types implementing CustomerRepository interface are able to store and retrieve customer data from a data store
type CustomerRepository interface {
	AddCustomer(c adding.Customer) (int64, error)
	GetCustomer(id int64) (*getting.Customer, error)
}

type psqlDataStore struct {
	db *sql.DB
}

func (d *psqlDataStore) AddCustomer(c adding.Customer) (int64, error) {
	var id int64
	err := d.db.QueryRow("INSERT INTO customer(NAME, CC_NUMBER) VALUES ($1,$2) RETURNING id",
		c.Name, c.AccountNumber).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (d *psqlDataStore) GetCustomer(id int64) (*getting.Customer, error) {
	var c getting.Customer
	row := d.db.QueryRow("SELECT ID, NAME, CC_NUMBER FROM customer WHERE id=$1", id)
	err := row.Scan(&c.ID, &c.Name, &c.AccountNumber)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// NewCustomerRepository creates a customer repo from where we can add or retrieve customer info
func NewCustomerRepository(dataSourceName string) (CustomerRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &psqlDataStore{db}, nil
}