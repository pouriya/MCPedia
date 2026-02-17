---
title: "Go Concurrency Patterns"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, concurrency, goroutines, channels]
description: "Goroutines, channels, and concurrency in Go"
---

# Go Concurrency Patterns

Use goroutines for concurrent work and channels or sync primitives for coordination. Prefer clear ownership and structured patterns over ad-hoc locking.

- Start work with `go f()`; ensure goroutines can exit (e.g. via context cancellation or a done channel). Avoid leaking goroutines; use `errgroup` or worker pools when you need bounded concurrency and error aggregation.
- Use channels for communication: unbuffered for synchronization, buffered when decoupling producer/consumer. Prefer "share memory by communicating" but use `sync.Mutex` or `sync.RWMutex` when shared state is simpler.
- Use `context.Context` for cancellation and timeouts; pass it as the first argument. Respect context in loops and I/O; check `ctx.Done()` or use `select` with `ctx.Done()`.
- Use `sync.WaitGroup` to wait for N goroutines; use `errgroup.Group` when you need to stop on first error. Run the race detector (`go test -race`) in CI.
