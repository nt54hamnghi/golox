package main

import (
	"fmt"
	"iter"
	"log"
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

	err := defineAst(outputDir, "Expr", []typeDesc{
		{"Literal", []field{{"Value", "any"}}},
		{"Grouping", []field{{"Expression", "Expr"}}},
		{"Unary", []field{
			{"Operator", "Token"},
			{"Right", "Expr"},
		}},
		{"Variable", []field{{"Name", "Token"}}},
		{"Assignment", []field{
			{"Name", "Token"},
			{"Value", "Expr"},
		}},
		{"Binary", []field{
			{"Left", "Expr"},
			{"Operator", "Token"},
			{"Right", "Expr"},
		}},
		{"Logical", []field{
			{"Left", "Expr"},
			{"Operator", "Token"},
			{"Right", "Expr"},
		}},
		{"Call", []field{
			{"Callee", "Expr"},
			{"Paren", "Token"},
			{"Arguments", "[]Expr"},
		}},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = defineAst(outputDir, "Stmt", []typeDesc{
		{"Expression", []field{{"Expression", "Expr"}}},
		{"Print", []field{{"Expression", "Expr"}}},
		{"Var", []field{
			{"Name", "Token"},
			{"Initializer", "Expr"},
		}},
		{"Function", []field{
			{"Name", "Token"},
			{"Params", "[]Token"},
			{"Body", "[]Stmt"},
		}},
		{"If", []field{
			{"Condition", "Expr"},
			{"ThenBranch", "Stmt"},
			{"ElseBranch", "Stmt"},
		}},
		{"While", []field{
			{"Condition", "Expr"},
			{"Body", "Stmt"},
		}},
		{"Return", []field{
			{"Keyword", "Token"},
			{"Value", "Expr"},
		}},
		{"Block", []field{{"Stmts", "[]Stmt"}}},
	})
	if err != nil {
		log.Fatal(err)
	}

	if err = defineNodeIdGo(outputDir); err != nil {
		log.Fatal(err)
	}
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

func defineNodeIdGo(outputDir string) error {
	b := strings.Builder{}

	fmt.Fprintln(&b, "package main")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "import (")
	fmt.Fprintln(&b, "\t\"bytes\"")
	fmt.Fprintln(&b, "\t\"encoding/gob\"")
	fmt.Fprintln(&b, "\t\"hash/maphash\"")
	fmt.Fprintln(&b, "\t\"sync/atomic\"")
	fmt.Fprintln(&b, ")")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "var root atomic.Uint64")
	fmt.Fprintln(&b, "var seed maphash.Seed = maphash.MakeSeed()")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "type NodeID struct {")
	fmt.Fprintln(&b, "\tid     uint64")
	fmt.Fprintln(&b, "\tdigest uint64")
	fmt.Fprintln(&b, "}")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "func NewNodeIDFrom(v any) NodeID {")
	fmt.Fprintln(&b, "\tid := root.Add(1)")
	fmt.Fprintln(&b, "\treturn NodeID{id: id, digest: nodeDigest(id, v)}")
	fmt.Fprintln(&b, "}")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "func nodeDigest(id uint64, v any) uint64 {")
	fmt.Fprintln(&b, "\tvar buf bytes.Buffer")
	fmt.Fprintln(&b, "\tif err := gob.NewEncoder(&buf).Encode(v); err != nil {")
	fmt.Fprintln(&b, "\t\tpanic(err)")
	fmt.Fprintln(&b, "\t}")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "\tvar h maphash.Hash")
	fmt.Fprintln(&b, "\th.SetSeed(seed)")
	fmt.Fprintln(&b, "\tmaphash.WriteComparable(&h, id)")
	fmt.Fprintln(&b, "\tmaphash.WriteComparable(&h, string(buf.Bytes()))")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "\treturn h.Sum64()")
	fmt.Fprintln(&b, "}")

	path := filepath.Join(outputDir, strings.ToLower("nodeid")+".go")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		return err
	}

	return nil
}

func defineAst(outputDir string, base string, types []typeDesc) error {
	b := strings.Builder{}

	fmt.Fprintln(&b, "package main")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "import (")
	fmt.Fprintln(&b, "\t\"encoding/gob\"")
	fmt.Fprintln(&b, "\t\"fmt\"")
	fmt.Fprintln(&b, ")")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "type %s interface {\n", base)
	fmt.Fprintf(&b, "\tAccept(visitor %sVisitor) (any, error)\n", base)
	fmt.Fprint(&b, "\tId() NodeID\n")
	fmt.Fprintln(&b, "}")
	fmt.Fprintln(&b)
	defineGobRegistration(&b, types)

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

func defineGobRegistration(b *strings.Builder, types []typeDesc) {
	fmt.Fprintln(b, "func init() {")
	for _, t := range types {
		fmt.Fprintf(b, "\tgob.Register(%s{})\n", t.name)
	}
	fmt.Fprintln(b, "}")
	fmt.Fprintln(b)
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
	fmt.Fprintf(b, "\tid %s\n", "NodeID")
	fmt.Fprintln(b, "}")

	// constructor
	fmt.Fprintln(b)
	defineTypeConstructor(b, t)

	// implement interface
	fmt.Fprintf(b, "func (self %s) Accept(visitor %sVisitor) (any, error) {\n", t.name, base)
	fmt.Fprintf(b, "\treturn visitor.Visit%s%s(self)\n", t.name, base)
	fmt.Fprintln(b, "}")

	fmt.Fprintln(b)
	fmt.Fprintf(b, "func (self %s) Id() NodeID {\n", t.name)
	makeTempDataWithSource(b, t, "self")
	fmt.Fprintln(b, "\tif nodeDigest(self.id.id, tmp) != self.id.digest {")
	fmt.Fprintf(b, "\t\tpanic(fmt.Sprintf(\"node id hash mismatch, a copied value was modified: %%#v\", self))\n")

	fmt.Fprintln(b, "\t}")
	fmt.Fprintf(b, "\treturn self.id\n")
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
	fmt.Fprintf(b, "\tnode := %s{\n", t.name)

	for _, f := range t.fieldList {
		fmt.Fprintf(b, "\t\t%s: %s,\n", f.name, strings.ToLower(f.name))
	}

	fmt.Fprintln(b, "\t}")
	fmt.Fprintln(b)
	makeTempData(b, t)
	fmt.Fprintln(b, "\tnode.id = NewNodeIDFrom(tmp)")
	fmt.Fprintln(b, "\treturn node")
	fmt.Fprintln(b, "}")
	fmt.Fprintln(b)
}

func makeTempData(b *strings.Builder, t typeDesc) {
	makeTempDataWithSource(b, t, "node")
}

func makeTempDataWithSource(b *strings.Builder, t typeDesc, source string) {
	fmt.Fprint(b, "\ttmp := struct{ ")
	for i, f := range t.fieldList {
		if i > 0 {
			fmt.Fprint(b, " ")
		}
		fmt.Fprintf(b, "%s %s;", f.name, f.typ)
	}
	fmt.Fprint(b, " }{")
	for i, f := range t.fieldList {
		if i > 0 {
			fmt.Fprint(b, ", ")
		}
		fmt.Fprintf(b, "%s: %s.%s", f.name, source, f.name)
	}
	fmt.Fprintln(b, "}")
}
