---
title: "Rust Concurrency"
kind: skill
language: rust
domain: ""
project: ""
tags: [rust, concurrency, threads, sync]
description: "Threads, Sync, and shared-state concurrency in Rust"
---

# Rust Concurrency

Use threads, `Send`/`Sync`, and standard sync primitives for shared-state or CPU-bound concurrency. Prefer message passing (channels) where it fits.

- Use `std::thread::spawn` for CPU-bound work; ensure closure and data are `Send`. Use `JoinHandle` to wait and propagate panics or results. Prefer `std::sync::mpsc` or crossbeam channels for thread communication.
- Use `Mutex<T>` and `RwLock<T>` for shared mutable state; keep critical sections short. Prefer `Arc<Mutex<T>>` or `Arc<RwLock<T>>` when sharing across threads. Avoid holding locks across await points in async code.
- Respect `Send` and `Sync`: only implement them when the type is safe to transfer or share across threads. Use `Rayon` for data-parallel iteration when appropriate.
- Prefer higher-level patterns (e.g. worker pools, pipelines) over raw locks when they simplify reasoning and avoid deadlocks.
