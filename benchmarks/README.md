<!-- SPDX-License-Identifier: BSD-3-Clause -->
# `go-ruby-securerandom` library-level benchmark harness

Reproducible, cross-runtime benchmark of the **pure-Go `go-ruby-securerandom`
library** against the reference Ruby runtimes (MRI, MRI + YJIT, JRuby,
TruffleRuby). It measures the **library primitive** through its Go API, isolated
from the rbgo interpreter, so the numbers answer: *is the pure-Go implementation
as fast as the reference runtime's own `SecureRandom`?*

## Layout

- `go/`               — self-contained Go driver; `go.mod` pins the published
  library by pseudo-version (no `replace`). The built `go/bench` binary is
  `.gitignore`d.
- `ruby/securerandom.rb` — the equivalent workload; `ruby/_harness.rb` is the
  shared timer.
- `run.sh`            — runs every available runtime and prints one Markdown table
  per sub-benchmark (ns/op + ratio vs MRI).

## Run

```sh
GOWORK=off bash benchmarks/run.sh
```

Environment knobs: `OUTER` (timed passes, default 25), `WARM` (untimed warm-up
passes, default 3), and `RUBY`/`JRUBY`/`TRUFFLERUBY` to select runtime binaries.

## Method

Each process runs `WARM` untimed passes (to let the JVM/GraalVM JITs warm up),
then `OUTER` timed passes of a fixed inner loop, timed with a monotonic clock;
the **best** pass is reported as **ns/op**. Interpreter start-up is outside the
timed region.

**RNG output is nondeterministic.** `SecureRandom` draws the OS CSPRNG, so its
output canNOT be byte-compared across runtimes. The Go driver and the Ruby script
instead **verify the LENGTH / FORMAT** of each op before timing (e.g. `hex(16)`
→ 32 hex chars, `uuid` matches the UUID regex, `base64(24)` decodes to 24 bytes,
`random_number(1_000_000)` ∈ [0, 1e6)). Both sides run the **same ops with the
same inner counts**. Because these ops all pull the same OS CSPRNG for entropy,
the comparison measures the **wrapper + hex/base64 encoding overhead** (where
go-simd helps), not the entropy source. Results are published, dated, in
`../docs/performance.md`.
