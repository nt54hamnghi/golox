package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

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
		runFile(args[0])
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

	run(string(bytes))

	if hadError {
		os.Exit(65)
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
		run(line)
		hadError = false
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func run(src string) {
	scanner := NewScanner(src)
	tokens := scanner.scanTokens()

	for _, token := range tokens {
		fmt.Println(token)
	}
}

func err(line int, message string) {
	report(line, "", message)
}

func report(line int, where string, message string) {
	fmt.Fprintf(os.Stderr, "[line %d ] Error%s: %s\n", line, where, message)
	hadError = true
}
