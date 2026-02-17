---
title: "Rust Ownership and Borrowing"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, ownership, borrowing, lifetimes]
description: "Ownership and borrowing rules for memory-safe Rust"
---

# Rust Ownership and Borrowing

Rust’s ownership model ensures memory safety without a garbage collector. Apply these rules when writing or reviewing Rust code.

- Each value has a single owner. Assigning or passing by value moves ownership; the previous owner cannot use the value.
- Use references to avoid moving: `&T` for shared (read-only), `&mut T` for exclusive (read-write). Only one `&mut` or many `&` at a time for the same data.
- Lifetimes tie references to the data they refer to. Use lifetime parameters when a function returns a reference or stores one; prefer elision when the compiler can infer.
- Prefer borrowing over cloning when possible. Clone only when you need an independent copy (e.g. `String`, `Vec`). Use `Cow` when you might need either borrowed or owned data.
