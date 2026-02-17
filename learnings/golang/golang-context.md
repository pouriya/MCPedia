---
title: "Go Context"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, context, cancellation, timeout]
description: "Using context for cancellation and request scope in Go"
---

# Go Context

Use `context.Context` for cancellation, timeouts, and request-scoped values. Pass it as the first parameter to functions that do I/O or start goroutines.

- Create contexts with `context.Background()` at entry points and `context.TODO()` only as a temporary placeholder. Derive with `context.WithCancel`, `context.WithTimeout`, or `context.WithDeadline`; always call the cancel function to release resources.
- Pass context into HTTP handlers (from `r.Context()`), gRPC calls, DB queries, and any function that may block. Check `ctx.Err()` or `select { case <-ctx.Done(): return ctx.Err() }` in loops and long operations.
- Use context values only for request-scoped data (e.g. request ID, auth); use typed keys and avoid storing optional data that should be function parameters. Do not pass context for passing optional parameters.
- When spawning goroutines, pass a context so they can be cancelled. Prefer `errgroup.WithContext` when you need to run multiple operations and cancel on first error.
