# go-cart-api

A Cart REST API written in Go with [Gin](https://gin-gonic.com/), ported from a TypeScript + Fastify version.

## Running

```sh
go test ./... -v     # run the tests
go run .             # start the server (once main.go exists)
```

## Technical decisions

| Topic | Choice | Why |
|---|---|---|
| HTTP framework | **Gin** | Close to Fastify; `binding` tags replace TypeBox validation |
| Storage | **In-memory** (map + `sync.Mutex`) behind a `CartStore` interface | Simple to start with; MongoDB can be added later |
| Layout | `main.go` + `internal/cart/` | `internal/` keeps the package private to the module |
| `date_checkout` | `*time.Time` with `omitempty` | `nil` = not checked out yet; omitted from JSON |
| Price API | `Product.UnitPrice(checkout)` per unit; `Cart.TotalPrice()` multiplies by quantity | |

## Layout

```
go-cart-api/
  go.mod                      module cart-api
  main.go                     (planned) wiring + server start
  internal/cart/
    model.go                  Status, FileType, Product, Cart
    pricing.go                pricing policies
    pricing_test.go           table-driven tests
    store.go                  (planned) CartStore interface + in-memory impl
    handlers.go               (planned) Gin handlers
```

## API

Endpoints (prefix `/api/v1`):

| Method | Path | Notes |
|---|---|---|
| POST | `/cart/create` | body: `ecommerce_id`, `customer_id`, `status`, `created_at`, `updated_at` → 201; 400 if it already exists |
| PUT | `/cart/update` | body: `ecommerce_id`, `customer_id`, `item_list` → status `building`, clears `date_checkout` |
| GET | `/cart/:ecommerce_id/:customer_id` | returns cart + `total_price` + per-item `calculated_price` |
| PUT | `/cart/checkout` | body: `ecommerce_id`, `customer_id` → status `checkout`, sets `date_checkout = now` |
| DELETE | `/cart/delete/:ecommerce_id/:customer_id` | 204 |

Item fields: `product_sku`, `product_name`, `file_type` (`PDF`/`PSD`/`AI`), `quantity`, `delivery_date`.

## Pricing

Base price 1.00 € per item, then:

- Quantity discount: >100 → 5%, >250 → 10%, >500 → 15%, >1000 → 20%
- Delivery vs checkout date: within 24h → +30%, 48h → +20%, 72h → +10%, otherwise +0%
  (only applied once the cart has a `date_checkout`)
- File type: PDF +15%, AI +25%, PSD +35%
