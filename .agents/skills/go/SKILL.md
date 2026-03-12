---
name: go
description: How to develop Go projects.
---

Test:

```sh
go test ./... -race -covermode=atomic
```

Lint:

```sh
go vet ./...
golangci-lint run
```

Auto fix lint errors:

Note that only a few errors can be fixed by this command.
Many lint errors needs to be fixed manually.

```sh
golangci-lint fmt
```

## JSON Schema

If `cmd/gen-jsonschema` is present, JSON Schema for the application can be updated by the following command:

```sh
go run ./cmd/gen-jsonschema
```

## Test Framework Guidelines

- **DO NOT** use `testify` for writing tests
- **DO** use `google/go-cmp` for comparing expected and actual values

## File Naming Conventions

- Internal test files: append `_internal_test.go` for internal testing

## Error Handling

Use slog and [slog-error](github.com/suzuki-shunsuke/slog-error).
