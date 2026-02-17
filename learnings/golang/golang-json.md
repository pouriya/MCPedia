---
title: "Go JSON"
kind: skill
language: golang
domain: ""
project: ""
tags: [golang, json, encoding, api]
description: "Encoding and decoding JSON in Go"
---

# Go JSON

Use `encoding/json` for JSON serialization. Use struct tags to map fields; handle optional fields and unknown types explicitly.

- Marshal with `json.Marshal(v)`; unmarshal with `json.Unmarshal(data, &v)`. Use struct tags: `json:"field_name"`, `json:"field,omitempty"`, `json:"-"`. Use `json.Number` for numbers when you need to preserve precision or distinguish int/float.
- For streams use `json.Encoder`/`json.Decoder` with `Encode`/`Decode`. For HTTP: encode with `json.NewEncoder(w).Encode(v)` and set `Content-Type: application/json`; decode with `json.NewDecoder(r.Body).Decode(&v)` and check errors.
- Handle optional fields with pointers (`*string`, `*int`) or custom types; use `interface{}` or `map[string]interface{}` only when the structure is dynamic. Validate after unmarshaling when business rules apply. Return clear errors for invalid or unexpected JSON.
