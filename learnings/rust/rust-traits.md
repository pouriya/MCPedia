---
title: "Rust Traits"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, traits, polymorphism, derive]
description: "Defining and using traits in Rust"
---

# Rust Traits

Use traits for shared behavior and constraints. Prefer trait bounds on generics and `impl Trait` for clean APIs.

- Define traits with method signatures (and optional default implementations). Use `impl Trait for Type` to implement. Require `Self` or associated types when needed; use generic parameters for flexibility.
- Use trait bounds: `fn f<T: Clone + Debug>(x: T)` or `where T: Clone + Debug`. Use `impl Trait` in argument and return position to simplify signatures. Use `dyn Trait` for dynamic dispatch when you need type erasure (e.g. collections of different types).
- Derive common traits: `Debug`, `Clone`, `Copy`, `PartialEq`, `Eq`, `Hash`, `Default`, `Serialize`, `Deserialize`. Implement `From`/`Into` for conversions; use `TryFrom` for fallible conversion. Use `Send` and `Sync` for concurrency; only implement when safe.
