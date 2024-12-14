package data

import (
	"sync"
	"time"

	"fetch.trungnng.github.io/internal/validator"
	"github.com/google/uuid"
)

// Receipt represents a purchase receipt record in the database.
type Receipt struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Retailer     string
	PurchaseDate time.Time
	PurchaseTime time.Time
	Items        []*Item
	Total        float32
}

// Item represents a receipt's item record in the database.
type Item struct {
	ID               string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ShortDescription string
	Price            float32
}

func ValidateReceipt(v *validator.Validator, rc *Receipt) {
	v.Check(rc.Retailer != "", "retailer", "must be provided")
	v.Check(len(rc.Retailer) <= 500, "retailer", "must not be more than 500 bytes long")

	v.Check(!rc.PurchaseDate.IsZero(), "purchaseDate", "must be provided")
	v.Check(rc.PurchaseDate.Before(time.Now().UTC()), "purchaseDate", "must not be in the future")

	v.Check(!rc.PurchaseTime.IsZero(), "purchaseTime", "must be provided")

	v.Check(rc.Items != nil, "items", "must be provided")
	v.Check(len(rc.Items) >= 1, "items", "must contain at least 1 item")

	v.Check(!(rc.Total < 0.0), "total", "must not be negative")

	// Validate reciept's items
	for _, item := range rc.Items {
		v.Check(item.ShortDescription != "", "shortDescription", "must be provided")
		v.Check(!(item.Price < 0.0), "price", "must not be negative")
	}
}

// In-memory store for Receipt records.
type ReceiptModel struct {
	data map[string]*Receipt
	mu   sync.RWMutex
}

// NewReceiptModel initializes a new instance of ReceiptModel with an empty data store.
func NewReceiptModel() *ReceiptModel {
	return &ReceiptModel{
		data: make(map[string]*Receipt),
	}
}

// Create a new Receipt, use UUID for reciept's ID.
func (r *ReceiptModel) NewReceipt() *Receipt {
	return &Receipt{
		ID:        uuid.NewString(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Items:     []*Item{},
	}
}

func (r *ReceiptModel) NewReceiptItem() *Item {
	return &Item{
		ID:        uuid.NewString(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
}

// Insert adds a new Receipt to the in-memory data store.
// If a receipt with the same ID already exists, it returns an error.
func (r *ReceiptModel) Insert(receipt *Receipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[receipt.ID]; exists {
		return ErrDuplicateRecord
	}

	r.data[receipt.ID] = receipt
	return nil
}

// Get retrieves a Receipt from the in-memory data store by its ID.
// Returns ErrRecordNotFound if no matching receipt is found.
func (r *ReceiptModel) Get(id string) (*Receipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	receipt, exists := r.data[id]
	if !exists {
		return nil, ErrRecordNotFound
	}

	return receipt, nil
}

// Update modifies an existing Receipt in the in-memory data store.
// If no receipt with the given ID exists, it returns ErrRecordNotFound.
func (r *ReceiptModel) Update(receipt *Receipt) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[receipt.ID]; !exists {
		return ErrRecordNotFound
	}

	r.data[receipt.ID] = receipt
	return nil
}

// Delete removes a Receipt from the in-memory data store by its ID.
// If no receipt with the given ID exists, it returns ErrRecordNotFound.
func (r *ReceiptModel) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.data[id]; !exists {
		return ErrRecordNotFound
	}

	delete(r.data, id)
	return nil
}
