---
title: "Effective Go"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, effective-go, idioms, style]
description: "Official Go style and idioms for AI agents"
---

# Effective Go

Follow the Effective Go guide and standard Go style so code is consistent and maintainable.

- Use `gofmt` (and enforce in CI). Keep names short and scoped: short in small scope, longer when needed for clarity. Use `MixedCaps` or `mixedCaps`; no underscores in names.
- Prefer composition over inheritance. Use small interfaces (one or few methods); define them where they are used. Accept interfaces and return concrete types when it makes sense.
- Use table-driven tests; keep tests in `*_test.go` files. Use `t.Helper()` for test helpers. Prefer `testing` and standard library; use testify or similar only if the team agrees.
- Document exported names with clear comments. Use `//` for line comments and doc comments for packages and exported symbols. Run `go vet` and `staticcheck`; fix warnings.
