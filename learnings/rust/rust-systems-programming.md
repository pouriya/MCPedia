---
title: "Rust Systems Programming"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, systems, performance, memory]
description: "High-performance and memory-safe systems programming in Rust"
---

# Rust Systems Programming

Rust is well-suited for systems code: zero-cost abstractions, no GC, and control over layout and allocation. Use it for performance-critical and low-level code.

- Prefer stack allocation and references. Use `Box`, `Vec`, or other collections when heap allocation is needed. Avoid unnecessary clones and allocations in hot paths.
- Use `unsafe` only when necessary; document invariants and wrap in safe APIs. Prefer standard and well-reviewed crates over hand-rolled unsafe code.
- Use iterators and slice operations for bulk data; leverage SIMD or dedicated crates when profiling shows benefit. Measure with `criterion` or similar.
- Consider `no_std` for embedded or kernel-style targets; use `alloc` when you need heap without the full standard library.
