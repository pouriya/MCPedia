---
title: "Rust Testing"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, testing, unit-tests, cargo]
description: "Testing practices for Rust projects"
---

# Rust Testing

Rust’s built-in test harness and Cargo make it easy to write and run tests. Prefer unit tests close to the code and integration tests in `tests/`.

- Put unit tests in the same file under `#[cfg(test)]` with `mod tests { ... }`. Use `#[test]` and `assert!`, `assert_eq!`, `assert_ne!`. Use `Result<(), E>` in tests to propagate errors with `?`.
- Put integration tests in `tests/*.rs`; each file is a crate. Test public API and realistic workflows. Use a common helper module (e.g. `tests/common/mod.rs`) to share setup.
- Use `cargo test`; filter by name with `cargo test <substring>`. Use `#[ignore]` for slow tests and run them with `cargo test -- --ignored` when needed.
- Prefer table-driven or parameterized tests for multiple cases. Use `proptest` or `quickcheck` for property-based testing when it adds value.
