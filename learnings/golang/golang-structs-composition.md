---
title: "Go Structs and Composition"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, structs, composition, embedding]
description: "Struct design and composition in Go"
---

# Go Structs and Composition

Use structs for data and embed types to compose behavior. Prefer composition and small interfaces over deep inheritance.

- Define structs with exported fields for external use and unexported for internal state. Use struct literals and constructor functions when validation is needed. Use tags for JSON, DB, and validation. Prefer value types for small structs; use pointers when you need to distinguish nil or share the same instance.
- Embedding: embed a type without a field name to get its methods on the outer type. Use for composition (e.g. embed `sync.Mutex` for locking, embed a logger or config). Avoid embedding when the relationship is "has-a" and you want to expose only specific methods; use a named field instead.
- Keep structs focused; split large structs or use embedding to group related fields. Use interfaces for behavior that can vary (e.g. store an `io.Writer` instead of a concrete type). Document zero value meaning and when pointers are required.
