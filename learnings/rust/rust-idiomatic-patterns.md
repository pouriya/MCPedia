---
title: "Rust Idiomatic Patterns"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, idiomatic, patterns, style]
description: "Idiomatic Rust style and patterns for maintainable code"
---

# Rust Idiomatic Patterns

Follow Rust conventions so code is readable and maintainable. Use the standard library and common crates idiomatically.

- Prefer `Option` and `Result` combinators: `map`, `and_then`, `unwrap_or`, `ok_or`. Use `match` when you need to handle each variant explicitly.
- Use iterators instead of manual loops where clear: `iter()`, `into_iter()`, `iter_mut()`, and methods like `filter`, `map`, `collect`, `find`, `fold`.
- Prefer `&str` for string views and `String` for owned strings. Use `format!` for simple formatting and avoid unnecessary allocations.
- Run `cargo clippy` and `cargo fmt`; fix warnings. Use `#[must_use]` and `#[allow(...)]` only when justified. Prefer `Default`, `From`, and `Into` for conversions.
