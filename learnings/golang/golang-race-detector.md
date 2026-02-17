---
title: "Go Race Detector"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, race, concurrency, testing]
description: "Using the race detector to find data races"
---

# Go Race Detector

Use the race detector (`-race`) in tests and in staging to find data races. Fix races with proper synchronization; do not ignore or hide them.

- Enable with `go test -race`, `go run -race`, or `go build -race`. The runtime will report when two goroutines access the same memory with at least one write without synchronization. Reports include stack traces for both goroutines.
- Run race-enabled tests in CI (they are slower and use more memory). Run integration tests and main flows under `-race`. Use a limited number of concurrent tests or short timeouts so the run stays feasible.
- Fix races by protecting shared data with a mutex, using channels to pass ownership, or avoiding shared mutable state. Do not use the race detector as the only concurrency test; combine with code review and clear ownership rules. Document any intentional lock-free or race-prone code if it cannot be avoided.
