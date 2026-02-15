package architecture_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePrefix = "github.com/felixgeelhaar/aios/"

// TestDDDLayerBoundaries enforces the strict dependency direction:
//
//	domain -> application -> adapters/runtime -> cmd
//
// Domain must never import application or infrastructure packages.
// Application must never import adapter, runtime, or infrastructure packages.
func TestDDDLayerBoundaries(t *testing.T) {
	repoRoot := findRepoRoot(t)
	checkNoForbiddenImports(t, filepath.Join(repoRoot, "internal", "domain"), []string{
		modulePrefix + "internal/application/",
		modulePrefix + "internal/adapters/",
		modulePrefix + "internal/agents",
		modulePrefix + "internal/core",
		modulePrefix + "internal/mcp",
		modulePrefix + "internal/runtime",
		modulePrefix + "internal/sync",
	})
	checkNoForbiddenImports(t, filepath.Join(repoRoot, "internal", "application"), []string{
		modulePrefix + "internal/adapters/",
		modulePrefix + "internal/agents",
		modulePrefix + "internal/core",
		modulePrefix + "internal/mcp",
		modulePrefix + "internal/runtime",
		modulePrefix + "internal/sync",
	})
}

// TestDomainNoIO ensures domain packages never import I/O or infrastructure
// standard library packages. Domain models must remain pure and free of
// side effects — all I/O is performed through injected port interfaces.
func TestDomainNoIO(t *testing.T) {
	repoRoot := findRepoRoot(t)
	forbiddenStdlib := []string{
		"os",
		"io",
		"io/fs",
		"net",
		"net/http",
		"net/url",
		"database/sql",
		"os/exec",
		"os/signal",
		"syscall",
	}
	checkNoExactImports(t, filepath.Join(repoRoot, "internal", "domain"), forbiddenStdlib)
}

// TestDomainNoLogging ensures domain packages never import logging packages.
// Domain logic must not log — logging is an infrastructure concern handled
// by the application or adapter layers via injected abstractions.
func TestDomainNoLogging(t *testing.T) {
	repoRoot := findRepoRoot(t)
	forbiddenLog := []string{
		"log",
		"log/slog",
	}
	checkNoExactImports(t, filepath.Join(repoRoot, "internal", "domain"), forbiddenLog)
	// Also reject third-party loggers.
	checkNoForbiddenImports(t, filepath.Join(repoRoot, "internal", "domain"), []string{
		"go.uber.org/zap",
		"github.com/sirupsen/logrus",
		"github.com/rs/zerolog",
		"github.com/felixgeelhaar/bolt",
	})
}

// TestGovernanceNoIO ensures the governance package never imports I/O standard
// library packages. After extracting WriteBundle/LoadBundle into adapter ports,
// governance must remain free of filesystem side effects.
func TestGovernanceNoIO(t *testing.T) {
	repoRoot := findRepoRoot(t)
	forbiddenStdlib := []string{
		"os",
		"io",
		"io/fs",
		"net",
		"net/http",
		"net/url",
		"database/sql",
		"os/exec",
		"os/signal",
		"syscall",
	}
	checkNoExactImports(t, filepath.Join(repoRoot, "internal", "governance"), forbiddenStdlib)
}

// TestObservabilityNoIO ensures the observability package never imports I/O
// standard library packages. After extracting AppendSnapshot/LoadSnapshots
// into adapter ports, observability must remain free of filesystem side effects.
func TestObservabilityNoIO(t *testing.T) {
	repoRoot := findRepoRoot(t)
	forbiddenStdlib := []string{
		"os",
		"io",
		"io/fs",
		"net",
		"net/http",
		"net/url",
		"database/sql",
		"os/exec",
		"os/signal",
		"syscall",
	}
	checkNoExactImports(t, filepath.Join(repoRoot, "internal", "observability"), forbiddenStdlib)
}

// TestDomainBoundedContextIsolation ensures each bounded context in the domain
// layer owns its own model and does not import from other bounded contexts.
// Cross-BC coordination must happen at the application layer.
func TestDomainBoundedContextIsolation(t *testing.T) {
	repoRoot := findRepoRoot(t)
	domainDir := filepath.Join(repoRoot, "internal", "domain")

	entries, err := os.ReadDir(domainDir)
	if err != nil {
		t.Fatalf("read domain dir: %v", err)
	}

	var contexts []string
	for _, e := range entries {
		if e.IsDir() {
			contexts = append(contexts, e.Name())
		}
	}

	for _, bc := range contexts {
		bcDir := filepath.Join(domainDir, bc)
		// Build forbidden list: all other bounded contexts.
		var forbidden []string
		for _, other := range contexts {
			if other != bc {
				forbidden = append(forbidden, modulePrefix+"internal/domain/"+other)
			}
		}
		if len(forbidden) > 0 {
			checkNoForbiddenImports(t, bcDir, forbidden)
		}
	}
}

// TestApplicationBoundedContextAlignment ensures each application service only
// imports its own bounded context's domain package, not other BCs' domain
// packages. Cross-BC orchestration requires an explicit coordination service.
func TestApplicationBoundedContextAlignment(t *testing.T) {
	repoRoot := findRepoRoot(t)
	appDir := filepath.Join(repoRoot, "internal", "application")

	entries, err := os.ReadDir(appDir)
	if err != nil {
		t.Fatalf("read application dir: %v", err)
	}

	domainEntries, err := os.ReadDir(filepath.Join(repoRoot, "internal", "domain"))
	if err != nil {
		t.Fatalf("read domain dir: %v", err)
	}
	var domainContexts []string
	for _, e := range domainEntries {
		if e.IsDir() {
			domainContexts = append(domainContexts, e.Name())
		}
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bc := e.Name()
		bcDir := filepath.Join(appDir, bc)
		// Forbid imports of domain BCs other than this service's own.
		var forbidden []string
		for _, domBC := range domainContexts {
			if domBC != bc {
				forbidden = append(forbidden, modulePrefix+"internal/domain/"+domBC)
			}
		}
		if len(forbidden) > 0 {
			checkNoForbiddenImports(t, bcDir, forbidden)
		}
	}
}

func checkNoForbiddenImports(t *testing.T, dir string, forbidden []string) {
	t.Helper()
	files := goFilesIn(t, dir)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s failed: %v", file, err)
		}
		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, "\"")
			for _, prefix := range forbidden {
				if strings.HasPrefix(path, prefix) {
					t.Errorf("DDD boundary violation in %s: imports forbidden package %s", file, path)
				}
			}
		}
	}
}

// checkNoExactImports rejects exact import path matches (for stdlib packages
// where prefix matching would be too broad, e.g. "os" should not match
// "os/user" unless "os/user" is also listed).
func checkNoExactImports(t *testing.T, dir string, forbidden []string) {
	t.Helper()
	forbiddenSet := make(map[string]bool, len(forbidden))
	for _, f := range forbidden {
		forbiddenSet[f] = true
	}

	files := goFilesIn(t, dir)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s failed: %v", file, err)
		}
		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, "\"")
			if forbiddenSet[path] {
				t.Errorf("DDD invariant violation in %s: imports forbidden package %q", file, path)
			}
		}
	}
}

// goFilesIn returns all .go files in dir and its immediate subdirectories.
func goFilesIn(t *testing.T, dir string) []string {
	t.Helper()
	// First try single-level (bounded context packages under domain/).
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("glob %s failed: %v", dir, err)
	}
	// Also pick up files one level deeper (e.g., domain/skillsync/*.go).
	deeper, err := filepath.Glob(filepath.Join(dir, "*", "*.go"))
	if err != nil {
		t.Fatalf("glob %s/* failed: %v", dir, err)
	}
	return append(files, deeper...)
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	// Resolve from current package dir by walking up until go.mod exists.
	start, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("abs path failed: %v", err)
	}
	cur := start
	for i := 0; i < 8; i++ {
		if hasGoMod(cur) {
			return cur
		}
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		cur = next
	}
	t.Fatalf("could not locate repository root from %s", start)
	return ""
}

func hasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}
