---
title: "Go HTTP Client"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, http, client, requests]
description: "Making HTTP requests with the Go client"
---

# Go HTTP Client

Use `net/http.Client` with timeout and optional transport. Create requests with context so they can be cancelled; handle redirects and status codes explicitly when needed.

- Create a client: `client := &http.Client{ Timeout: 10 * time.Second }` or set `Transport` for connection pooling and TLS. Use `http.DefaultClient` only for trivial scripts; it has no timeout. Reuse the same client for multiple requests to reuse connections.
- Build requests with `http.NewRequestWithContext(ctx, method, url, body)`; set headers and optional body. Call `client.Do(req)`; always close the response body (`defer resp.Body.Close()`). Check `resp.StatusCode` and handle non-2xx; read body with `io.ReadAll(resp.Body)` or a JSON decoder.
- Use context for cancellation and timeout; the request will be aborted when context is done. For retries, use a new request (or clone) and exponential backoff. Do not reuse the request body; create a new request per attempt if the body was consumed.
