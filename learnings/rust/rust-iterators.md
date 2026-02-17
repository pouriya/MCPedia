---
title: "Rust Iterators"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, iterators, functional, collect]
description: "Using iterators effectively in Rust"
---

# Rust Iterators

Prefer iterator chains over manual loops for clarity and zero-cost abstraction. Use `iter()`, `into_iter()`, and `iter_mut()` as appropriate.

- Chain methods: `map`, `filter`, `filter_map`, `flat_map`, `take`, `skip`, `zip`, `enumerate`, `chain`, `fold`, `reduce`, `find`, `any`, `all`, `collect`. Use `collect()` into `Vec`, `HashMap`, `String`, or custom types that implement `FromIterator`.
- Prefer `iter()` when borrowing, `into_iter()` when consuming ownership, and `iter_mut()` when mutating in place. Use `Iterator` trait methods; implement `Iterator` for custom types when they represent a sequence.
- Lazy: iterators do not run until consumed (e.g. by `next()`, `collect()`, or a for-loop). Avoid side effects in `map`/`filter` that you rely on; use `for_each` for side effects. Use `by_ref()` to reuse an iterator after partial consumption.
