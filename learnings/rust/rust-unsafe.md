---
title: "Rust Unsafe"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, unsafe, safety, invariants]
description: "Using unsafe Rust correctly and sparingly"
---

# Rust Unsafe

Use `unsafe` only when necessary; document invariants and wrap in a safe API so callers cannot violate them.

- Unsafe allows: dereferencing raw pointers, calling unsafe functions, accessing or modifying statics, implementing unsafe traits. The compiler cannot verify memory safety and thread safety in these blocks; the programmer must guarantee them.
- Prefer safe abstractions: use standard library types, well-reviewed crates, or your own safe wrappers. When you need unsafe, keep the block small and document the contract (e.g. "caller must ensure the pointer is valid and aligned"). Use `assert!` and `debug_assert!` to catch violations in development.
- Do not expose raw pointers or unsafe functions in the public API unless the type is explicitly an unsafe abstraction. Use `#[deny(unsafe_code)]` in application code if you want to forbid unsafe entirely. Prefer `MaybeUninit` and `NonNull` over raw pointers when they fit.
