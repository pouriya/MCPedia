---
title: "Go Slices and Maps"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, slices, maps, collections]
description: "Slices and maps in Go: usage and pitfalls"
---

# Go Slices and Maps

Slices are variable-length views over arrays; maps are hash tables. Understand reference semantics and when to copy or preallocate.

- Slices: `make([]T, len, cap)`, `append`, `copy`, and slicing `s[a:b]` share the same backing array. Do not assume a slice is independent after append (capacity may have been reused). Preallocate with `make([]T, 0, n)` when you know an upper bound to reduce allocations.
- Maps: `make(map[K]V)`; keys must be comparable. Check with `v, ok := m[k]`. Do not assume iteration order. Maps are reference types; pass by value is cheap but modifications are visible to all references. Initialize with `make` or a literal; nil map is read-only (no insert).
- Prefer slices for ordered sequences; use maps for lookup by key. Use `range` for iteration. For concurrent access to maps, use `sync.RWMutex` or `sync.Map` (when specific patterns apply). Avoid storing slice/map values that you will mutate later if other code holds a reference.
