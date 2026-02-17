---
title: "Go Code Review"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, review, quality, style]
description: "What to look for when reviewing Go code"
---

# Go Code Review

When reviewing Go code, focus on correctness, error handling, concurrency safety, and idiomatic style.

- Check that all errors are handled or explicitly ignored with a comment. Look for missing `err != nil` checks and improper error wrapping. Ensure context is passed where needed and cancellation is respected.
- Check concurrency: no data races, proper use of channels or mutexes, no goroutine leaks. Ensure shared state is protected and that `go` statements have a clear exit path. Suggest `-race` in CI if not present.
- Check style: `gofmt`, naming, comment quality. Prefer small functions and clear control flow. Look for unnecessary allocations in hot paths and suggest benchmarks if performance is critical.
- Run `go vet` and staticcheck (or similar); ensure tests exist for new behavior and that table-driven tests are used where appropriate. Prefer constructive comments and small, incremental changes.
