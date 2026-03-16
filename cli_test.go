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
