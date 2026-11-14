# AGENTS.md

Instructions for AI assistants working in this repository. For what the code does, see [README.md](README.md).

## Context

This is a **learning project**: a Go port of a Cart REST API originally written in TypeScript + Fastify
(`cart-api`). The goal is to learn Go before a coding assessment with two phases:

1. coding with an AI assistant
2. coding **without** AI (manual)

So the code here must be **understood**, not just generated.

## How we work

- The user writes the code. The assistant **explains concepts, gives small examples and reviews**:
  it does not write whole files or say "done".
- Each step: concept explanation (mapped to TypeScript) → exercise → user writes it → review.
- Reviews point out bugs and give hints rather than corrected code.
- Larger pieces may be written by the assistant only when they're pure boilerplate.
- After each review, update the `[X]` marks in the checklist below.

## Review commands

```sh
go build ./...
gofmt -l .
go vet ./...
go test ./... -v -cover
```

## Checklist

### Setup
- [X] `go.mod` (module `cart-api`)
- [X] VS Code Go extension + gopls working
- [X] Remove `hello.go` placeholder

### Step 1: Domain types (`model.go`)
- [X] `Status` and `FileType` typed constants
- [X] `Product` and `Cart` structs with JSON tags
- [X] Idiomatic naming (`ID`, `SKU`, `CreatedAt`)

### Step 2: Pricing (`pricing.go` + tests)
- [X] `quantityDiscount`, `deliveryIncrementPct`, `fileTypeIncrement`
- [X] `Product.UnitPrice`, `Cart.TotalPrice`
- [X] Table-driven tests: no checkout, with checkout, delivery increment
- [ ] Decide + comment + test behavior for delivery dates **before** checkout (negative diff)
- [ ] `TestCartTotalPrice` (cart with 2+ items, with and without checkout)
- [~] Boundary tests: skipped (note: Go uses `< 24h`, TS uses `<= 24h`)
- [~] Tidy test names: skipped

### Step 3: Storage (`store.go`)
- [ ] `CartStore` interface (create, get, update, delete)
- [ ] In-memory implementation (`map` + `sync.Mutex`)
- [ ] Sentinel errors (`ErrNotFound`, `ErrAlreadyExists`) and `errors.Is`
- [ ] Tests

### Step 4: HTTP handlers with Gin (`handlers.go`)
- [ ] `go get github.com/gin-gonic/gin`
- [ ] Request structs with `binding` tags
- [ ] Response structs (`total_price`, per-item `calculated_price`)
- [ ] 5 handlers (create, update, get, checkout, delete)
- [ ] Error responses and status codes (400 / 404 / 500)

### Step 5: Wiring (`main.go`)
- [ ] Router group `/api/v1`, inject the store into the handlers
- [ ] `PORT` from env with default
- [ ] Manual test with `curl`

### Step 6: Optional
- [ ] Handler tests with `httptest`
- [ ] Validation: `file_type` in `PDF|PSD|AI`, `quantity > 0`
- [ ] Money rounding (2 decimals / integer cents)
- [ ] MongoDB implementation of `CartStore`

### Assessment prep
- [ ] Rewrite `pricing.go` from scratch without AI
- [ ] Rewrite one Gin handler + route from scratch without AI
- [ ] Be able to explain: pointers, interfaces, errors, exported names, zero values

## Go concepts covered so far

- Exported (Uppercase) vs unexported (lowercase): visibility is per **package**
- Naming: MixedCaps, no snake_case; initialisms all caps (`ID`, `SKU`, `HTTP`)
- Struct tags: `json:"name,omitempty"`; `omitempty` doesn't work on struct values like `time.Time`
- Typed constants as enums (`type Status string` + `const`)
- `:=` declares, `=` assigns; the literal decides the type (`1` is int, `1.0` is float64)
- No implicit numeric conversion: `float64(qty)`
- No truthiness: `if p != nil`
- Pointers: `*T` type, `&x` address, `*p` dereference; used for "may be missing" and "modify the original"
- Methods with value receivers: `func (p Product) UnitPrice(...)`
- Tagless `switch`, no fallthrough
- Maps and the comma-ok idiom: `v, ok := m[k]`
- `time.Time`, `time.Duration` (nanoseconds!), `t.Sub`, `t.Add`, `d.Hours()`
- Tests: `_test.go`, `testing.T`, table-driven tests, `t.Run`, float comparison with tolerance
- Watch a test fail before trusting it
