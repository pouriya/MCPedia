---
title: "Go Interfaces"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, interfaces, design, composition]
description: "Designing and using interfaces in Go"
---

# Go Interfaces

Use small, focused interfaces. Prefer "accept interfaces, return structs" so callers depend on behavior, not implementation.

- Define interfaces where they are consumed, not where they are implemented. One or a few methods per interface (e.g. `io.Reader`, `io.Writer`) keeps them flexible and easy to satisfy.
- Use interfaces for testing: pass a mock or stub that implements the same interface. Avoid large "god" interfaces; split into smaller ones (e.g. read vs write, or narrow operations).
- Prefer composition: embed types and interfaces to build behavior. Use interface types in function parameters and struct fields when you need polymorphism; use concrete types when you need a specific implementation.
- Do not define interfaces speculatively. Add them when you have at least two implementations or when you need to decouple (e.g. for tests or pluggable backends).
