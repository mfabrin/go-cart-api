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
- [X] Decide + comment + test behavior for delivery dates **before** checkout (no delivery increment)
- [X] `TestCartTotalPrice` (cart with 2+ items, with and without checkout)
- [X] Boundary tests (`TestQuantityDiscount` edges at 100/250/500/1000, delivery edges at 0/24h/48h/72h ± 1ns, negative offset)
- [X] Tidy test names (no `%%`, distinct row names, `TestCartTotalPrice`)

### Step 3: Storage (`store.go`)
- [X] `CartStore` interface (create, get, update, delete)
- [X] In-memory implementation (`map` + `sync.Mutex`, `defer Unlock`)
- [X] Sentinel errors (`ErrNotFound`, `ErrAlreadyExists`) and `errors.Is`
- [~] Tests: written, `switch i` removed, Update/Delete asserts added; pending (later): Update `Status`/`DateCheckout`/returned value asserts, split Update test, Delete not-found case, shared consts

### Step 4: HTTP handlers with Gin (`handlers.go`)
- [X] `go get github.com/gin-gonic/gin` (run `go mod tidy` to drop `// indirect`)
- [X] Request structs with `binding` tags
- [X] Response structs (`total_price`, per-item `unit_price` + `price`, embedding)
- [X] 5 handlers (create, update, get, checkout, delete)
- [X] Error responses and status codes (400 / 404 / 500)

### Step 5: Wiring (`main.go`)
- [X] Router group `/api/v1`, inject the store into the handlers
- [X] `PORT` from env with default
- [X] Manual test with Bruno
- [X] Bruno collection in `bruno/` (YAML format, `localhost` environment); optional: move `ecommerceId`/`customerId` from per-request vars into the environment

### Step 6: Optional
- [X] Handler tests with `httptest` (lifecycle, duplicate, 404s, validation 400s, malformed JSON; written by the assistant on request; 500 branches untested, needs a failing fake store)
- [X] Validation: `file_type` in `PDF|PSD|AI`, `quantity > 0` (done via `binding` tags, covered by tests)
- [ ] Money rounding (2 decimals / integer cents)
- [ ] MongoDB implementation of `CartStore`

### Step 7: Package split
- [X] Handlers moved to `internal/httpapi` (`cart.` prefixes, exported names, shadowed `cart` variable renamed)
- [X] `MemoryStore` moved to `internal/store/memory` (`memory.New()`, `memory.Store`), tests moved with it
- [X] `CartStore` interface and `Err...` values stay in `internal/cart` (domain vocabulary)
- [X] Compile-time check `var _ cart.CartStore = (*Store)(nil)`
- [ ] Optional: shared `RegisterRoutes` so `main.go` and the handler tests don't duplicate the route table

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
- Gin binding: `ShouldBindJSON(&req)`, `binding:"required,oneof=a b"`, rules before `dive` hit the slice and after it hit the elements, `required` zero-value trap
- Struct embedding: promoted fields, JSON flattens embedded structs, shallower field wins on a key clash (response structs extend domain types)
- `sync.Mutex`: `Lock` + `defer Unlock`, zero value usable, never copy (pointer receivers), goroutine-per-request makes it necessary
- `time.Time`, `time.Duration` (nanoseconds!), `t.Sub`, `t.Add`, `d.Hours()`
- Packages: folder = import path, `package` line = name used in code; organize by domain (`cart`, `store/memory`), not by kind (`models`, `types`)
- `internal/`: compiler-enforced, only importable from the parent of `internal`
- Cross-package visibility: only Uppercase names are reachable, referenced as `pkg.Name`
- A local variable can shadow an imported package name (`cart := ...`): rename it
- Methods must be declared in the same package as their type
- Interfaces are satisfied implicitly; define them where they are used; `var _ I = (*T)(nil)` is a compile-time check
- Package names are part of the identity: `memory.Store`, not `memory.MemoryStore`
- `net/http/httptest`: `NewRequest`, `NewRecorder`, `router.ServeHTTP`, `gin.TestMode`, `t.Helper()`
- Tests: `_test.go`, `testing.T`, table-driven tests, `t.Run`, float comparison with tolerance
- Watch a test fail before trusting it
