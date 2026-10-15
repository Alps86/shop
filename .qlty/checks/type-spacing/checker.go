package main

import (
	"go/ast"
)

func (check checker) violations() []int {
	var lines []int
	var previous ast.Decl
	for _, declaration := range check.file.Decls {
		current := declaration
		lines = check.appendViolation(lines, previous, current)

		previous = current
	}

	return lines
}

func (check checker) start(declaration ast.Decl) int {
	start := check.positions.Position(declaration.Pos()).Line
	if doc := check.documentation(declaration); doc != nil {
		start = check.positions.Position(doc.Pos()).Line
	}

	return start
}

func (check checker) end(declaration ast.Decl) int {
	end := check.positions.Position(declaration.End()).Line
	for _, comment := range check.file.Comments {
		end = check.trailingCommentEnd(comment, end)
	}

	return end
}

func (check checker) appendViolation(lines []int, previous, current ast.Decl) []int {
	if current == nil {
		return lines
	}

	if previous == nil {
		return lines
	}

	if check.start(current)-check.end(previous) == 2 {
		return lines
	}

	return append(lines, check.positions.Position(current.Pos()).Line)
}

func (check checker) documentation(declaration ast.Decl) *ast.CommentGroup {
	if function, ok := declaration.(*ast.FuncDecl); ok {
		return function.Doc
	}

	return declaration.(*ast.GenDecl).Doc
}

func (check checker) trailingCommentEnd(comment *ast.CommentGroup, end int) int {
	if check.positions.Position(comment.Pos()).Line != end {
		return end
	}

	return check.positions.Position(comment.End()).Line
}
