# go-cart-api

A Cart REST API written in Go with [Gin](https://gin-gonic.com/), ported from a TypeScript + Fastify version.

## Running

```sh
go test ./... -v     # run the tests
go run .             # start the server on :8081
PORT=3000 go run .   # start it on another port
```

`PORT` is read from the environment and defaults to `8081`.

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
  main.go                     wiring + server start
  internal/cart/
    model.go                  Status, FileType, Product, Cart
    pricing.go                pricing policies
    pricing_test.go           table-driven tests
    store.go                  CartStore interface + in-memory impl (map + mutex)
    store_test.go             store tests
    handlers.go               Gin handlers, request and response structs
```

## API

Endpoints (prefix `/api/v1`):

| Method | Path | Notes |
|---|---|---|
| POST | `/cart/create` | body: `ecommerce_id`, `customer_id`, `status` (`created`/`building`/`checkout`) → 201 with the cart; 400 if the body is invalid or the cart already exists. Timestamps are set by the server |
| PUT | `/cart/update` | body: `ecommerce_id`, `customer_id`, `item_list` (not empty) → 200; the server sets status `building` and clears `date_checkout`; 400 if invalid, 404 if the cart doesn't exist |
| GET | `/cart/:ecommerce_id/:customer_id` | 200 with the cart, `total_price`, and per item `unit_price` (price of one unit) and `price` (`unit_price` × `quantity`); 404 if not found |
| PUT | `/cart/checkout` | body: `ecommerce_id`, `customer_id` → 200; status `checkout`, `date_checkout = now`; 400 if invalid, 404 if not found |
| DELETE | `/cart/:ecommerce_id/:customer_id` | 204 with no body; 404 if not found |

Item fields (all required): `product_sku`, `product_name`, `file_type` (`PDF`/`PSD`/`AI`), `quantity` (> 0), `delivery_date` (RFC 3339).

Errors are returned as `{"error": "..."}`; unexpected store errors give 500.

## Pricing

Base price 1.00 € per item, then:

- Quantity discount: >100 → 5%, >250 → 10%, >500 → 15%, >1000 → 20%
- Delivery vs checkout date: within 24h → +30%, 48h → +20%, 72h → +10%, otherwise +0%
  (only applied once the cart has a `date_checkout`)
- File type: PDF +15%, AI +25%, PSD +35%
