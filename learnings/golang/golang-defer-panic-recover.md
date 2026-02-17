---
title: "Go defer, panic, recover"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, defer, panic, recover]
description: "Using defer for cleanup and handling panics in Go"
---

# Go defer, panic, recover

Use `defer` for cleanup (close files, unlock mutexes); use `panic` only for truly unrecoverable cases. Use `recover` only at the top of a goroutine to convert panics into errors.

- Defer runs when the function returns (normal or panic), in LIFO order. Use for `defer f.Close()`, `defer mu.Unlock()`, or custom cleanup. Pass arguments to defer at the call site; use a closure when you need the value at execution time. Do not overuse defer in hot paths.
- Panic stops normal execution and unwinds the stack until a recover or the program exits. Use panic for programmer errors (e.g. nil dereference, impossible state), not for expected failures—return errors instead. In tests, use `recover` or test that code panics when expected.
- Recover is only useful inside a deferred function; it captures a panic in the same goroutine and returns the value passed to panic. Use at the edge (e.g. HTTP server handler wrapper) to log and return 500 instead of crashing. Do not use recover to hide bugs; fix the cause.
