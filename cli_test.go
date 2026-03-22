package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

func buildCLIForTest(t *testing.T) string {
	t.Helper()
	r := require.New(t)

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "golox")
	cacheDir := filepath.Join(tmpDir, "gocache")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(), "GOCACHE="+cacheDir)

	output, err := cmd.CombinedOutput()
	r.NoError(err, string(output))

	return binaryPath
}

func runCLIForTest(t *testing.T, binaryPath string, source string) cliResult {
	t.Helper()
	r := require.New(t)

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test.lox")
	r.NoError(os.WriteFile(scriptPath, []byte(source), 0o644))

	cmd := exec.Command(binaryPath, scriptPath)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		r.True(errors.As(err, &exitErr), err.Error())
		exitCode = exitErr.ExitCode()
	}

	return cliResult{
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		exitCode: exitCode,
	}
}

func TestCLIPrintStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "print single expressions",
			source: `
print true;
print "bar" + "quz" + "world";
print (45 * 2 + 58 * 2) / (2);
`,
			wantStdout: "true\nbarquzworld\n103\n",
		},
		{
			name: "multiple print statements with multiline string",
			source: `
print true != false;
print "98
99
14
";
print "There should be an empty line above this.";
print "(" + "" + ")";
print "non-ascii: ॐ";
`,
			wantStdout: "true\n98\n99\n14\n\nThere should be an empty line above this.\n()\nnon-ascii: ॐ\n",
		},
	}

	r := require.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLIRuntimeErrorsReportStderrAndExit70(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
		wantStderr string
	}{
		{
			name: "undefined variable after previous successful print",
			source: `
print 38;
print x;
`,
			wantStdout: "38\n",
			wantStderr: "Undefined variable 'x'.\n[line 3]\n",
		},
		{
			name: "out of scope variable after nested block completes",
			source: `
{
  var world = "outer world";
  var quz = "outer quz";
  {
    world = "modified world";
    var quz = "inner quz";
    print world;
    print quz;
  }
  print world;
  print quz;
}
print world;
`,
			wantStdout: "modified world\ninner quz\nmodified world\nouter quz\n",
			wantStderr: "Undefined variable 'world'.\n[line 14]\n",
		},
	}

	r := require.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(70, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Equal(tt.wantStderr, result.stderr)
		})
	}
}

func TestCLIParseErrorsReflectCurrentBehavior(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStderr string
	}{
		{
			name: "print without expression writes parser error and exits zero",
			source: `
print;
`,
			wantStderr: "[line 2] Error at ';': Expect expression.\n",
		},
		{
			name: "missing closing brace writes parser error and exits zero",
			source: `
{
    var world = 73;
    var hello = 73;
    {
        print world + hello;
    // Missing closing curly brace
}
`,
			wantStderr: "[line 9] Error at end: Expect '}' after block.\n",
		},
	}

	r := require.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Empty(result.stdout)
			r.Equal(tt.wantStderr, result.stderr)
		})
	}
}

func TestCLIIfStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "skip false single-line body",
			source: `
// This should print the string if the condition
// evaluates to True
if (false) print "bar";
`,
			wantStdout: "",
		},
		{
			name: "skip false block body",
			source: `
// This should print "block body" if the condition
// evaluates to True
if (false) {
  print "block body";
}
`,
			wantStdout: "",
		},
		{
			name: "assignment in condition returns assigned value",
			source: `
// This program tests whether the assignment
// operation returns the value assigned.
// The if condition should evaluate to true and
// the inner boolean expression must be printed.
// So, in this case the if condition evaluates to
//true and prints the inner boolean expression
var a = false;
if (a = true) {
  print (a == false);
}
`,
			wantStdout: "false\n",
		},
		{
			name: "multiple if statements update stage and voting eligibility",
			source: `
// This program should print a different string
// based on the value of age
var stage = "unknown";
var age = 58;
if (age < 18) { stage = "child"; }
if (age >= 18) { stage = "adult"; }
print stage;

var isAdult = age >= 18;
if (isAdult) { print "eligible for voting"; }
if (!isAdult) { print "not eligible for voting"; }
`,
			wantStdout: "adult\neligible for voting\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLIElseStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "single-line else branch runs",
			source: `
// This program uses a random boolean to decide
// which branch to execute,
// and then prints the appropriate string
if (false) print "if"; else print "else";
`,
			wantStdout: "else\n",
		},
		{
			name: "age check chooses adult branch",
			source: `
// This program initializes age with a random
// integer and then prints "adult"
// if the age is greater than 18, otherwise it
// prints "child"
var age = 77;
if (age > 18) print "adult"; else print "child";
`,
			wantStdout: "adult\n",
		},
		{
			name: "mix block and single-line branches",
			source: `
// This program uses a random boolean to decide
// which branch to execute,
// and then prints the appropriate string
if (true) {
  print "if block";
} else print "else statement";

if (true) print "if statement"; else {
  print "else block";
}
`,
			wantStdout: "if block\nif statement\n",
		},
		{
			name: "temperature branch prints cold path",
			source: `
// This program converts a random integer from
// Celsius to Fahrenheit
// and prints the result. It also prints a message
// based on the temperature.
var celsius = 22;
var fahrenheit = 0;
var isHot = false;

{
  fahrenheit = celsius * 9 / 5 + 32;
  print celsius; print fahrenheit;

  if (celsius > 30) {
    isHot = true;
    print "It's a hot day. Stay hydrated!";
  } else {
    print "It's cold today. Wear a jacket!";
  }

  if (isHot) { print "Use sunscreen!"; }
}
`,
			wantStdout: "22\n71.6\nIt's cold today. Wear a jacket!\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLIElseIfStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "if branch wins over else-if branch",
			source: `
// This program uses a random boolean to decide
// which branch to execute,
// and then prints the appropriate string
if (true) print "if branch";
else if (true) print "else-if branch";
`,
			wantStdout: "if branch\n",
		},
		{
			name: "all false branches print nothing",
			source: `
// This program uses a random boolean to decide
// which branch to execute,
// and then prints the appropriate string
if (false) {
  print "bar";
} else if (false) print "bar";

if (false) print "bar"; else if (false) {
  print "bar";
}
`,
			wantStdout: "",
		},
		{
			name: "else-if assigns adult stage",
			source: `
// This program uses multiple if statements to
// categorize a person
// into different life stages based on their age
var age = 28;
var stage = "unknown";
if (age < 18) { stage = "child"; }
else if (age >= 18) { stage = "adult"; }
else if (age >= 65) { stage = "senior"; }
else if (age >= 100) { stage = "centenarian"; }
print stage;
`,
			wantStdout: "adult\n",
		},
		{
			name: "else-if chain for life permissions",
			source: `
// This program uses multiple if statements to
// determine eligibility for
// voting, driving, and drinking based on a random
// integer age
var age = 40;

var isAdult = age >= 18;
if (isAdult) { print "eligible for voting: true"; }
else { print "eligible for voting: false"; }

if (age < 16) { print "not eligible for driving"; }
else if (age < 18) { print "learner's permit"; }
else { print "eligible for driving"; }

if (age >= 21) { print "eligible for drinking"; }
else { print "not eligible for drinking"; }
`,
			wantStdout: "eligible for voting: true\neligible for driving\neligible for drinking\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLINestedIfStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "nested true branches print once",
			source: `
// This program uses nested if statements to print
// a message
if (true) if (true) print "nested true";
`,
			wantStdout: "nested true\n",
		},
		{
			name: "nested else attaches to inner if",
			source: `
// This program uses nested if statements to print
// a message
if (true) {
  if (true) print "baz"; else print "baz";
}
`,
			wantStdout: "baz\n",
		},
		{
			name: "nested conditions classify adult permissions",
			source: `
// This program categorizes a person into
// different life stages based on their age
// Then based on the age, it prints a message
// about the person's eligibility for voting,
// driving, and drinking
var stage = "unknown";
var age = 57;
if (age < 18) {
    if (age < 13) { stage = "child"; }
    else if (age < 16) {
        stage = "young teenager";
    }
    else { stage = "teenager"; }
}
else if (age < 65) {
    if (age < 30) { stage = "young adult"; }
    else if (age < 50) { stage = "adult"; }
    else { stage = "middle-aged adult"; }
}
else { stage = "senior"; }
print stage;

var isAdult = age >= 18;
if (isAdult) {
    print "eligible for voting: true";
    if (age < 25) {
        print "first-time voter: likely";
    }
    else { print "first-time voter: unlikely"; }
}
else { print "eligible for voting: false"; }

if (age < 16) { print "not eligible for driving"; }
else if (age < 18) {
    print "eligible for driving: learner's permit";
    if (age < 17) {
        print "supervised driving required";
    }
    else {
        print "driving allowed with restrictions";
    }
}
else { print "eligible for driving"; }

if (age < 21) {
    print "not eligible for drinking";
}
else {
    print "eligible for drinking";
    if (age < 25) {
        print "remember: drink responsibly!";
    }
}
`,
			wantStdout: "middle-aged adult\neligible for voting: true\nfirst-time voter: unlikely\neligible for driving\neligible for drinking\n",
		},
		{
			name: "dangling else binds to inner if",
			source: `
// This program uses nested if statements to print
// a message
if (true) if (false) print "bar";
else print "foo";
`,
			wantStdout: "foo\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLILogicalOrOperatorSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "or in if conditions uses first truthy value",
			source: `
// The logical OR operator should return the first
// value that is truthy
if (false or "ok") print "baz";
if (nil or "ok") print "baz";

if (false or false) print "bar";
if (true or "bar") print "bar";

if (57 or "hello") print "hello";
if ("hello" or "hello") print "hello";
`,
			wantStdout: "baz\nbaz\nbar\nhello\nhello\n",
		},
		{
			name: "or expressions print first truthy or final falsy value",
			source: `
// This program uses the logical OR operator to
// print the first value that is truthy
print 15 or true;
print false or 15;
print false or false or true;

print false or false;
print false or false or false;
print false or false or false or false;
`,
			wantStdout: "15\n15\ntrue\nfalse\nfalse\nfalse\n",
		},
		{
			name: "or short-circuits assignment chain",
			source: `
// This program relies on the fact that
// assignments return the assigned value
// And that the logical OR operator short-circuits
// So, if the first assignment is truthy, it
// wouldn't proceed to the subsequent assignments
// And then prints the assigned values
var a = "bar";
var b = "bar";
(a = false) or (b = true) or (a = "bar");
print a;
print b;
`,
			wantStdout: "false\ntrue\n",
		},
		{
			name: "or stage example preserves adult output",
			source: `
// This program uses if conditions to get the stage
// of a person's life based on their age, and then
// prints if they are eligible for voting
var stage = "unknown";
var age = 64;
if (age < 18) { stage = "child"; }
if (age >= 18) { stage = "adult"; }
print stage;

var isAdult = age >= 18;
if (isAdult) { print "eligible for voting"; }
if (!isAdult) { print "not eligible for voting"; }
`,
			wantStdout: "adult\neligible for voting\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLILogicalAndOperatorSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "and in if conditions uses first falsy or final truthy value",
			source: `
// The logical AND operator should return the
// first falsy value
if (false and "bad") print "hello";
if (nil and "bad") print "hello";

// If all values are truthy, it returns the last
// value
if (true and "quz") print "quz";
if (91 and "world") print "world";
if ("world" and "world") print "world";
if ("" and "bar") print "bar";
`,
			wantStdout: "quz\nworld\nworld\nbar\n",
		},
		{
			name: "and expressions print first falsy or final truthy value",
			source: `
// This program uses the logical AND operator to
// print the first falsy value
// Or the last value if all values are truthy
print false and 1;
print true and 1;
print 47 and "quz" and false;

print 47 and true;
print 47 and "quz" and 47;
`,
			wantStdout: "false\n1\nfalse\ntrue\n47\n",
		},
		{
			name: "and short-circuits after falsy assignment",
			source: `
// This program relies on the fact that
// assignments return the assigned value
// And that the logical AND operator short-circuits
// So, when it encounters a falsy value, it
// wouldn't proceed to the subsequent assignments
// And then prints the assigned values
var a = "quz";
var b = "quz";
(a = true) and (b = false) and (a = "bad");
print a;
print b;
`,
			wantStdout: "true\nfalse\n",
		},
		{
			name: "and stage example preserves adult output",
			source: `
// This program uses if conditions to get the stage
// of a person's life based on their age, and then
// prints if they are eligible for voting
var stage = "unknown";
var age = 83;
if (age < 18) { stage = "child"; }
if (age >= 18) { stage = "adult"; }
print stage;

var isAdult = age >= 18;
if (isAdult) { print "eligible for voting"; }
if (!isAdult) { print "not eligible for voting"; }
`,
			wantStdout: "adult\neligible for voting\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLIWhileStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "while loop prints incrementing assignment values",
			source: `
// This program uses a while loop to print the
// numbers from 0 to N
// The assignment operation returns the assigned
// value
var foo = 0;
while (foo < 3) print foo = foo + 1;
`,
			wantStdout: "1\n2\n3\n",
		},
		{
			name: "while loop block prints zero through two",
			source: `
// This program uses a while loop to print the
// numbers from 0 to 3
// The statement inside the block is executed
// every time the loop condition is true
var hello = 0;
while (hello < 3) {
  print hello;
  hello = hello + 1;
}
`,
			wantStdout: "0\n1\n2\n",
		},
		{
			name: "while loop computes factorial after skipped false loop",
			source: `
// This program uses a while loop to calculate the
// factorial of 5
// The first while loop never runs because the
// condition is false
while (false) { print "should not print"; }

var product = 1;
var i = 1;

while (i <= 5) {
  product = product * i;
  i = i + 1;
}

print "Factorial of 5: "; print product;
`,
			wantStdout: "Factorial of 5: \n120\n",
		},
		{
			name: "while loop prints first ten fibonacci numbers",
			source: `
// This program uses a while loop to generate and
// print the first N Fibonacci numbers
var n = 10;
var fm = 0;
var fn = 1;
var index = 0;

while (index < n) {
    print fm;
    var temp = fm;
    fm = fn;
    fn = temp + fn;
    index = index + 1;
}
`,
			wantStdout: "0\n1\n1\n2\n3\n5\n8\n13\n21\n34\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}

func TestCLIForStatementsSuccess(t *testing.T) {
	binaryPath := buildCLIForTest(t)

	tests := []struct {
		name       string
		source     string
		wantStdout string
	}{
		{
			name: "for loop without increment clause prints incrementing assignment values",
			source: `
// This program uses a for loop to print the
// numbers from 0 to 3
// The assignment operation returns the assigned
// value
for (var foo = 0; foo < 3;) print foo = foo + 1;
`,
			wantStdout: "1\n2\n3\n",
		},
		{
			name: "for loop block prints zero through two",
			source: `
// This program uses a for loop to print the
// numbers from 0 to 3
for (var baz = 0; baz < 3; baz = baz + 1) {
  print baz;
}
`,
			wantStdout: "0\n1\n2\n",
		},
		{
			name: "for loop ignores missing initializer and increment clauses as logged",
			source: `
// This program uses a for loop to print the
// numbers from 0 to 2
// The loop initializer is ignored in this loop
var hello = 0;
for (; hello < 2; hello = hello + 1) print hello;

// This program uses a for loop to print the
// numbers from 0 to 2
// The loop increment clause is ignored in this
// loop
for (var quz = 0; quz < 2;) {
  print quz;
  quz = quz + 1;
}
`,
			wantStdout: "0\n1\n0\n1\n",
		},
		{
			name: "for loop scopes shadowing variables as logged",
			source: `
// This program uses for loops and block scopes
// to print the updates to the same variable
var baz = "after";
{
  var baz = "before";

  for (var baz = 0; baz < 1; baz = baz + 1) {
    print baz;
    var baz = -1;
    print baz;
  }
}

{
  for (var baz = 0; baz > 0; baz = baz + 1) {}

  var baz = "after";
  print baz;

  for (baz = 0; baz < 1; baz = baz + 1) {
    print baz;
  }
}
`,
			wantStdout: "0\n-1\nafter\n0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			result := runCLIForTest(t, binaryPath, tt.source)

			r.Equal(0, result.exitCode)
			r.Equal(tt.wantStdout, result.stdout)
			r.Empty(result.stderr)
		})
	}
}
