---
title: "Rust Option and Result"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, option, result, combinators]
description: "Using Option and Result effectively in Rust"
---

# Rust Option and Result

`Option<T>` and `Result<T, E>` are the standard types for optional values and fallible operations. Use them and their combinators instead of null checks or exceptions.

- **Option**: Use `Some(x)` and `None`. Prefer `map`, `and_then`, `filter`, `unwrap_or`, `unwrap_or_else`, and `?` (in `Option`-returning code) over manual `match` when they keep code clear. Use `get_or_insert` for lazy default when mutating.
- **Result**: Use `Ok(x)` and `Err(e)`. Propagate with `?`; add context with `map_err` or `.context()` (anyhow). Use `Result::from_iter`, `collect()` into `Result`, and `unwrap_or_else` for fallbacks.
- Prefer returning `Option` or `Result` from functions over panicking or returning sentinel values. Document when `None` or `Err` can occur so callers handle them.
