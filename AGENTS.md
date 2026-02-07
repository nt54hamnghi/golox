# Repository Guidelines

## Project Structure & Module Organization
- `main.go` is the entry point for the interpreter.
- Core language pieces live at the repo root: `scanner.go`, `token.go`, `tokenType.go`, `expr.go`, and `printer.go`.
- Generator tooling is in `tool/` (e.g., `tool/generateAst.go`).
- Sample input lives in `test.lox`.

## Build, Test, and Development Commands
Use `just` for the common workflow:
- `just run <args>` runs the interpreter (e.g., `just run test.lox`).
- `just gen <args>` runs the AST generator (`tool/generateAst.go`).
- `just fmt` formats all Go files with `go fmt`.

Direct Go commands also work:
- `go run . <args>` runs the interpreter.
- `go fmt .` formats the project.

## Coding Style & Naming Conventions
- Go formatting is standard `go fmt`; keep files gofmt-clean.
- Indentation is tabs as enforced by `go fmt`.
- Names follow Go conventions: exported `CamelCase`, unexported `camelCase`, constants in `CamelCase` (or `SCREAMING_SNAKE` only if already established).
- Filenames mirror their responsibility (`scanner.go`, `token.go`).

## Testing Guidelines
There are currently no automated tests in the repository. If you add tests, prefer Go’s standard testing package with files named `*_test.go`, and document how to run them in this file (e.g., `go test ./...`).

## Commit & Pull Request Guidelines
Git history uses Conventional Commits with scopes, for example:
- `feat(scanning): string literal`
- `feat(scanning): reserved keywords and identifier`

Follow the same pattern for new commits: `<type>(<scope>): <summary>`.

For pull requests:
- Provide a short summary of behavior changes and reasoning.
- Include the commands you ran (e.g., `just fmt`, `go run . test.lox`).
- If the change affects language behavior, add or update a `.lox` example.

## Optional Notes
AST generation is a separate step; if you modify `expr.go` or related structures, ensure the generator and generated output stay in sync using `just gen`.
