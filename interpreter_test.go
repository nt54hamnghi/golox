package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func parseProgramForTest(t *testing.T, source string) []Stmt {
	t.Helper()
	r := require.New(t)

	scanner := NewScanner(source)
	tokens, err := scanner.ScanTokens()
	r.NoError(err)

	parser := NewParser(tokens)
	return parser.Parse()
}

func interpretProgramForTest(t *testing.T, source string) (Interpreter, error) {
	t.Helper()

	globals = NewEnvironment()
	interpreter := NewInterpreter()
	err := interpreter.Interpret(parseProgramForTest(t, source))
	return interpreter, err
}

func assertGlobalValues(t *testing.T, got map[string]Object, want map[string]Object) {
	t.Helper()
	r := require.New(t)

	r.Contains(got, "clock")
	for name, value := range want {
		r.Equal(value, got[name])
	}
	r.Len(got, len(want)+1)
}

func TestInterpreterExpressionStatementsSuccess(t *testing.T) {
	r := require.New(t)
	source := `
(53 + 11 - 22) > (52 - 53) * 2;
"hello" + "quz" + "bar" == "helloquzbar";
24 - 76 >= -73 * 2 / 73 + 68;
false == false;
`

	interpreter, err := interpretProgramForTest(t, source)

	r.NoError(err)
	assertGlobalValues(t, interpreter.environment.values, map[string]Object{})
}

func TestInterpreterVariableDeclarationsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   map[string]Object
	}{
		{
			name: "single variable declaration",
			source: `
var foo = 10;
`,
			want: map[string]Object{"foo": float64(10)},
		},
		{
			name: "multiple variables and arithmetic",
			source: `
var bar = 20;
var hello = 20;
var foo = 20;
var total = bar + hello + foo;
`,
			want: map[string]Object{
				"bar":   float64(20),
				"hello": float64(20),
				"foo":   float64(20),
				"total": float64(60),
			},
		},
		{
			name: "variable initialized from another variable",
			source: `
var hello = 43;
var quz = hello;
`,
			want: map[string]Object{
				"hello": float64(43),
				"quz":   float64(43),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			interpreter, err := interpretProgramForTest(t, tt.source)

			r.NoError(err)
			assertGlobalValues(t, interpreter.environment.values, tt.want)
		})
	}
}

func TestInterpreterVariableRuntimeErrors(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name: "undefined variable in expression statement",
			source: `
var foo = 88;
bar;
`,
			wantErr: "Undefined variable 'bar'.\n[line 3]",
		},
		{
			name: "undefined variable in initializer",
			source: `
var foo = 94;
var result = (foo + quz) / hello;
`,
			wantErr: "Undefined variable 'quz'.\n[line 3]",
		},
		{
			name: "initializer references undeclared variable",
			source: `
var hello = baz;
`,
			wantErr: "Undefined variable 'baz'.\n[line 2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			_, err := interpretProgramForTest(t, tt.source)

			r.Error(err)
			r.Equal(tt.wantErr, err.Error())
		})
	}
}

func TestInterpreterVariableInitializationSuccess(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   map[string]Object
	}{
		{
			name: "uninitialized variable defaults to nil",
			source: `
var foo;
`,
			want: map[string]Object{"foo": nil},
		},
		{
			name: "mixed initialized and uninitialized variables",
			source: `
var world = 22;
var hello;
var quz;
`,
			want: map[string]Object{
				"world": float64(22),
				"hello": nil,
				"quz":   nil,
			},
		},
		{
			name: "computed value with trailing nil variable",
			source: `
var hello = 35 + 63 * 58;
var quz = 63 * 58;
var bar;
`,
			want: map[string]Object{
				"hello": float64(3689),
				"quz":   float64(3654),
				"bar":   nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			interpreter, err := interpretProgramForTest(t, tt.source)

			r.NoError(err)
			assertGlobalValues(t, interpreter.environment.values, tt.want)
		})
	}
}

func TestInterpreterVariableRedeclarationsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   map[string]Object
	}{
		{
			name: "redeclare replaces prior value",
			source: `
var bar = "before";
var bar = "after";
`,
			want: map[string]Object{"bar": "after"},
		},
		{
			name: "redeclare with current variable value",
			source: `
var baz = "after";
var baz = "before";
var baz = baz;
`,
			want: map[string]Object{"baz": "before"},
		},
		{
			name: "redeclare after unrelated variable",
			source: `
var quz = 2;
var quz = 3;
var hello = 5;
var quz = hello;
`,
			want: map[string]Object{
				"quz":   float64(5),
				"hello": float64(5),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			interpreter, err := interpretProgramForTest(t, tt.source)

			r.NoError(err)
			assertGlobalValues(t, interpreter.environment.values, tt.want)
		})
	}
}

func TestInterpreterAssignmentsSuccess(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   map[string]Object
	}{
		{
			name: "assignment updates variable and returns assigned value",
			source: `
var foo;
foo = 1;
var result = foo = 2;
`,
			want: map[string]Object{
				"foo":    float64(2),
				"result": float64(2),
			},
		},
		{
			name: "assignment across declared variables",
			source: `
var baz = 89;
var world = 89;
world = baz;
baz = world;
`,
			want: map[string]Object{
				"baz":   float64(89),
				"world": float64(89),
			},
		},
		{
			name: "right associative chained assignment",
			source: `
var baz;
var quz;
baz = quz = 69 + 36 * 37;
`,
			want: map[string]Object{
				"baz": float64(1401),
				"quz": float64(1401),
			},
		},
		{
			name: "chained assignment from computed value",
			source: `
var foo = 82;
var bar;
var baz;
foo = bar = baz = foo * 2;
`,
			want: map[string]Object{
				"foo": float64(164),
				"bar": float64(164),
				"baz": float64(164),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			interpreter, err := interpretProgramForTest(t, tt.source)

			r.NoError(err)
			assertGlobalValues(t, interpreter.environment.values, tt.want)
		})
	}
}

func TestInterpreterBlockScopesSuccess(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		wantGlobals    map[string]Object
		missingGlobals []string
	}{
		{
			name: "block local variable does not leak",
			source: `
{
	var baz = "world";
}
`,
			wantGlobals:    map[string]Object{},
			missingGlobals: []string{"baz"},
		},
		{
			name: "nested block can read outer variable",
			source: `
var quz = 28;
{
	var hello = 28;
	var copy = quz;
}
`,
			wantGlobals: map[string]Object{
				"quz": float64(28),
			},
			missingGlobals: []string{"hello", "copy"},
		},
		{
			name: "inner scope shadows outer variable without mutating it",
			source: `
var world = "before";
{
	var world = "after";
}
`,
			wantGlobals: map[string]Object{
				"world": "before",
			},
			missingGlobals: []string{},
		},
		{
			name: "inner scope assignment updates outer variable",
			source: `
var world = "outer world";
var quz = "outer quz";
{
	world = "modified world";
	var quz = "inner quz";
}
`,
			wantGlobals: map[string]Object{
				"world": "modified world",
				"quz":   "outer quz",
			},
			missingGlobals: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			interpreter, err := interpretProgramForTest(t, tt.source)

			r.NoError(err)
			assertGlobalValues(t, interpreter.environment.values, tt.wantGlobals)
			for _, name := range tt.missingGlobals {
				_, exists := interpreter.environment.values[name]
				r.False(exists, "expected %q to remain local to a block", name)
			}
		})
	}
}
