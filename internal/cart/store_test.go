package cart

import (
	"errors"
	"math"
	"testing"
)

func TestCreateTwice_SuccessThenFailure(t *testing.T) {
	cartStore := NewMemoryStore()

	const ecommerceTestID = "test-ecommerce-id"
	const customerTestID = "test-customer-id"
	const testStatus = StatusCreated

	testCases := []struct {
		testName    string
		ecommerceID string
		customerID  string
		status      Status
		wantErr     error
	}{
		{
			testName:    "Create a cart success",
			ecommerceID: ecommerceTestID,
			customerID:  customerTestID,
			status:      testStatus,
			wantErr:     nil,
		},
		{
			testName:    "Create a cart failure",
			ecommerceID: ecommerceTestID,
			customerID:  customerTestID,
			status:      testStatus,
			wantErr:     ErrAlreadyExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			cart, err := cartStore.Create(tc.ecommerceID, tc.customerID, tc.status)
			assertError(t, err, tc.wantErr)

			if tc.wantErr == nil {
				assertStrings(t, "ecommerceID", cart.EcommerceID, tc.ecommerceID)
				assertStrings(t, "customerID", cart.CustomerID, tc.customerID)
			}
		})
	}
}

func TestGet_SuccessAndFailure(t *testing.T) {
	cartStore := NewMemoryStore()

	testEcommerceId := "ecommerceID"
	testCustomerID := "customerID"
	testStatus := StatusCreated

	cartStore.Create(testEcommerceId, testCustomerID, testStatus)

	testCases := []struct {
		testName    string
		ecommerceID string
		customerID  string
		wantErr     error
	}{
		{
			testName:    "Get an existing cart",
			ecommerceID: "ecommerceID",
			customerID:  "customerID",
			wantErr:     nil,
		},
		{
			testName:    "Get a non existing cart",
			ecommerceID: "not-exists-ecommerceID",
			customerID:  "not-exists-customerID",
			wantErr:     ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			cart, err := cartStore.Get(tc.ecommerceID, tc.customerID)

			assertError(t, err, tc.wantErr)

			if tc.wantErr == nil {
				assertStrings(t, "ecommerceID", cart.EcommerceID, tc.ecommerceID)
				assertStrings(t, "customerID", cart.CustomerID, tc.customerID)
			}
		})
	}
}

func TestUpdate_SuccessAndFailure(t *testing.T) {
	cartStore := NewMemoryStore()

	const testEcommerceID = "ecommerceID"
	const testCustomerID = "customerID"
	const testStatus = StatusCreated

	cart, _ := cartStore.Create(testEcommerceID, testCustomerID, testStatus)

	testCases := []struct {
		testName    string
		ecommerceID string
		customerID  string
		wantErr     error
	}{
		{
			testName:    "Update existing cart",
			ecommerceID: cart.EcommerceID,
			customerID:  cart.CustomerID,
			wantErr:     nil,
		},
		{
			testName:    "Update not existing cart",
			ecommerceID: "not-existing-ecommerceID",
			customerID:  "not-existing-customerID",
			wantErr:     ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.wantErr != nil {
				dummyCart := Cart{EcommerceID: tc.ecommerceID, CustomerID: tc.customerID}
				_, err := cartStore.Update(dummyCart)
				assertError(t, err, tc.wantErr)
			} else {
				product := Product{ProductSKU: "productSKU", ProductName: "productName", FileType: FileTypeAI, Quantity: 250}
				cart.ItemList = append(cart.ItemList, product)

				updatedCart, err := cartStore.Update(cart)

				assertError(t, err, tc.wantErr)
				assertStrings(t, "ecommerceID", updatedCart.EcommerceID, tc.ecommerceID)
				assertStrings(t, "customerID", updatedCart.CustomerID, tc.customerID)

				if updatedCart.CreatedAt.Equal(cart.UpdatedAt) {
					t.Errorf("updatedAt should be updated, got %v", updatedCart.UpdatedAt)
				}

				assertInt(t, "array length", len(updatedCart.ItemList), 1)
				assertStrings(t, "productSKU", updatedCart.ItemList[0].ProductSKU, product.ProductSKU)
				assertStrings(t, "productName", updatedCart.ItemList[0].ProductName, product.ProductName)
				assertInt(t, "quantity", updatedCart.ItemList[0].Quantity, product.Quantity)
				assertStrings(t, "productSKU", updatedCart.ItemList[0].ProductSKU, product.ProductSKU)
			}
		})
	}
}

func TestDelete_SuccessAndFailure(t *testing.T) {
	// Success: Delete returns nil, then Get returns ErrNotFound
	cartStore := NewMemoryStore()

	const testEcommerceID = "ecommerceID"
	const testCustomerID = "customerID"
	const testStatus = StatusCreated

	cart, _ := cartStore.Create(testEcommerceID, testCustomerID, testStatus)

	testCases := []struct {
		testName    string
		ecommerceID string
		customerID  string
		want        error
	}{
		{
			testName:    "Delete an existing cart",
			ecommerceID: cart.EcommerceID,
			customerID:  cart.CustomerID,
			want:        nil,
		},
		{
			testName:    "Delete a not existing cart",
			ecommerceID: "not-existing-ecommerceID",
			customerID:  "not-existing-customerID",
			want:        ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			err := cartStore.Delete(tc.ecommerceID, tc.customerID)
			assertError(t, err, tc.want)
		})
	}
}

func assertStrings(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("unexpected %s received: got %s, want %s", name, got, want)
	}
}

func assertError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("unexpected error: got %v, want %v", got, want)
	}
}

func assertInt(t *testing.T, name string, got, want int) {
	t.Helper()
	if math.Abs(float64(got-want)) > 1e-9 {
		t.Errorf("unexpected %s received: got %d, want %d", name, got, want)
	}
}
