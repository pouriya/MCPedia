---
title: "Go Error Handling"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, errors, handling, wrap]
description: "Idiomatic error handling in Go"
---

# Go Error Handling

Go uses explicit error returns. Handle errors at the right layer; add context when wrapping so callers and logs are useful.

- Check every error unless you have a documented reason not to. Use `if err != nil { return ... }` or wrap with `fmt.Errorf("...: %w", err)` to preserve the chain. Use `errors.Is` and `errors.As` to inspect wrapped errors.
- Return errors from your functions; do not swallow them. In HTTP handlers or top-level entry points, log and translate to status codes or user-facing messages. Use sentinel errors (`var ErrNotFound = errors.New("...")`) for expected conditions when callers need to distinguish.
- Prefer `%w` in `fmt.Errorf` for wrapping so `errors.Unwrap`, `errors.Is`, and `errors.As` work. Add context (e.g. "failed to load config: %w") so debugging is easier. Avoid logging the same error at multiple layers.
- Consider `errors.Join` for aggregating multiple errors. Keep error types simple; use custom types only when callers need to switch on them.
