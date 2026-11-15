package cart

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("cart not found")
	ErrAlreadyExists = errors.New("cart already exists")
)

type CartStore interface {
	Create(ecommerceID, customerID string, status Status) (Cart, error)
	Get(ecommerceID, customerID string) (Cart, error)
	Update(cart Cart) (Cart, error)
	Delete(ecommerceID, customerID string) error
}

type cartKey struct {
	ecommerceID, customerID string
}

type MemoryStore struct {
	carts map[cartKey]Cart
}

var _ CartStore = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{carts: make(map[cartKey]Cart)}
}

func (ms *MemoryStore) Get(ecommerceID, customerID string) (Cart, error) {
	key := cartKey{ecommerceID, customerID}
	cart, ok := ms.carts[key]

	if !ok {
		return Cart{}, ErrNotFound
	}

	return cart, nil
}

func (ms *MemoryStore) Create(ecommerceID, customerID string, status Status) (Cart, error) {
	key := cartKey{ecommerceID, customerID}
	_, ok := ms.carts[key]

	if ok {
		return Cart{}, ErrAlreadyExists
	}

	timeNow := time.Now()

	cart := Cart{
		EcommerceID: ecommerceID,
		CustomerID:  customerID,
		Status:      status,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	ms.carts[key] = cart

	return cart, nil
}

func (ms *MemoryStore) Update(newCart Cart) (Cart, error) {
	key := cartKey{newCart.EcommerceID, newCart.CustomerID}

	cart, ok := ms.carts[key]

	if !ok {
		return Cart{}, ErrNotFound
	}

	cart.DateCheckout = newCart.DateCheckout
	cart.ItemList = newCart.ItemList
	cart.Status = newCart.Status
	cart.UpdatedAt = time.Now()

	ms.carts[key] = cart

	return cart, nil
}

func (ms *MemoryStore) Delete(ecommerceID, customerID string) error {
	key := cartKey{ecommerceID, customerID}

	_, ok := ms.carts[key]

	if !ok {
		return ErrNotFound
	}

	delete(ms.carts, key)

	return nil
}
