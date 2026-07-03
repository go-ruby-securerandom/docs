# Performance

`go-ruby-securerandom/securerandom` is the pure-Go library that
[`rbgo`](https://github.com/go-embedded-ruby/ruby) binds for Ruby's
`SecureRandom`. Its hex path runs on [go-simd/hex](https://github.com/go-simd/hex)
and its base64 paths on [go-simd/base64](https://github.com/go-simd/base64). This
page records the **methodology** for measuring it — both against the scalar
standard-library encoders and against the reference Ruby runtimes.

## Result (best of 5, ms)

Measured 2026-06-30 on **Apple M4 Max**, macOS (darwin/arm64), Go 1.26.4, with
`ruby 4.0.5 +PRISM`, `jruby 10.1.0.0` (OpenJDK 25) and `truffleruby 34.0.1`
(GraalVM CE Native). The cross-runtime workload is a
`hex(32)` / `base64(32)` / `uuid` / `random_bytes(48)` loop; since the bytes are
random, the script checksums only the produced **lengths** (fixed by the API),
which is identical across runtimes.

| Runtime | time | vs MRI |
| --- | ---: | ---: |
| **rbgo** (go-ruby-securerandom) | 120 | **0.36×** |
| MRI (ruby 4.0.5) | 330 | 1.00× |
| MRI + YJIT | 310 | 0.94× |
| JRuby 10.1.0.0 | 1220 | 3.70× |
| TruffleRuby 34.0.1 | 270 | 0.82× |

rbgo runs on **go-ruby-securerandom** and is **~2.8× faster than MRI** here
(0.36×) — the SIMD payoff of this module. The Ruby-visible cost is a CSPRNG draw
*plus* a hex/base64 **encode**: go-ruby-securerandom routes the encode through
go-simd/hex and go-simd/base64 (NEON on arm64) while MRI hex-encodes in its C
stdlib byte-by-byte, and Go's `crypto/rand` supplies the entropy cheaply. Combined,
the pure-Go library wins outright on this loop.

!!! note "Honest framing"
    JRuby and TruffleRuby are timed **cold, single-shot**, so they carry JVM /
    Graal startup on every run — read them as one-shot `ruby file.rb` costs, the
    same way `rbgo` and MRI are measured, not as steady-state JIT numbers. These
    are **real measured numbers** from the 2026-06-30 run (Apple M4 Max;
    `ruby 4.0.5 +PRISM`, `jruby 10.1.0.0`, `truffleruby 34.0.1`) — nothing is
    fabricated or cherry-picked.

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

## Library-level benchmark (Go API vs runtimes) — 2026-07-03

This section measures the **pure-Go library directly, through its Go API** — not
the `rbgo` interpreter path recorded above. It isolates the library primitive
from Ruby-interpreter dispatch, answering the parity question head-on: *is the
pure-Go implementation as fast as the reference runtime's own `SecureRandom`?*
The **same ops, same inputs (fixed byte counts), same iteration counts** run
through the Go library and through each reference runtime's stdlib.

!!! danger "RNG output is nondeterministic — format/length verified, NOT bytes"
    `SecureRandom` draws the OS CSPRNG on every call, so its output is
    **nondeterministic and canNOT be byte-compared** across runtimes (unlike a
    pure encoder such as base64). Both drivers therefore **verify the LENGTH /
    FORMAT** of each op before timing — `hex(16)` → 32 hex chars,
    `random_bytes(32)` → 32 bytes, `uuid` matches the RFC-4122 UUID regex,
    `base64(24)` decodes to exactly 24 bytes, `random_number(1_000_000)` ∈
    [0, 1e6) — never the byte values. What the numbers isolate is the
    **wrapper + hex/base64 encoding overhead** (where go-simd/hex and
    go-simd/base64 help), *not* the entropy source: every runtime pulls the same
    OS CSPRNG for entropy, so that cost is common to all columns.

- **Host:** Apple M4 Max (`Mac16,5`, arm64), macOS 26.5.1 — **date 2026-07-03**.
- **Runtimes:** Go 1.26.4 · MRI `ruby 4.0.5 +PRISM` · MRI + YJIT · JRuby 10.1.0.0
  (OpenJDK 25) · TruffleRuby 34.0.1 (GraalVM CE Native).
- **Method:** each process runs 3 untimed warm-up passes, then 25 timed passes of
  a fixed inner loop (20 000 ops), timed with a monotonic clock; the **best** pass
  is reported as **ns/op** (lower is better). `vs MRI` < 1.00× means *faster than
  MRI*. Interpreter start-up is outside the timed region, so these are operation
  costs, not `ruby file.rb` process costs.

#### hex-16

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 233.1 | 0.20× |
| MRI | 1151.8 | 1.00× |
| MRI + YJIT | 1081.5 | 0.94× |
| JRuby | 160.7 | 0.14× |
| TruffleRuby | 323.7 | 0.28× |

#### random_bytes-32

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 215.9 | 0.20× |
| MRI | 1086.2 | 1.00× |
| MRI + YJIT | 997.7 | 0.92× |
| JRuby | 225.8 | 0.21× |
| TruffleRuby | 524.9 | 0.48× |

#### uuid

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 229.4 | 0.15× |
| MRI | 1575.3 | 1.00× |
| MRI + YJIT | 1419.9 | 0.90× |
| JRuby | 431.4 | 0.27× |
| TruffleRuby | 463.1 | 0.29× |

#### base64-24

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 244.4 | 0.21× |
| MRI | 1190.7 | 1.00× |
| MRI + YJIT | 1087.3 | 0.91× |
| JRuby | 234.0 | 0.20× |
| TruffleRuby | 487.1 | 0.41× |

#### random_number-1e6

| Runtime | ns/op | vs MRI |
| --- | ---: | ---: |
| **go-ruby (pure Go)** | 135.3 | 0.14× |
| MRI | 988.0 | 1.00× |
| MRI + YJIT | 680.9 | 0.69× |
| JRuby | 82.9 | 0.08× |
| TruffleRuby | 116.5 | 0.12× |

go-ruby-securerandom **beats MRI on every op** (0.14×–0.21×) — roughly **5× faster**.
Two effects compound: Go's `crypto/rand` supplies entropy cheaply, and the hex /
base64 formatting rides the [go-simd/hex](https://github.com/go-simd/hex) and
[go-simd/base64](https://github.com/go-simd/base64) NEON kernels, whereas MRI
hex-encodes byte-by-byte in its C stdlib. On the shortest loops (`random_number`,
`hex`, `base64`) warm **JRuby** edges out the pure-Go column: its JIT has fully
warmed these trivial ops within the warm-up budget and it draws from a JVM CSPRNG
— read those as steady-state JIT rows, not one-shot costs. Since entropy is the
common cost on both sides, what this table really compares is **wrapper + encoding
overhead**, and there the pure-Go library is at or ahead of MRI throughout.

!!! note "Reproduce"
    The harness is committed under
    [`benchmarks/`](https://github.com/go-ruby-securerandom/docs/tree/main/benchmarks):
    a self-contained Go driver (`go/`, pins the published library via `go.mod`
    pseudo-version, no `replace`), the equivalent `ruby/securerandom.rb` workload,
    and `run.sh`. Run `GOWORK=off bash benchmarks/run.sh`; env `OUTER`/`WARM` tune
    the pass budget and `RUBY`/`JRUBY`/`TRUFFLERUBY` select the runtime binaries.

!!! warning "Warm-up budget & noise — honest framing"
    Numbers reflect a **fixed warm-process budget** (3 warm-up + 25 timed passes
    in one process). The JVM/GraalVM JITs (JRuby, TruffleRuby) may need a larger
    warm-up to reach steady state, so their columns can **understate** peak
    throughput. These are all **sub-microsecond** ops, which carry the most
    relative noise; treat the ratios as order-of-magnitude. Every number here is a
    **real measured value** from the dated run above — nothing is fabricated,
    estimated, or cherry-picked. Because the RNG output is nondeterministic, only
    the **format/length** of each op is verified (never the bytes). The go-ruby
    column is the pure-Go library; every other column is that interpreter's own
    `securerandom` stdlib doing the equivalent work.
