# Roadmap

`go-ruby-securerandom/securerandom` is grown **test-first**, each capability
differential-tested against MRI rather than built in isolation. Ruby's
`SecureRandom` formatting layer — the deterministic, interpreter-independent slice
over the entropy source — is **complete**.

| Stage | What | Status |
| --- | --- | --- |
| Encoders | `Hex(n)`, `Base64(n)`, `UrlsafeBase64(n, padding)` and `RandomBytes(n)` with MRI's defaults (`n = 16`; URL-safe base64 unpadded unless asked). Hex is SIMD-accelerated via go-simd/hex; the base64 family via go-simd/base64. | **Done** |
| UUID v4 | `Uuid()` — an RFC 4122 version-4 UUID with the exact `xxxxxxxx-xxxx-4xxx-yxxx-…` layout (version nibble `4`, variant `8`/`9`/`a`/`b`). | **Done** |
| UUID v7 | `UuidV7()` — MRI 4.0's version-7 UUID: a 48-bit big-endian Unix-milliseconds prefix from an injectable clock plus random bits, version nibble `7`. | **Done** |
| Alphanumeric & random_number | `Alphanumeric(n, chars…)` via the `Random::Formatter#choose` base-`size` construction; `RandomNumber` returning a `Float` in `[0,1)` for no/non-positive `n`, an `Integer` in `[0,n)` for a positive integer, a `Float` in `[0,n)` for a positive float (`RandomInt` / `RandomFloat` typed conveniences). | **Done** |
| SIMD acceleration | The hex path on [go-simd/hex](https://github.com/go-simd/hex) and the base64 paths on [go-simd/base64](https://github.com/go-simd/base64) — go-asmgen kernels, byte-identical drop-ins for the standard library, CGO=0 SIMD across all six 64-bit arches. | **Done** |
| Differential oracle & coverage | Each method run under the system `ruby` and checked for the same format, length, charset and UUID bit-layout (never exact bytes, since both sides draw independent CSPRNGs); 100% coverage, gofmt + go vet clean, green across all six 64-bit Go arches and three OSes. | **Done** |

## Documented out-of-scope boundaries

These are **deliberate**, recorded so the module's surface is unambiguous:

- **No interpreter.** The library implements the deterministic formatting over an
  injectable entropy source; it never runs arbitrary Ruby. That is why `rbgo`
  binds this module rather than the reverse.
- **Format, not exact bytes.** Because the default source is a CSPRNG, conformance
  is asserted on format, length, charset and UUID bit-layout — exact bytes are
  only reproducible when a fixed `RandSource` is injected.
- **Edge cases match MRI by behaviour.** Non-positive `random_number` returns a
  float in `[0,1)`; a single-element/empty `Alphanumeric` set returns the only
  reachable string rather than looping forever as MRI's `#choose` would.
- **Reference is reference Ruby (MRI 4.0.5).** Conformance targets MRI's
  behaviour, pinned by the differential oracle.

See [Usage & API](api.md) for the surface and [Why pure Go](why.md) for the
randomness/formatting split.
