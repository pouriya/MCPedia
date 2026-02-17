---
title: "Rust Custom Error Types"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, errors, enums, thiserror]
description: "Designing and implementing error types in Rust"
---

# Rust Custom Error Types

Design error types as enums with clear variants; implement `Error` and `Display` and use `source` for chaining. Use thiserror to reduce boilerplate.

- Use an enum for multiple error cases: `enum AppError { Io(io::Error), Parse(ParseError), NotFound }`. Implement `Display` with user- or log-friendly messages. Implement `Error` with `fn source(&self)` returning inner errors where applicable.
- Use `#[derive(thiserror::Error)]` and `#[from]` for automatic `From` and `source`; use `#[error("...")]` per variant. Use `#[source]` for chaining. Provide context in application code with `.map_err()` or `anyhow::Context`.
- In libraries, expose typed errors so callers can match or use `errors::Is`/`As`. Use sentinel variants (e.g. `NotFound`) for expected cases. Avoid exposing implementation details in error messages; use debug or separate error codes if needed.
