package main

import (
	"go/ast"
	"go/token"
)

type checker struct {
	positions *token.FileSet
	file      *ast.File
}

type runner struct{}

type fixture struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	Violations int    `json:"violations"`
}
