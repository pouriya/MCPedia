---
title: "Go Performance"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, performance, pprof, benchmarking]
description: "Profiling and optimizing Go programs"
---

# Go Performance

Measure before optimizing. Use pprof and benchmarks to find bottlenecks; reduce allocations and improve locality in hot paths.

- Use `go test -bench=.` and `testing.B` for micro-benchmarks. Use `b.ReportAllocs()` to see allocations. Avoid compiler optimizations that remove your code (e.g. use result or sink). Compare with benchstat across runs.
- Use pprof for CPU and memory: import `_ "net/http/pprof"` and hit `/debug/pprof/profile` and `/debug/pprof/heap`. Use `go tool pprof` to inspect; find hot functions and allocation sites. Use trace for latency and scheduling.
- Reduce allocations in hot paths: reuse buffers (e.g. `sync.Pool`), avoid unnecessary string concatenation (use `strings.Builder`), and prefer passing pointers or reusing slices when appropriate. Prefer small, cache-friendly data structures.
- Use the race detector (`go test -race`, `go run -race`) to find data races. Fix races with proper synchronization; do not rely on "benign" races.
