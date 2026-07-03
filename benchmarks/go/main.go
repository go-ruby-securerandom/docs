// SPDX-License-Identifier: BSD-3-Clause
//
// Cross-runtime library-level driver for go-ruby-securerandom/securerandom.
//
// SecureRandom draws the OS CSPRNG, so its output is nondeterministic and canNOT
// be byte-compared against the Ruby workload. Instead we verify each op returns
// the correct LENGTH / FORMAT (fixed by the API) before timing — mirroring the
// verify! block in ruby/securerandom.rb. The timed loop then measures the
// wrapper + hex/base64 encoding overhead (where go-simd helps), not entropy.
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"regexp"

	securerandom "github.com/go-ruby-securerandom/securerandom"
)

var uuidRE = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func verify(cond bool, msg string) {
	if !cond {
		fmt.Fprintln(os.Stderr, "verify failed:", msg)
		os.Exit(1)
	}
}

var hexRE = regexp.MustCompile(`^[0-9a-f]{32}$`)

func main() {
	// Format / length verification before timing (RNG output is nondeterministic).
	verify(hexRE.MatchString(securerandom.Hex(16)), "hex(16) -> 32 hex chars")
	verify(len(securerandom.RandomBytes(32)) == 32, "random_bytes(32) -> 32 bytes")
	verify(uuidRE.MatchString(securerandom.Uuid()), "uuid -> UUID format")
	if dec, err := base64.StdEncoding.DecodeString(securerandom.Base64(24)); err != nil || len(dec) != 24 {
		verify(false, "base64(24) -> 24 bytes")
	}
	rn := securerandom.RandomInt(1_000_000)
	verify(rn >= 0 && rn < 1_000_000, "random_number(1e6) -> [0,1e6)")

	// Timed workload — identical ops / inner counts to ruby/securerandom.rb.
	bench("hex-16", 20_000, func() { sink = securerandom.Hex(16) })
	bench("random_bytes-32", 20_000, func() { sink = securerandom.RandomBytes(32) })
	bench("uuid", 20_000, func() { sink = securerandom.Uuid() })
	bench("base64-24", 20_000, func() { sink = securerandom.Base64(24) })
	bench("random_number-1e6", 20_000, func() { sink = securerandom.RandomInt(1_000_000) })
}
