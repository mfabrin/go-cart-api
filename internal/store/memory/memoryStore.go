package memory

import (
	"sync"
	"time"

	"cart-api/internal/cart"
)

type cartKey struct {
	ecommerceID, customerID string
}

type Store struct {
	mu    sync.Mutex
	carts map[cartKey]cart.Cart
}

func New() *Store {
	return &Store{carts: make(map[cartKey]cart.Cart)}
}

var _ cart.CartStore = (*Store)(nil)

func (ms *Store) Get(ecommerceID, customerID string) (cart.Cart, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := cartKey{ecommerceID, customerID}
	crt, ok := ms.carts[key]

	if !ok {
		return cart.Cart{}, cart.ErrNotFound
	}

	return crt, nil
}

func (ms *Store) Create(ecommerceID, customerID string, status cart.Status) (cart.Cart, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := cartKey{ecommerceID, customerID}
	_, ok := ms.carts[key]

	if ok {
		return cart.Cart{}, cart.ErrAlreadyExists
	}

	timeNow := time.Now()

	crt := cart.Cart{
		EcommerceID: ecommerceID,
		CustomerID:  customerID,
		Status:      status,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}

	ms.carts[key] = crt

	return crt, nil
}

func (ms *Store) Update(newCart cart.Cart) (cart.Cart, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := cartKey{newCart.EcommerceID, newCart.CustomerID}

	crt, ok := ms.carts[key]

	if !ok {
		return cart.Cart{}, cart.ErrNotFound
	}

	crt.DateCheckout = newCart.DateCheckout
	crt.ItemList = newCart.ItemList
	crt.Status = newCart.Status
	crt.UpdatedAt = time.Now()

	ms.carts[key] = crt

	return crt, nil
}

func (ms *Store) Delete(ecommerceID, customerID string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := cartKey{ecommerceID, customerID}

	_, ok := ms.carts[key]

	if !ok {
		return cart.ErrNotFound
	}

	delete(ms.carts, key)

	return nil
}
