# go-ruby-securerandom documentation

**Ruby's `SecureRandom` in pure Go — MRI-compatible, SIMD-accelerated, no cgo.**

`go-ruby-securerandom/securerandom` is a faithful, pure-Go (zero cgo)
reimplementation of Ruby's
[`SecureRandom`](https://docs.ruby-lang.org/en/master/SecureRandom.html) — the
formatting layer MRI 4.0.5 builds over a cryptographically secure entropy source.
`Hex`, `Base64`, `UrlsafeBase64`, `RandomBytes`, `Uuid` / `UuidV7`,
`Alphanumeric`, and `RandomNumber` all produce byte-identical output to MRI. The
module path is `github.com/go-ruby-securerandom/securerandom`.

The randomness comes from an injectable [`RandSource`](api.md#the-randsource-seam)
(the default is `crypto/rand`); everything else — hex/base64/URL-safe encoding, the
UUID bit-layout, the `#choose` builder behind `Alphanumeric`, and the
`random_number` distribution — is deterministic, interpreter-independent
formatting and lives here as pure Go. The hex path runs on
[**go-simd/hex**](https://github.com/go-simd/hex) and the base64 paths on
[**go-simd/base64**](https://github.com/go-simd/base64), both byte-identical SIMD
drop-ins for the standard library — a **CGO=0 SIMD path on all six of Go's 64-bit
targets**.

It was **extracted into a reusable standalone library**: importable by any Go
program with no Ruby runtime, and bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) by `rbgo` as a native
module — just like [go-ruby-regexp](https://github.com/go-ruby-regexp),
[go-ruby-erb](https://github.com/go-ruby-erb) and
[go-ruby-yaml](https://github.com/go-ruby-yaml).

!!! success "Status: complete — SecureRandom, MRI-faithful"
    **`Hex`** / **`Base64`** / **`UrlsafeBase64`** / **`RandomBytes`** (MRI's defaults), **`Uuid`** (RFC 4122 v4) and **`UuidV7`** (MRI 4.0's v7), **`Alphanumeric`** (the `#choose` builder), and **`RandomNumber`** / `RandomInt` / `RandomFloat`. The hex path is SIMD-accelerated via go-simd/hex and the base64 family via go-simd/base64. Validated by a **differential oracle** against the system `ruby` — format, length, charset and UUID bit-layout — at 100% coverage, `gofmt` + `go vet` clean, CI green across the six 64-bit Go targets and three OSes.

## Randomness vs. formatting

The randomness comes from an injectable `RandSource` (the default is
`crypto/rand`); everything else is deterministic, interpreter-independent
formatting and lives here as pure Go. Because the default source is
cryptographically random, callers (and the tests) assert on the **format, length,
charset and UUID bit-layout**, not on exact bytes — feed a fixed `RandSource` when
a deterministic result is needed.

## Quick taste

```go
securerandom.Hex()                    // 32 hex chars   (n=16)
securerandom.Base64()                 // 24 chars, padded
securerandom.UrlsafeBase64(16, false) // 22 chars, no padding
securerandom.Uuid()                   // v4: 4-version, 8/9/a/b-variant
securerandom.UuidV7()                 // v7: 48-bit unix-ms prefix + 7-version
securerandom.Alphanumeric(16)         // A-Za-z0-9
securerandom.RandomInt(100)           // Integer in [0,100)
securerandom.RandomFloat()            // Float   in [0,1)
```

## Repositories

| Repo | What it is |
| --- | --- |
| [`securerandom`](https://github.com/go-ruby-securerandom/securerandom) | the library — Ruby's `SecureRandom` in pure Go, SIMD-accelerated |
| [`docs`](https://github.com/go-ruby-securerandom/docs) | this documentation site (MkDocs Material, versioned with mike) |
| [`go-ruby-securerandom.github.io`](https://github.com/go-ruby-securerandom/go-ruby-securerandom.github.io) | the organization landing page (Hugo) |
| [`brand`](https://github.com/go-ruby-securerandom/brand) | logo and brand assets |

## Principles

- **Pure Go, `CGO_ENABLED=0`** — trivial cross-compilation, a single static
  binary, no C toolchain.
- **SIMD-accelerated, bit-identical.** The hex path runs on go-simd/hex and the
  base64 paths on go-simd/base64 — go-asmgen kernels across all six 64-bit Go
  arches, stdlib-identical output.
- **Randomness vs. formatting.** Entropy comes from an injectable `RandSource`;
  all encoding and layout is deterministic, pure-Go formatting that reproduces MRI
  exactly.
- **100% test coverage** is the target, enforced as a CI gate, across 6 arches
  and 3 OSes.

## Where to go next

- [Why pure Go](why.md) — the randomness/formatting split, and why the formatting
  lives as a standalone, interpreter-independent Go library.
- [Usage & API](api.md) — the public surface, the `RandSource` seam, and worked
  examples.
- [Roadmap](roadmap.md) — what is done and what is downstream by design.

Source lives at [github.com/go-ruby-securerandom/securerandom](https://github.com/go-ruby-securerandom/securerandom).
