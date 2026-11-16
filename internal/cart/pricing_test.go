package cart

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestProductPrice_NoCheckout(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		expected float64
	}{
		{
			name:     "PDF no discount",
			product:  Product{Quantity: 1, FileType: FileTypePDF},
			expected: 1.15,
		},
		{
			name:     "PSD with 5 discount and 35 increment",
			product:  Product{Quantity: 200, FileType: FileTypePSD},
			expected: 1.2825,
		},
		{
			name:     "AI with 20 discount and 25 increment",
			product:  Product{Quantity: 1001, FileType: FileTypeAI},
			expected: 1,
		},
		{
			name:     "PNG file type with 5 discount and 0 increment",
			product:  Product{Quantity: 1, FileType: "PNG"},
			expected: 1,
		},
		{
			name:     "AI with 5 discount and 10 increment",
			product:  Product{Quantity: 270, FileType: FileTypeAI},
			expected: 1.125,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.product.UnitPrice(nil)
			assertFloat(t, got, tt.expected)
		})
	}
}

func TestProductPrice_WithCheckout(t *testing.T) {
	checkoutDate := time.Date(2026, 11, 14, 5, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		product  Product
		expected float64
	}{
		{
			name: "PDF: no discount, 30%% checkout increment",
			product: Product{
				Quantity:     1,
				FileType:     FileTypePDF,
				DeliveryDate: checkoutDate.Add(20 * time.Hour),
			},
			expected: 1.495,
		},
		{
			name: "PSD: 5 discount, no increment for checkout",
			product: Product{
				Quantity:     200,
				FileType:     FileTypePSD,
				DeliveryDate: checkoutDate.Add(80 * time.Hour),
			},
			expected: 1.2825,
		},
		{
			name: "AI: 20 discount, 10%% increment by checkout",
			product: Product{
				Quantity:     1001,
				FileType:     FileTypeAI,
				DeliveryDate: checkoutDate.Add(60 * time.Hour),
			},
			expected: 1.1,
		},
		{
			name: "PNG: no discount, no increments",
			product: Product{
				Quantity:     1,
				FileType:     "PNG",
				DeliveryDate: checkoutDate.Add(100 * time.Hour),
			},
			expected: 1,
		},
		{
			name: "PNG: no discount, no increments",
			product: Product{
				Quantity:     1,
				FileType:     "PNG",
				DeliveryDate: checkoutDate.Add(-10 * time.Hour),
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.product.UnitPrice(&checkoutDate)
			assertFloat(t, got, tt.expected)
		})
	}
}

func TestQuantityDiscount(t *testing.T) {

	testCases := []struct {
		testName string
		quantity int
		want     float64
	}{
		{quantity: 0, want: 0},
		{quantity: 1, want: 0},
		{quantity: 100, want: 0},
		{quantity: 101, want: 5},
		{quantity: 250, want: 5},
		{quantity: 251, want: 10},
		{quantity: 500, want: 10},
		{quantity: 501, want: 15},
		{quantity: 1000, want: 15},
		{quantity: 1001, want: 20},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("get %v discount with %d quantity", tc.want, tc.quantity)
		t.Run(testName, func(t *testing.T) {
			pct := quantityDiscount(tc.quantity)
			assertFloat(t, pct, tc.want)
		})

	}
}

func TestDeliveryIncrementPct(t *testing.T) {
	checkoutToday := time.Date(2026, 11, 14, 1, 0, 0, 0, time.UTC)
	checkoutFiveDaysAgo := time.Date(2026, 11, 9, 1, 0, 0, 0, time.UTC)

	checkoutDates := []time.Time{checkoutToday, checkoutFiveDaysAgo}
	testCases := []struct {
		offset  time.Duration
		want    float64
		wantErr bool
	}{
		{offset: 0 * time.Hour, want: 30},
		{offset: 24 * time.Hour, want: 30},
		{offset: 24*time.Hour + time.Nanosecond, want: 20},
		{offset: 48 * time.Hour, want: 20},
		{offset: 48*time.Hour + time.Nanosecond, want: 10},
		{offset: 72 * time.Hour, want: 10},
		{offset: 72*time.Hour + time.Nanosecond, want: 0},
		{offset: 1000 * time.Hour, want: 0},
		{wantErr: true, offset: -80 * time.Hour, want: 0},
		{wantErr: true, offset: -1 * time.Nanosecond, want: 0},
	}

	for _, cd := range checkoutDates {
		for _, tc := range testCases {
			deliveryDate := cd.Add(tc.offset)

			testName := fmt.Sprintf("%s: %s -> %v", cd.String(), tc.offset.String(), tc.want)

			t.Run(testName, func(t *testing.T) {
				res, err := deliveryIncrementPct(deliveryDate, cd)

				gotErr := err != nil

				if gotErr != tc.wantErr {
					t.Errorf("error = %v, want error: %v", err, tc.wantErr)
				}

				assertFloat(t, res, tc.want)
			})
		}
	}
}

func TestTotalPriceCalculation(t *testing.T) {
	checkoutDate := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		cart     Cart
		expected float64
	}{
		{
			name:     "Cart with no items",
			cart:     Cart{EcommerceID: "123"},
			expected: 0,
		},
		{
			name: "Cart with 1 SKU: no checkout",
			cart: Cart{
				ItemList: []Product{
					{FileType: FileTypeAI, Quantity: 200},
				}},
			expected: 237.5,
		},
		{
			name: "Cart with 2 SKU, with checkout",
			cart: Cart{
				DateCheckout: &checkoutDate,
				ItemList: []Product{
					{FileType: FileTypeAI, Quantity: 200, DeliveryDate: checkoutDate.Add(20 * time.Hour)},
					{FileType: FileTypePSD, Quantity: 600, DeliveryDate: checkoutDate.Add(60 * time.Hour)},
				}},
			expected: 1066.1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := test.cart.TotalPrice()
			assertFloat(t, res, test.expected)
		})
	}
}

func assertFloat(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("got %v, want %v", got, want)
	}
}
