---
title: "Go vet and staticcheck"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, vet, staticcheck, linting]
description: "Static analysis with go vet and staticcheck"
---

# Go vet and staticcheck

Run `go vet` and staticcheck (or similar) in CI to catch common bugs and style issues. Fix reported findings or justify exceptions.

- `go vet`: built-in checks for suspicious constructs (e.g. printf args, unreachable code, copylocks, struct tags). Run with `go vet ./...` or as part of `go test` with a suitable setup. Enable additional analyzers with `go vet -vettool=...` if needed.
- staticcheck (honnef.co/go/tools): more checks for correctness, style, and performance. Configure with a config file or flags; disable specific checks only when justified. Use for nil checks, error handling, redundant code, and API misuse.
- Run in CI and block merges on failures. Fix or document false positives; do not disable checks globally. Keep Go version and tool versions consistent so results are reproducible. Add a single `staticcheck.conf` or similar at repo root if the team agrees.
