package llm

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// newClientPattern matches constructor function names like newFooClient.
var newClientPattern = regexp.MustCompile(`^new[A-Z][A-Za-z]*Client$`)

// TestAllLLMConstructorsRegistered verifies that every new*Client constructor
// function defined in this package is called from NewClient, so new LLM
// providers cannot be added as dead code.  A failing test means a constructor
// exists but is not wired into the factory switch in llm.go.
func TestAllLLMConstructorsRegistered(t *testing.T) {
	fset := token.NewFileSet()

	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob *.go: %v", err)
	}

	var parsed []*ast.File
	for _, name := range matches {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, parseErr := parser.ParseFile(fset, name, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", name, parseErr)
		}
		parsed = append(parsed, f)
	}

	// Collect all new*Client constructor names defined in the package.
	constructors := map[string]bool{} // value: true = seen in NewClient
	for _, f := range parsed {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}
			if newClientPattern.MatchString(fn.Name.Name) {
				constructors[fn.Name.Name] = false
			}
		}
	}

	// Mark each constructor that is called inside NewClient.
	for _, f := range parsed {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "NewClient" {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				if ident, ok := call.Fun.(*ast.Ident); ok {
					if _, exists := constructors[ident.Name]; exists {
						constructors[ident.Name] = true
					}
				}
				return true
			})
		}
	}

	for name, registered := range constructors {
		if !registered {
			t.Errorf("constructor %q is defined but never called from NewClient; "+
				"add a case for it in the switch statement in llm.go", name)
		}
	}
}
