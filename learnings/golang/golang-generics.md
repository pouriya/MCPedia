---
title: "Go Generics"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, generics, type-parameters, constraints]
description: "Using generics in Go 1.18+"
---

# Go Generics

Use type parameters when you need to write reusable code that works with multiple types without interface overhead or reflection. Use constraints from `golang.org/x/exp/constraints` or define your own.

- Declare type parameters with `[T constraint]` or `[K comparable, V any]`. Use `any` as alias for `interface{}` when you do not need constraints. Constraints restrict what types can be used (e.g. `constraints.Ordered` for comparison).
- Use generics for container types, map/reduce style functions, and shared logic across concrete types. Do not overuse: prefer interfaces when behavior is what matters and only one implementation exists.
- Use `~T` in constraints for types whose underlying type is T. Use type sets (interfaces with type terms) for complex constraints. Keep generic code readable; add a small example or test for each type parameter combination if non-obvious.
- Generics are compiled to concrete types; no runtime cost. Use `go build` and ensure all used type instantiations compile and tests pass.
