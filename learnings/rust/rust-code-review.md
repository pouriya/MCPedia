---
title: "Rust Code Review Guidelines"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, review, safety, idioms]
description: "What to look for when reviewing Rust code"
---

# Rust Code Review Guidelines

When reviewing Rust code, focus on safety, clarity, and idiomatic use of the language and ecosystem.

- Check ownership and borrowing: no unnecessary clones, no redundant `&` or `&mut`, correct use of references vs owned types. Look for lifetime issues and unnecessary `'static` or `clone()`.
- Check error handling: prefer `Result` and `?`; avoid `unwrap()`/`expect()` in library or production paths. Ensure error types are meaningful and chainable where appropriate.
- Check concurrency: no data races; correct use of `Send`/`Sync`; appropriate use of `Arc`, `Mutex`, or channels. Ensure async code does not block the executor.
- Run `cargo clippy` and `cargo test`; ensure new code is formatted and documented. Prefer small, focused PRs and incremental refactors.
