package memory

import (
	"errors"
	"testing"

	"cart-api/internal/cart"
)

const ecommerceTestID = "test-ecommerce-id"
const customerTestID = "test-customer-id"

func TestCreateTwice_SuccessAndFailure(t *testing.T) {
	t.Run("Create a cart", func(t *testing.T) {
		ms := New()

		c, firstErr := ms.Create(ecommerceTestID, customerTestID, cart.StatusCreated)

		assertFatalError(t, firstErr, nil)

		if !c.CreatedAt.Equal(c.UpdatedAt) {
			t.Errorf("createdAt and updatedAt should be equal after creation")
		}

		assertStrings(t, "ecommerceID", c.EcommerceID, ecommerceTestID)
		assertStrings(t, "customerID", c.CustomerID, customerTestID)
		assertStatus(t, "status", c.Status, cart.StatusCreated)

		_, err := ms.Create(ecommerceTestID, customerTestID, cart.StatusCreated)

		assertError(t, err, cart.ErrAlreadyExists)
	})
}

func TestGet_SuccessAndFailure(t *testing.T) {
	ms := New()

	testEcommerceID := "ecommerceID"
	testCustomerID := "customerID"
	testStatus := cart.StatusCreated

	_, createErr := ms.Create(testEcommerceID, testCustomerID, testStatus)

	assertFatalError(t, createErr, nil)

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
			wantErr:     cart.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			cart, err := ms.Get(tc.ecommerceID, tc.customerID)

			assertError(t, err, tc.wantErr)

			if tc.wantErr == nil {
				assertStrings(t, "ecommerceID", cart.EcommerceID, tc.ecommerceID)
				assertStrings(t, "customerID", cart.CustomerID, tc.customerID)
			}
		})
	}
}

func TestUpdate_SuccessAndFailure(t *testing.T) {
	ms := New()

	const testEcommerceID = "ecommerceID"
	const testCustomerID = "customerID"
	const testStatus = cart.StatusCreated

	c, createErr := ms.Create(testEcommerceID, testCustomerID, testStatus)
	assertFatalError(t, createErr, nil)

	testCases := []struct {
		testName    string
		ecommerceID string
		customerID  string
		wantErr     error
	}{
		{
			testName:    "Update existing cart",
			ecommerceID: c.EcommerceID,
			customerID:  c.CustomerID,
			wantErr:     nil,
		},
		{
			testName:    "Update not existing cart",
			ecommerceID: "not-existing-ecommerceID",
			customerID:  "not-existing-customerID",
			wantErr:     cart.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			if tc.wantErr != nil {
				dummyCart := cart.Cart{EcommerceID: tc.ecommerceID, CustomerID: tc.customerID}
				_, err := ms.Update(dummyCart)
				assertError(t, err, tc.wantErr)
			} else {
				product := cart.Product{ProductSKU: "productSKU", ProductName: "productName", FileType: cart.FileTypeAI, Quantity: 250}
				c.ItemList = append(c.ItemList, product)

				_, err := ms.Update(c)

				assertError(t, err, tc.wantErr)

				updatedCart, getErr := ms.Get(c.EcommerceID, c.CustomerID)

				assertError(t, getErr, nil)

				assertStrings(t, "ecommerceID", updatedCart.EcommerceID, tc.ecommerceID)
				assertStrings(t, "customerID", updatedCart.CustomerID, tc.customerID)

				if !updatedCart.UpdatedAt.After(c.UpdatedAt) {
					t.Errorf("updatedAt after the update should be later")
				}

				assertInt(t, "array length", len(updatedCart.ItemList), 1)
				assertStrings(t, "productSKU", updatedCart.ItemList[0].ProductSKU, product.ProductSKU)
				assertStrings(t, "productName", updatedCart.ItemList[0].ProductName, product.ProductName)
				assertInt(t, "quantity", updatedCart.ItemList[0].Quantity, product.Quantity)
				assertFileType(t, "fileType", updatedCart.ItemList[0].FileType, product.FileType)
			}
		})
	}
}

func TestDelete_SuccessAndFailure(t *testing.T) {
	ms := New()

	const testEcommerceID = "ecommerceID"
	const testCustomerID = "customerID"
	const testStatus = cart.StatusCreated

	c, createErr := ms.Create(testEcommerceID, testCustomerID, testStatus)
	assertFatalError(t, createErr, nil)

	testCases := []struct {
		testName    string
		ecommerceID string
		customerID  string
		wantErr     error
	}{
		{
			testName:    "Delete an existing cart",
			ecommerceID: c.EcommerceID,
			customerID:  c.CustomerID,
			wantErr:     nil,
		},
		{
			testName:    "Delete a not existing cart",
			ecommerceID: "not-existing-ecommerceID",
			customerID:  "not-existing-customerID",
			wantErr:     cart.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			err := ms.Delete(tc.ecommerceID, tc.customerID)
			assertError(t, err, tc.wantErr)

			_, getErr := ms.Get(tc.ecommerceID, tc.customerID)
			assertError(t, getErr, cart.ErrNotFound)
		})
	}
}

func assertStrings(t *testing.T, name, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("unexpected %s received: got %s, want %s", name, got, want)
	}
}

func assertStatus(t *testing.T, name, got, want cart.Status) {
	t.Helper()
	if got != want {
		t.Errorf("unexpected %s received: got %s, want %s", name, got, want)
	}
}

func assertFileType(t *testing.T, name, got, want cart.FileType) {
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

func assertFatalError(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("unexpected error: got %v, want %v", got, want)
	}
}

func assertInt(t *testing.T, name string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("unexpected %s received: got %d, want %d", name, got, want)
	}
}
