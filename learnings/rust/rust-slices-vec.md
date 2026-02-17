---
title: "Rust Slices and Vec"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, slices, vec, collections]
description: "Slices, Vec, and common collection patterns in Rust"
---

# Rust Slices and Vec

Use slices (`&[T]`, `&mut [T]`) for views and `Vec<T>` for owned growable buffers. Prefer slice methods and avoid unnecessary allocations.

- Slices are a view (pointer + length); they are `Copy`-like (the reference). Use `&v[..]`, `&v[a..b]`, or `v.as_slice()`. Use methods: `len`, `is_empty`, `first`, `last`, `get`, `split_at`, `iter`, `windows`, `chunks`. Use `&str` for string slices.
- `Vec<T>`: grow with `push`, `extend`; shrink with `pop`, `truncate`, `clear`. Reserve capacity with `reserve`, `reserve_exact` to avoid reallocation. Use `into_iter()` to consume; use `Vec::from_iter` or `.collect()` to build. Prefer `vec![]` for literals.
- Prefer accepting `&[T]` or `&str` in functions when you do not need ownership; return `Vec` or `String` when you allocate. Use `Cow<[T]>` or `Cow<str>` when you might return either borrowed or owned.
