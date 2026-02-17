---
title: "Go errgroup"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, errgroup, goroutines, cancellation]
description: "Running goroutines with errgroup for cancellation and error aggregation"
---

# Go errgroup

Use `golang.org/x/sync/errgroup` to run multiple goroutines and cancel all when one fails or when the context is done.

- Create with `errgroup.WithContext(ctx)` to get a group and a derived context that is cancelled when the first goroutine returns an error or when the group's `Wait` returns. Run work with `g.Go(func() error { ... })`; call `g.Wait()` to block and get the first non-nil error.
- Use when you have N independent operations (e.g. parallel HTTP calls, parallel DB queries) and want to stop all of them on first error or context cancel. Pass the group's context to child operations so they see cancellation.
- Do not use errgroup for unbounded goroutines; use worker pools or semaphores for that. Ensure each `Go` callback returns (and does not block indefinitely) so `Wait` can complete. Use for startup and shutdown orchestration as well as request-scoped parallelism.
