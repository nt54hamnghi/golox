package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	internalErrors "github.com/nt54hamnghi/golox/internal/errors"
	"github.com/nt54hamnghi/golox/internal/interpreter"
	"github.com/nt54hamnghi/golox/internal/parser"
	"github.com/nt54hamnghi/golox/internal/resolver"
	"github.com/nt54hamnghi/golox/internal/scanner"
)

var in interpreter.Interpreter = interpreter.NewInterpreter()

func main() {
	args := os.Args
	if len(args) < 1 {
		panic("not enough arguments")
	}

	args = args[1:]

	if len(args) > 1 {
		fmt.Println("Usage: glox [script]")
		os.Exit(64)
	} else if len(args) == 1 {
		if err := runFile(args[0]); err != nil {
			panic(err)
		}
	} else {
		runPrompt()
	}
}

var hadError bool

// Reads the file path and executes its content.
func runFile(path string) error {

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	bytes, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	err = run(string(bytes))
	if err != nil {
		exit(err)
	}

	return nil
}

// Execute in interactive mode (REPL)
func runPrompt() error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if err := run(line); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func run(src string) error {
	sc := scanner.NewScanner(src)
	tokens, err := sc.ScanTokens()
	if err != nil {
		return err
	}

	pa := parser.NewParser(tokens)
	prog := pa.Parse()

	re := resolver.NewResolver(&in)
	if _, err := re.Resolve(prog); err != nil {
		return err
	}

	err = in.Interpret(prog)
	if err != nil {
		return err
	}

	// var printer AstPrinter
	// repr := printer.String(expr)
	// fmt.Println(repr)

	return nil
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, err)

	var runtimeErr internalErrors.RuntimeError
	if errors.As(err, &runtimeErr) {
		os.Exit(70)
	}

	os.Exit(65)
}
