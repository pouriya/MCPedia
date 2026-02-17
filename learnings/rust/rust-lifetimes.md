---
title: "Rust Lifetimes"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, lifetimes, references, borrowing]
description: "Understanding and using lifetimes in Rust"
---

# Rust Lifetimes

Lifetimes tie references to the data they refer to. Use explicit lifetime parameters when the compiler cannot infer them; prefer elision when it can.

- Every reference has a lifetime. The compiler uses them to ensure no reference outlives its data. Use `'a` in annotations: `fn f<'a>(x: &'a str) -> &'a str`. Use the same lifetime when input and output references must match (e.g. return a slice from a struct field).
- Elision rules: one input reference → one output lifetime; multiple inputs but one is `&self` or `&mut self` → output borrows from that. Add explicit `<'a>` when elision is ambiguous or when you need to relate multiple references (e.g. in structs holding references).
- Use `'static` only for data that lives for the whole program (e.g. string literals, leaked data). Prefer owned data or shorter lifetimes when possible. Use lifetime bounds in structs (`struct S<'a> { r: &'a str }`) and document who owns the underlying data.
