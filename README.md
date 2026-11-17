# go-cart-api

A Cart REST API written in Go with [Gin](https://gin-gonic.com/), ported from a TypeScript + Fastify version.

## Running

```sh
go test ./... -v     # run the tests
go run .             # start the server on :8081
PORT=3000 go run .   # start it on another port
```

`PORT` is read from the environment and defaults to `8081`.

## Manual testing with Bruno

The [`bruno/`](bruno/) folder is a [Bruno](https://www.usebruno.com/) collection in YAML (OpenCollection) format.

1. Start the server (`go run .`)
2. In Bruno: **Open Collection** → select the `bruno/` folder
3. Select the `localhost` environment (sets `baseUrl` to `http://localhost:8081/api`; change it if you use another `PORT`)
4. Run the requests in order: create → get → update → checkout → delete

Each request sets `ecommerceId` and `customerId` (`e1` / `c1`) as request variables.

## Technical decisions

| Topic | Choice | Why |
|---|---|---|
| HTTP framework | **Gin** | Close to Fastify; `binding` tags replace TypeBox validation |
| Storage | **In-memory** (map + `sync.Mutex`) behind a `CartStore` interface | Simple to start with; MongoDB can be added later as another package under `internal/store/` |
| Layout | `main.go` + `internal/{cart,httpapi,store/memory}` | `internal/` keeps the packages private to the module; packages are split by responsibility, not by kind of type |
| Dependencies | `httpapi` and `store/memory` import `cart`; `cart` imports neither | The domain package stays free of HTTP and storage details |
| `date_checkout` | `*time.Time` with `omitempty` | `nil` = not checked out yet; omitted from JSON |
| Price API | `Product.UnitPrice(checkout)` per unit; `Cart.TotalPrice()` multiplies by quantity | |

## Layout

```
go-cart-api/
  go.mod                      module cart-api
  main.go                     wiring + server start
  internal/cart/              domain: no HTTP, no storage
    model.go                  Status, FileType, Product, Cart, CartStore interface, ErrNotFound, ErrAlreadyExists
    pricing.go                pricing policies
    pricing_test.go           table-driven tests
  internal/store/memory/
    memoryStore.go            CartStore implementation (map + mutex)
    memoryStore_test.go       store tests
  internal/httpapi/
    handlers.go               Gin handlers, request and response structs
    handlers_test.go          end-to-end tests with httptest (router + handlers + memory store)
  bruno/                      Bruno collection for manual testing
    environments/localhost.yml
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
