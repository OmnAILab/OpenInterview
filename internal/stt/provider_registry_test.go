package stt

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// factoryStructPattern matches struct type names like xxxFactory (one or more letters before "Factory").
var factoryStructPattern = regexp.MustCompile(`[A-Za-z]+Factory$`)

// TestNewFactory_SelectsProviderImplementation verifies that NewFactory returns
// the correct concrete type for each supported provider string.
func TestNewFactory_SelectsProviderImplementation(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	if _, ok := NewFactory(Config{Provider: "mock"}, logger).(*mockFactory); !ok {
		t.Fatal("mock provider should return *mockFactory")
	}

	sherpa := NewFactory(Config{
		Provider: "sherpa",
		Sherpa:   &SherpaConfig{WSURL: "ws://test"},
	}, logger)
	if _, ok := sherpa.(*sherpaWebSocketFactory); !ok {
		t.Fatal("sherpa provider should return *sherpaWebSocketFactory")
	}

	sherpaAliases := []string{"sherpa-websocket", "sherpa_onnx", "sherpa-onnx"}
	for _, alias := range sherpaAliases {
		f := NewFactory(Config{Provider: alias, Sherpa: &SherpaConfig{WSURL: "ws://test"}}, logger)
		if _, ok := f.(*sherpaWebSocketFactory); !ok {
			t.Fatalf("provider %q should return *sherpaWebSocketFactory", alias)
		}
	}

	tencent := NewFactory(Config{
		Provider: "tencent",
		Tencent: &TencentConfig{
			AppID:     "app",
			SecretID:  "id",
			SecretKey: "key",
		},
	}, logger)
	if _, ok := tencent.(*tencentAsrFactory); !ok {
		t.Fatal("tencent provider should return *tencentAsrFactory")
	}

	tencentAlias := NewFactory(Config{
		Provider: "tencent-asr",
		Tencent: &TencentConfig{
			AppID:     "app",
			SecretID:  "id",
			SecretKey: "key",
		},
	}, logger)
	if _, ok := tencentAlias.(*tencentAsrFactory); !ok {
		t.Fatal("tencent-asr provider should return *tencentAsrFactory")
	}

	if _, ok := NewFactory(Config{Provider: "unknown"}, logger).(*unsupportedFactory); !ok {
		t.Fatal("unknown provider should return *unsupportedFactory")
	}
}

// TestAllSTTFactoriesRegistered verifies that every *Factory struct type
// defined in provider-specific files (i.e. not stt.go) is instantiated inside
// NewFactory, so new STT providers cannot be added as dead code.
func TestAllSTTFactoriesRegistered(t *testing.T) {
	fset := token.NewFileSet()

	matches, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob *.go: %v", err)
	}

	var allFiles []*ast.File
	var providerFiles []*ast.File // excludes stt.go (which holds internal helpers)
	for _, name := range matches {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, parseErr := parser.ParseFile(fset, name, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", name, parseErr)
		}
		allFiles = append(allFiles, f)
		if name != "stt.go" {
			providerFiles = append(providerFiles, f)
		}
	}

	// Collect all *Factory struct type names defined in provider files.
	factories := map[string]bool{} // value: true = seen in NewFactory
	for _, f := range providerFiles {
		for _, decl := range f.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range genDecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}
				if factoryStructPattern.MatchString(ts.Name.Name) {
					factories[ts.Name.Name] = false
				}
			}
		}
	}

	// Mark each factory struct that is instantiated inside NewFactory.
	for _, f := range allFiles {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "NewFactory" {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				unary, ok := n.(*ast.UnaryExpr)
				if !ok {
					return true
				}
				comp, ok := unary.X.(*ast.CompositeLit)
				if !ok {
					return true
				}
				if ident, ok := comp.Type.(*ast.Ident); ok {
					if _, exists := factories[ident.Name]; exists {
						factories[ident.Name] = true
					}
				}
				return true
			})
		}
	}

	for name, registered := range factories {
		if !registered {
			t.Errorf("factory struct %q is defined but never instantiated in NewFactory; "+
				"add a case for it in the switch statement in stt.go", name)
		}
	}
}
