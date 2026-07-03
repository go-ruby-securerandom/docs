# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
require "securerandom"
require "base64"
require_relative "_harness"

# ---------------------------------------------------------------------------
# Format / length verification BEFORE timing.
#
# SecureRandom pulls the OS CSPRNG, so its output is nondeterministic — it
# canNOT be byte-compared against the Go driver. Instead we assert each op
# returns the correct LENGTH / FORMAT, which the API fixes deterministically:
#   hex(16)             -> 32 lowercase hex chars
#   random_bytes(32)    -> 32 raw bytes
#   uuid                -> matches the RFC-4122 UUID regex
#   base64(24)          -> decodes to exactly 24 bytes
#   random_number(1e6)  -> Integer in [0, 1_000_000)
# ---------------------------------------------------------------------------
UUID_RE = /\A\h{8}-\h{4}-\h{4}-\h{4}-\h{12}\z/

def verify!(cond, msg)
  raise "verify failed: #{msg}" unless cond
end

verify!(SecureRandom.hex(16).match?(/\A\h{32}\z/), "hex(16) -> 32 hex chars")
verify!(SecureRandom.random_bytes(32).bytesize == 32, "random_bytes(32) -> 32 bytes")
verify!(SecureRandom.uuid.match?(UUID_RE), "uuid -> UUID format")
verify!(Base64.strict_decode64(SecureRandom.base64(24)).bytesize == 24, "base64(24) -> 24 bytes")
verify!((0...1_000_000).cover?(SecureRandom.random_number(1_000_000)), "random_number -> [0,1e6)")

# ---------------------------------------------------------------------------
# Timed workload. Each op = one CSPRNG draw plus the wrapper/encoding step.
# The entropy source is the same OS CSPRNG on both sides; this measures the
# wrapper + hex/base64 encoding overhead (where go-simd helps), not entropy.
# ---------------------------------------------------------------------------
bench("hex-16",        20_000) { SecureRandom.hex(16) }
bench("random_bytes-32", 20_000) { SecureRandom.random_bytes(32) }
bench("uuid",          20_000) { SecureRandom.uuid }
bench("base64-24",     20_000) { SecureRandom.base64(24) }
bench("random_number-1e6", 20_000) { SecureRandom.random_number(1_000_000) }
