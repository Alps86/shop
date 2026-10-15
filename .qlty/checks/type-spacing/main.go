package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	(runner{}).exitOnError((runner{}).run(os.Args[1:]))
}

func (runner runner) run(arguments []string) error {
	if runner.isRegression(arguments) {
		return runner.regression(arguments[1])
	}

	for _, path := range arguments {
		runner.exitOnError(runner.check(path))
	}

	return nil
}

func (runner runner) isRegression(arguments []string) bool {
	if len(arguments) != 2 {
		return false
	}

	return arguments[0] == "--regression"
}

func (runner runner) check(path string) error {
	if path == "--" {
		return nil
	}

	positions := token.NewFileSet()
	file, err := parser.ParseFile(positions, path, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	check := checker{positions: positions, file: file}
	for _, line := range check.violations() {
		fmt.Printf("%s:%d:1: shop-go-type-declaration-spacing Zwischen eigenständigen Top-Level-Deklarationen muss genau eine Leerzeile stehen.\n", path, line)
	}

	return nil
}

func (runner runner) exitOnError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
