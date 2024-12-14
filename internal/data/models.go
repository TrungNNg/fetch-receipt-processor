package data

import (
	"errors"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.
var (
	ErrRecordNotFound  = errors.New("record not found")
	ErrDuplicateRecord = errors.New("record with same ID exist")
)

// Models acts as a container for different database models.
type Models struct {
	Receipts *ReceiptModel
}

// NewModels initializes and returns an instance of Models.
func NewModels() *Models {
	return &Models{
		Receipts: NewReceiptModel(),
	}
}
