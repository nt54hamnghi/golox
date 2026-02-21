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

	types := []typeDesc{
		{"Literal", []field{{"Value", "any"}}},
		{"Grouping", []field{{"Expression", "Expr[T]"}}},
		{"Unary", []field{{"Operator", "Token"}, {"Right", "Expr[T]"}}},
		{"Binary", []field{{"Left", "Expr[T]"}, {"Operator", "Token"}, {"Right", "Expr[T]"}}},
	}

	defineAst(outputDir, "Expr", types)
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
	fmt.Fprintln(&b, "type Expr[T any] interface {")
	fmt.Fprintf(&b, "\tAccept(visitor Visitor[T]) (T, error)\n")
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
	fmt.Fprintln(b, "type Visitor[T any] interface {")

	for _, t := range types {
		fmt.Fprintf(b, "\tVisit%s%s(expr %s[T]) (T, error)", t.name, base, t.name)
		fmt.Fprintln(b)
	}

	fmt.Fprintln(b, "}")
	fmt.Fprintln(b)
}

func defineType(b *strings.Builder, base string, t typeDesc) {
	fmt.Fprintf(b, "type %s[T any] struct {\n", t.name)

	for fname, ftype := range t.fields() {
		fmt.Fprintf(b, "\t%s %s\n", fname, ftype)

	}
	fmt.Fprintln(b, "}")

	// constructor
	fmt.Fprintln(b)
	defineTypeConstructor(b, t)

	// implement Expr interface
	fmt.Fprintf(b, "func (self %s[T]) Accept(visitor Visitor[T]) (T, error) {\n", t.name)
	fmt.Fprintf(b, "\treturn visitor.Visit%s%s(self)\n", t.name, base)
	fmt.Fprintln(b, "}")

	fmt.Fprintln(b)
}

func defineTypeConstructor(b *strings.Builder, t typeDesc) {
	fmt.Fprintf(b, "func New%s[T any](", t.name)
	for i, f := range t.fieldList {
		if i > 0 {
			fmt.Fprint(b, ", ")
		}
		fmt.Fprintf(b, "%s %s", strings.ToLower(f.name), f.typ)
	}

	fmt.Fprintf(b, ") %s[T] {\n", t.name)
	fmt.Fprintf(b, "\treturn %s[T]{\n", t.name)

	for _, f := range t.fieldList {
		fmt.Fprintf(b, "\t\t%s: %s,\n", f.name, strings.ToLower(f.name))
	}

	fmt.Fprintln(b, "\t}")
	fmt.Fprintln(b, "}")
	fmt.Fprintln(b)
}
