---
title: "Rust Error Handling"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, errors, result, option]
description: "Idiomatic error handling in Rust for AI agents"
---

# Rust Error Handling

Use `Result<T, E>` for recoverable errors and `Option<T>` for optional values. Reserve `panic!` for truly unrecoverable situations (e.g. programming bugs), not expected failure cases.

- Prefer returning `Result` from fallible functions; use `?` to propagate errors. Convert at boundaries (e.g. `map_err`) to domain or library error types.
- Use enum-based error types for clarity. Implement `std::error::Error` and `Display`. Use `#[source]` for chaining and `#[from]` for conversion when using crates like `thiserror`.
- Avoid `unwrap()` and `expect()` in production code paths. Use them only in tests, examples, or where failure is impossible.
- For applications, consider `anyhow` for flexible error handling and context; for libraries, prefer explicit error types (e.g. with `thiserror`).
