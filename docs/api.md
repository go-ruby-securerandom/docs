# Usage & API

The public API lives at the module root
(`github.com/go-ruby-securerandom/securerandom`). It is **Ruby-shaped but
Go-idiomatic**: `Hex` / `Base64` / `Uuid` / … mirror MRI's `SecureRandom`
methods, while the surface follows Go conventions — an injectable `RandSource`,
variadic optional counts, typed conveniences, no hidden global state beyond a
swappable default source.

!!! success "Status: implemented"
    The library is built and importable as
    `github.com/go-ruby-securerandom/securerandom`, bound into `rbgo` as a native
    module; see [Roadmap](roadmap.md).

## Install

```sh
go get github.com/go-ruby-securerandom/securerandom
```

## Worked example

```go
fmt.Println(securerandom.Hex())              // 32 hex chars
fmt.Println(securerandom.Base64())           // 24 chars, padded
fmt.Println(securerandom.UrlsafeBase64(16, false)) // 22 chars, no padding
fmt.Println(securerandom.Uuid())             // 4-version, 8/9/a/b-variant
fmt.Println(securerandom.UuidV7())           // 7-version, unix-ms prefix
fmt.Println(securerandom.Alphanumeric(16))   // A-Za-z0-9

fmt.Println(securerandom.RandomInt(100))     // Integer in [0,100)
fmt.Println(securerandom.RandomFloat())      // Float   in [0,1)
```

## The `RandSource` seam

Every method ultimately draws from a `RandSource`; the default is `crypto/rand`.
Injecting a fixed source makes the formatters deterministic — exactly how the
test suite asserts exact bytes:

```go
type RandSource interface {
	Read(p []byte) (int, error)
}

// Bind a generator to a specific source (nil -> crypto/rand).
s := securerandom.New(myFixedSource)
s.Hex(4) // deterministic given myFixedSource

// Or swap the package-level default (restore it after):
securerandom.DefaultSource = myFixedSource
```

The top-level `Hex`, `Base64`, … functions read `securerandom.DefaultSource`; the
methods on `*SecureRandom` are preferred for concurrent or test use. A broken
source (a `Read` error) panics, mirroring MRI raising on a failed CSPRNG — a
condition that does not occur on the platforms targeted here.

## Shape

```go
func New(src RandSource) *SecureRandom // nil src -> crypto/rand

func (s *SecureRandom) RandomBytes(n ...int) []byte                 // default 16
func (s *SecureRandom) Hex(n ...int) string                        // 2*n chars
func (s *SecureRandom) Base64(n ...int) string                     // StdEncoding
func (s *SecureRandom) UrlsafeBase64(n int, padding bool) string   // URL alphabet
func (s *SecureRandom) Uuid() string                               // v4
func (s *SecureRandom) UuidV7() string                             // v7
func (s *SecureRandom) Alphanumeric(n int, chars ...string) string // #choose
func (s *SecureRandom) RandomNumber(n ...float64) (i int64, f float64, isInt bool)
func (s *SecureRandom) RandomInt(n int64) int64                    // [0,n)
func (s *SecureRandom) RandomFloat() float64                       // [0,1)

// Package-level equivalents read DefaultSource:
func Hex(n ...int) string
func Base64(n ...int) string
func UrlsafeBase64(n int, padding bool) string
func RandomBytes(n ...int) []byte
func Uuid() string
func UuidV7() string
func Alphanumeric(n int, chars ...string) string
func RandomNumber(n ...float64) (int64, float64, bool)
func RandomInt(n int64) int64
func RandomFloat() float64

var DefaultSource RandSource // defaults to crypto/rand
```

## Method semantics

- **`Hex(n)` / `Base64(n)` / `UrlsafeBase64(n, padding)` / `RandomBytes(n)`** —
  the encoders, with MRI's defaults (`n = 16`; URL-safe base64 unpadded unless
  asked). Hex is SIMD-accelerated via go-simd/hex; the base64 family via
  go-simd/base64.
- **`Uuid()`** — an RFC 4122 version-4 UUID with the exact
  `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx` layout (version nibble `4`, variant
  `8`/`9`/`a`/`b`).
- **`UuidV7()`** — MRI 4.0's version-7 UUID: a 48-bit big-endian Unix-milliseconds
  prefix (from an injectable clock) plus random bits, version nibble `7`.
- **`Alphanumeric(n, chars…)`** — MRI's `A-Z a-z 0-9` default (or a supplied
  character set), built with the same `Random::Formatter#choose` base-`size` digit
  construction so the distribution matches.
- **`RandomNumber(n)`** — no/zero/non-positive argument → a `Float` in `[0, 1)`; a
  positive integer → an `Integer` in `[0, n)`; a positive float → a `Float` in
  `[0, n)`. `RandomInt` / `RandomFloat` are typed conveniences.

## Notes on MRI fidelity

- `random_number` with a **non-positive** bound returns a `Float` in `[0, 1)`,
  matching MRI (which never raises there).
- `Alphanumeric` with a **single-element or empty** character set: MRI's `#choose`
  loops forever (no power of `size` exceeds the batch threshold). This library
  returns the only reachable string instead — `n` copies of the element, or `""`
  for an empty set — rather than hanging.

## MRI conformance

Correctness is defined by reference Ruby. A **differential oracle** runs each
method under the system `ruby` and checks the output for the same **format,
length, charset, and UUID bit-layout** — never exact bytes, since both sides draw
from independent CSPRNGs. The oracle scripts `$stdout.binmode` so a text-mode
runner never corrupts the bytes, and skip themselves on Windows and where `ruby`
is absent (e.g. the qemu arch lanes), so the cross-arch builds still validate the
library.

## Relationship to Ruby

`go-ruby-securerandom/securerandom` is **standalone and reusable**, and is the
`SecureRandom` backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a native
module — the same way [go-ruby-regexp](https://github.com/go-ruby-regexp),
[go-ruby-erb](https://github.com/go-ruby-erb) and
[go-ruby-yaml](https://github.com/go-ruby-yaml) are bound. The dependency runs the
other way: this library has no dependency on the Ruby runtime.
