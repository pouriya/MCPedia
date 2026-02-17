---
title: "Go Testing"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, testing, table-driven, benchmarks]
description: "Testing practices for Go code"
---

# Go Testing

Use the standard `testing` package and table-driven tests. Add benchmarks and fuzz tests where they add value.

- Put tests in `*_test.go` files in the same package (white-box) or in `_test` package (black-box). Use `func TestXxx(t *testing.T)` and `t.Error`, `t.Fatal`, or `t.Run` for subtests.
- Prefer table-driven tests: define a slice of `struct { name string; input ...; want ... }` and loop with `t.Run(tt.name, ...)`. Use `t.Helper()` in helpers so line numbers point to the caller.
- Use `go test -v`, `-run`, `-count`, and `-race`. Add benchmarks with `func BenchmarkXxx(b *testing.B)` and run with `go test -bench`. Use `b.ResetTimer()` and avoid per-iteration setup in the measured loop.
- Use `testing/fuzz` for fuzz tests when inputs are complex or security-sensitive. Mock external deps with interfaces and inject them in tests.
