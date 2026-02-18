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
- Tests live alongside code in `*_test.go` files.
- Use table-driven tests and clear case names.
- Use `testify` for assertions in parameterized tests.
- Use `testify/suite` for setup & teardown behaviors.
- Cover edge cases and error paths; test plan should be explicit for complex features.

Example test:

```go
func TestNewWidget(t *testing.T) {
	testCases := []struct {
		name  string
		label string
		value int
		tag   string
		want  Widget
	}{
		{
			name:  "default tag if empty",
			label: "primary",
			value: 42,
			tag:   "",
			want: Widget{
				Label: "primary",
				Value: 42,
				Tag:   "default",
			},
		},
		{
			name:  "use provided tag",
			label: "secondary",
			value: 7,
			tag:   "custom-tag",
			want: Widget{
				Label: "secondary",
				Value: 7,
				Tag:   "custom-tag",
			},
		},
	}
	r := require.New(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := NewWidget(tt.label, tt.value, tt.tag)
			r.Equal(tt.want, got)
		})
	}
}
```

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
