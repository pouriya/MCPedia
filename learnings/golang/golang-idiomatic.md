---
title: "Go Idiomatic Style"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, idioms, style, proverbs]
description: "Idiomatic Go and common proverbs"
---

# Go Idiomatic Style

Follow Go proverbs and community conventions: simplicity, clarity, and explicit over clever.

- "A little copying is better than a little dependency." Prefer small, focused packages and avoid large frameworks when the standard library suffices. Do not over-abstract early.
- "Clear is better than clever." Write readable code; avoid clever tricks. Use early returns and guard clauses; keep nesting shallow. Name variables and functions for clarity.
- "Error values are values." Handle errors explicitly; use `%w` for wrapping. Design APIs so that errors are easy to handle and inspect. Do not use panics for control flow.
- "Concurrency is not parallelism." Use goroutines and channels for structure; use sync primitives when sharing state. Prefer composition and interfaces. Use `go doc` and comments to document intent.
