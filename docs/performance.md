# Performance

`go-ruby-securerandom/securerandom` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's
`SecureRandom`. Its hex path runs on [go-simd/hex](https://github.com/go-simd/hex)
and its base64 paths on [go-simd/base64](https://github.com/go-simd/base64). This
page records the **methodology** for measuring it — both against the scalar
standard-library encoders and against the reference Ruby runtimes.

!!! note "No numbers are published here yet"
    This page documents *how* the comparison is run, not a result table. Numbers
    are only added once they have been measured on the host described below —
    never estimated or filled in from memory.

## Two comparisons

**1. SIMD encoders vs scalar (Go-internal).** The hex and base64 formatting paths
delegate to go-simd/hex and go-simd/base64, whose kernels are byte-identical
drop-ins for the standard library. Benchmarks time the SIMD kernel against the
scalar `encoding/hex` / `encoding/base64` reference across input sizes, isolating
the encoder speedup on whatever arch the bench runs on (amd64, arm64, ppc64le,
s390x). Reproduce:

```sh
go test -bench=. -benchmem ./...
```

Note the dominant cost of a `SecureRandom` call is the CSPRNG draw itself, which
is independent of this library; the SIMD win is on the **encoding** step.

**2. Ruby-visible operation (cross-runtime).** The **same** Ruby script — a mix of
`SecureRandom.hex` / `.base64` / `.uuid` / `.alphanumeric` calls — is run under
every runtime. `rbgo`'s number reflects **this pure-Go library doing the
formatting**; every other column is that interpreter's own `securerandom` stdlib.
Because both sides draw independent CSPRNGs, the comparison validates **shape and
throughput**, not byte-equality.

## How to reproduce the cross-runtime comparison

- **Host:** a single, recorded machine (CPU, OS, arch noted alongside any result
  table), so numbers are comparable run to run, and so the SIMD lane in use is
  unambiguous.
- **Method:** best-of-N wall time (best, not mean, to suppress scheduler noise);
  single-shot processes, no warm-up beyond the script's own loop.
- **Runtimes:** MRI (the oracle) and MRI `--yjit`; the JVM-based and GraalVM-based
  Rubies are timed **cold, single-shot**, so they carry VM startup on every run —
  read them as one-shot `ruby file.rb` costs, the same way `rbgo` and MRI are
  measured, not as steady-state JIT numbers.
- The benchmark script and harness live in rbgo's repo under
  [`bench/modules/`](https://github.com/go-embedded-ruby/ruby/tree/main/bench/modules).

!!! warning "Honest framing"
    Rows that complete in well under ~200 ms carry the most relative noise; treat
    their ratios as order-of-magnitude. Any numbers added here will be real
    measured numbers from a dated run, on a named host with the SIMD lane stated,
    with nothing cherry-picked.
