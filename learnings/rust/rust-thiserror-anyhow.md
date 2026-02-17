---
title: "Rust thiserror and anyhow"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, thiserror, anyhow, errors]
description: "Using thiserror and anyhow for error handling in Rust"
---

# Rust thiserror and anyhow

Use `thiserror` for library and application error types and `anyhow` for application-level error handling and context.

- **thiserror**: Derive `Error` and `Display` with `#[derive(thiserror::Error)]`. Use `#[from]` for automatic conversion from other error types, `#[source]` for chaining, and `#[error("...")]` for messages. Keep variants and context clear for consumers.
- **anyhow**: Use `anyhow::Result<T>` and `anyhow::Error` in applications when you need flexible error handling. Add context with `.context("...")` or `?`; use `anyhow::Context` trait. Convert to user-facing or log-friendly output as needed.
- Use thiserror in libraries so callers get typed errors; use anyhow in binaries and services where you want to bubble and log errors with context. At boundaries, convert between them with `From` or `.map_err()`.
