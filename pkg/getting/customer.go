package getting

// AccountNumber represents an account number
type AccountNumber string

// Customer represents customer data retrieved from a data store
type Customer struct {
	ID            int64
	Name          string
	AccountNumber AccountNumber
}