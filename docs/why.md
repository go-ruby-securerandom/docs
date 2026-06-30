# Why pure Go

`go-ruby-securerandom/securerandom` reimplements Ruby's `SecureRandom` in **pure
Go, with cgo disabled**. `SecureRandom` is really two things: a draw from a
cryptographically secure entropy source, and a **deterministic formatting layer**
on top of it. The formatting — hex, base64, URL-safe base64, the UUID bit-layout,
the `#choose` alphanumeric builder, the `random_number` distribution — is a pure
function of its random input. That is exactly the part that can — and should —
live as a standalone Go library, separate from the interpreter.

## Randomness vs. formatting

The randomness comes from an injectable `RandSource` (the default is
`crypto/rand`). Inject a fixed source and every formatter becomes deterministic —
which is exactly how the test suite asserts exact bytes. Because the default
source is cryptographically random, callers (and the differential oracle) assert
on the **format, length, charset and UUID bit-layout**, not on exact bytes, since
both this library and MRI draw from independent CSPRNGs.

## SIMD without cgo

The encoding paths are SIMD-accelerated without any C:

- the **hex** path runs on [**go-simd/hex**](https://github.com/go-simd/hex);
- the **base64** paths run on [**go-simd/base64**](https://github.com/go-simd/base64).

Both are [go-asmgen](https://github.com/go-asmgen) Plan 9 assembly kernels —
byte-identical drop-ins for the standard library that cross-compile and link into
a single static binary like the rest of the package. This is the same CGO=0 SIMD
approach the go-simd ecosystem applies across all six of Go's 64-bit targets.

## Extracted from rbgo, reusable by anyone

This library is the `SecureRandom` backend bound into
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby)'s `rbgo`, but it has
been **extracted into a reusable standalone library** so that:

- any Go program can import `github.com/go-ruby-securerandom/securerandom`
  directly, with no Ruby runtime;
- the dependency runs the *other* way — `rbgo` binds this module as a native
  module (the same pattern as [go-ruby-regexp](https://github.com/go-ruby-regexp),
  [go-ruby-erb](https://github.com/go-ruby-erb) and
  [go-ruby-yaml](https://github.com/go-ruby-yaml)), rather than this module
  depending on the interpreter;
- the formatting behaviour is pinned by a **differential oracle** against the
  system `ruby`, independent of any one consumer.

## Why pure Go matters here

Because the library is CGO-free and dependency-free, it:

- cross-compiles to every Go target with no C toolchain, and links into a single
  static binary — SIMD path included;
- has **no dependency on the Ruby runtime** — the dependency runs the other way;
- can be differentially tested against the `ruby` binary wherever one is on
  `PATH`, while the cross-arch lanes (where `ruby` is absent) still validate the
  library.

See [Usage & API](api.md) for the surface and [Roadmap](roadmap.md) for what is
in scope.
