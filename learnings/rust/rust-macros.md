---
title: "Rust Macros"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, macros, declarative, derive]
description: "When and how to use macros in Rust"
---

# Rust Macros

Use macros to reduce boilerplate or define DSLs. Prefer declarative macros and derive macros over procedural macros when they suffice.

- Declarative macros: `macro_rules! name { ... }` with pattern matching and repetition (`$( ... )*`, `$( ... ),*`). Use for print-style (e.g. `vec!`), assertions, or repeated structure. Keep hygiene in mind; use unique names.
- Derive macros: use `#[derive(...)]` for `Debug`, `Clone`, `Serialize`, or custom traits. Use procedural macros (e.g. `syn`, `quote`) when you need to implement custom derive; document the generated code.
- Prefer functions and generics when possible; use macros when you need to avoid runtime cost for variable arity or when you need to inspect or generate code structure. Document macro expansion behavior and any limitations.
