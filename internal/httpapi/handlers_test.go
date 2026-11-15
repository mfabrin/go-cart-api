package httpapi

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cart-api/internal/store/memory"

	"github.com/gin-gonic/gin"
)

const (
	testEcommerceID = "shop-1"
	testCustomerID  = "customer-1"
	cartPath        = "/api/v1/cart/" + testEcommerceID + "/" + testCustomerID
)

// newTestRouter mirrors the route table in main.go, backed by a fresh in-memory store.
func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := New(memory.New())
	router := gin.New()
	api := router.Group("/api/v1")

	api.GET("/cart/:ecommerce_id/:customer_id", h.GetCart)
	api.POST("/cart/create", h.CreateCart)
	api.PUT("/cart/update", h.UpdateCart)
	api.PUT("/cart/checkout", h.CheckoutCart)
	api.DELETE("/cart/:ecommerce_id/:customer_id", h.DeleteCart)

	return router
}

func do(t *testing.T, router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer

	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()

	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var out map[string]any

	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decoding response %q: %v", rec.Body.String(), err)
	}

	return out
}

func assertFloat(t *testing.T, name string, got any, want float64) {
	t.Helper()

	f, ok := got.(float64)
	if !ok {
		t.Fatalf("%s = %v (%T), want a number", name, got, got)
	}

	if math.Abs(f-want) > 1e-9 {
		t.Errorf("%s = %v, want %v", name, f, want)
	}
}

func createBody() map[string]any {
	return map[string]any{
		"ecommerce_id": testEcommerceID,
		"customer_id":  testCustomerID,
		"status":       "created",
	}
}

func updateBody() map[string]any {
	return map[string]any{
		"ecommerce_id": testEcommerceID,
		"customer_id":  testCustomerID,
		"item_list": []map[string]any{{
			"product_sku":   "SKU-1",
			"product_name":  "Business card",
			"file_type":     "PDF",
			"quantity":      10,
			"delivery_date": time.Now().Add(12 * time.Hour).UTC().Format(time.RFC3339),
		}},
	}
}

// TestCartLifecycle walks the whole happy path through the real router,
// handlers and in-memory store, and pins the observable behavior.
func TestCartLifecycle(t *testing.T) {
	router := newTestRouter()

	// create
	rec := do(t, router, http.MethodPost, "/api/v1/cart/create", createBody())
	assertStatus(t, rec, http.StatusCreated)

	created := decode(t, rec)
	if created["status"] != "created" {
		t.Errorf("created status = %v, want created", created["status"])
	}

	// update: adds items, status becomes building
	rec = do(t, router, http.MethodPut, "/api/v1/cart/update", updateBody())
	assertStatus(t, rec, http.StatusOK)

	if got := decode(t, rec)["status"]; got != "building" {
		t.Errorf("status after update = %v, want building", got)
	}

	// get before checkout: no delivery increment
	// unit = 1 * 1.15 (PDF) = 1.15, price = 11.5
	rec = do(t, router, http.MethodGet, cartPath, nil)
	assertStatus(t, rec, http.StatusOK)

	got := decode(t, rec)
	assertFloat(t, "total_price (before checkout)", got["total_price"], 11.5)

	if _, ok := got["date_checkout"]; ok {
		t.Errorf("date_checkout should be omitted before checkout")
	}

	// checkout
	rec = do(t, router, http.MethodPut, "/api/v1/cart/checkout", map[string]any{
		"ecommerce_id": testEcommerceID,
		"customer_id":  testCustomerID,
	})
	assertStatus(t, rec, http.StatusOK)

	if got := decode(t, rec)["status"]; got != "checkout" {
		t.Errorf("status after checkout = %v, want checkout", got)
	}

	// get after checkout: delivery within 24h => +30%, then PDF +15%
	// unit = 1 * 1.3 * 1.15 = 1.495, price = 14.95
	rec = do(t, router, http.MethodGet, cartPath, nil)
	assertStatus(t, rec, http.StatusOK)

	got = decode(t, rec)
	assertFloat(t, "total_price (after checkout)", got["total_price"], 14.95)

	items, ok := got["item_list"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("item_list = %v, want 1 item", got["item_list"])
	}

	item := items[0].(map[string]any)
	assertFloat(t, "unit_price", item["unit_price"], 1.495)
	assertFloat(t, "price", item["price"], 14.95)

	if _, ok := got["date_checkout"]; !ok {
		t.Errorf("date_checkout should be present after checkout")
	}

	// delete, then it's gone
	rec = do(t, router, http.MethodDelete, cartPath, nil)
	assertStatus(t, rec, http.StatusNoContent)

	rec = do(t, router, http.MethodGet, cartPath, nil)
	assertStatus(t, rec, http.StatusNotFound)
}

func TestCreateCart_Duplicate(t *testing.T) {
	router := newTestRouter()

	assertStatus(t, do(t, router, http.MethodPost, "/api/v1/cart/create", createBody()), http.StatusCreated)
	assertStatus(t, do(t, router, http.MethodPost, "/api/v1/cart/create", createBody()), http.StatusBadRequest)
}

func TestNotFound(t *testing.T) {
	checkoutBody := map[string]any{"ecommerce_id": testEcommerceID, "customer_id": testCustomerID}

	tests := []struct {
		name   string
		method string
		path   string
		body   any
	}{
		{"get", http.MethodGet, cartPath, nil},
		{"update", http.MethodPut, "/api/v1/cart/update", updateBody()},
		{"checkout", http.MethodPut, "/api/v1/cart/checkout", checkoutBody},
		{"delete", http.MethodDelete, cartPath, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, newTestRouter(), tt.method, tt.path, tt.body)
			assertStatus(t, rec, http.StatusNotFound)
		})
	}
}

func TestValidation_BadRequest(t *testing.T) {
	badUpdate := func(mutate func(item map[string]any)) map[string]any {
		body := updateBody()
		mutate(body["item_list"].([]map[string]any)[0])

		return body
	}

	tests := []struct {
		name string
		path string
		body any
	}{
		{"create: missing customer_id", "/api/v1/cart/create", map[string]any{"ecommerce_id": "x", "status": "created"}},
		{"create: invalid status", "/api/v1/cart/create", map[string]any{"ecommerce_id": "x", "customer_id": "y", "status": "nope"}},
		{"update: empty item_list", "/api/v1/cart/update", map[string]any{"ecommerce_id": "x", "customer_id": "y", "item_list": []any{}}},
		{"update: invalid file_type", "/api/v1/cart/update", badUpdate(func(i map[string]any) { i["file_type"] = "GIF" })},
		{"update: zero quantity", "/api/v1/cart/update", badUpdate(func(i map[string]any) { i["quantity"] = 0 })},
		{"update: negative quantity", "/api/v1/cart/update", badUpdate(func(i map[string]any) { i["quantity"] = -5 })},
		{"checkout: missing ids", "/api/v1/cart/checkout", map[string]any{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodPut
			if tt.path == "/api/v1/cart/create" {
				method = http.MethodPost
			}

			rec := do(t, newTestRouter(), method, tt.path, tt.body)
			assertStatus(t, rec, http.StatusBadRequest)
		})
	}
}

func TestMalformedJSON(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/create", bytes.NewBufferString("{not json"))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assertStatus(t, rec, http.StatusBadRequest)
}
