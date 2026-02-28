package main

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: generate <output dir>\n")
		os.Exit(64)
	}

	outputDir := args[0]

	defineAst(outputDir, "Expr", []typeDesc{
		{"Literal", []field{{"Value", "any"}}},
		{"Grouping", []field{{"Expression", "Expr"}}},
		{"Unary", []field{{"Operator", "Token"}, {"Right", "Expr"}}},
		{"Variable", []field{{"Name", "Token"}}},
		{"Assignment", []field{{"Name", "Token"}, {"Value", "Expr"}}},
		{"Binary", []field{{"Left", "Expr"}, {"Operator", "Token"}, {"Right", "Expr"}}},
	})

	defineAst(outputDir, "Stmt", []typeDesc{
		{"Expression", []field{{"Expression", "Expr"}}},
		{"Print", []field{{"Expression", "Expr"}}},
		{"Var", []field{{"Name", "Token"}, {"Initializer", "Expr"}}},
		{"Block", []field{{"Stmts", "[]Stmt"}}},
	})
}

type typeDesc struct {
	name      string
	fieldList []field
}

func (t typeDesc) fields() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, f := range t.fieldList {
			if !yield(f.name, f.typ) {
				return
			}
		}
	}
}

type field struct {
	name string
	typ  string
}

func defineAst(outputDir string, base string, types []typeDesc) error {
	b := strings.Builder{}

	fmt.Fprintln(&b, "package main")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "type %s interface {\n", base)
	fmt.Fprintf(&b, "\tAccept(visitor %sVisitor) (any, error)\n", base)
	fmt.Fprintln(&b, "}")
	fmt.Fprintln(&b)

	defineVisitor(&b, base, types)

	for _, t := range types {
		defineType(&b, base, t)
	}

	path := filepath.Join(outputDir, strings.ToLower(base)+".go")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return err
	}

	return nil
}

func defineVisitor(b *strings.Builder, base string, types []typeDesc) {
	fmt.Fprintf(b, "type %sVisitor interface {\n", base)

	for _, t := range types {
		fmt.Fprintf(b, "\tVisit%s%s(%s %s) (any, error)", t.name, base, strings.ToLower(base), t.name)
		fmt.Fprintln(b)
	}

	fmt.Fprintln(b, "}")
	fmt.Fprintln(b)
}

func defineType(b *strings.Builder, base string, t typeDesc) {
	fmt.Fprintf(b, "type %s struct {\n", t.name)

	for fname, ftype := range t.fields() {
		fmt.Fprintf(b, "\t%s %s\n", fname, ftype)

	}
	fmt.Fprintln(b, "}")

	// constructor
	fmt.Fprintln(b)
	defineTypeConstructor(b, t)

	// implement Expr interface
	fmt.Fprintf(b, "func (self %s) Accept(visitor %sVisitor) (any, error) {\n", t.name, base)
	fmt.Fprintf(b, "\treturn visitor.Visit%s%s(self)\n", t.name, base)
	fmt.Fprintln(b, "}")

	fmt.Fprintln(b)
}

func defineTypeConstructor(b *strings.Builder, t typeDesc) {
	fmt.Fprintf(b, "func New%s(", t.name)
	for i, f := range t.fieldList {
		if i > 0 {
			fmt.Fprint(b, ", ")
		}
		fmt.Fprintf(b, "%s %s", strings.ToLower(f.name), f.typ)
	}

	fmt.Fprintf(b, ") %s {\n", t.name)
	fmt.Fprintf(b, "\treturn %s{\n", t.name)

	for _, f := range t.fieldList {
		fmt.Fprintf(b, "\t\t%s: %s,\n", f.name, strings.ToLower(f.name))
	}

	fmt.Fprintln(b, "\t}")
	fmt.Fprintln(b, "}")
	fmt.Fprintln(b)
}
