package extract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nottolstybai/godex/internal/search"
)

func parsePackage(dir, importPath string, fNames []string) ([]search.Document, error) {
	files := make([]*ast.File, 0, len(fNames))
	fset := token.NewFileSet()

	for _, fName := range fNames {
		f, err := parser.ParseFile(fset, filepath.Join(dir, fName), nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parser.ParseFile: %w", err)
		}
		files = append(files, f)
	}

	pkg, err := doc.NewFromFiles(fset, files, importPath)
	if err != nil {
		return nil, fmt.Errorf("doc.NewFromFiles: %w", err)
	}

	var result []search.Document

	var printErr error

	add := func(name, text string, decl ast.Node) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}

		var declText strings.Builder
		if decl != nil {
			// Same settings as gofmt.
			config := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
			if err := config.Fprint(&declText, fset, decl); err != nil {
				printErr = errors.Join(printErr, fmt.Errorf("print declaration of %s: %w", name, err))
				return
			}
		}

		result = append(result, search.Document{Name: name, Package: importPath, Decl: declText.String(), Text: text})
	}

	addValues := func(values []*doc.Value) {
		for _, value := range values {
			add(importPath+"."+strings.Join(value.Names, ", "), value.Doc, value.Decl)
		}
	}

	addFuncs := func(funcs []*doc.Func) {
		for _, fn := range funcs {
			add(importPath+"."+fn.Name, fn.Doc, fn.Decl)
		}
	}

	add(importPath, pkg.Doc, nil)
	addValues(pkg.Consts)
	addValues(pkg.Vars)
	addFuncs(pkg.Funcs)

	for _, pType := range pkg.Types {
		add(importPath+"."+pType.Name, pType.Doc, pType.Decl)
		addValues(pType.Consts)
		addValues(pType.Vars)
		addFuncs(pType.Funcs)

		for _, method := range pType.Methods {
			add(importPath+"."+pType.Name+"."+method.Name, method.Doc, method.Decl)
		}
	}

	if printErr != nil {
		return nil, printErr
	}

	return result, nil
}

type listedPackage struct {
	Dir        string
	ImportPath string
	GoFiles    []string
	Standard   bool
}

// isHiddenStdPackage reports whether pkg is a standard library package users can't import.
func isHiddenStdPackage(pkg listedPackage) bool {
	if !pkg.Standard {
		return false
	}

	for part := range strings.SplitSeq(pkg.ImportPath, "/") {
		if part == "internal" || part == "vendor" {
			return true
		}
	}

	return false
}

func Load(pattern string) ([]search.Document, error) {
	result := make([]search.Document, 0)
	cmd := exec.Command("go", "list", "-json", pattern)

	var exitErr *exec.ExitError
	out, err := cmd.Output()
	switch {
	case errors.As(err, &exitErr):
		return nil, fmt.Errorf("go list %s: %w: %s", pattern, err, exitErr.Stderr)
	case err != nil:
		return nil, fmt.Errorf("go list %s: %w", pattern, err)
	}

	decoder := json.NewDecoder(bytes.NewReader(out))

	for {
		var pkg listedPackage
		err := decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			return result, nil
		}
		if err != nil {
			return nil, fmt.Errorf("decoder.Decode(%s): %w", pattern, err)
		}

		if isHiddenStdPackage(pkg) {
			continue
		}

		docs, err := parsePackage(pkg.Dir, pkg.ImportPath, pkg.GoFiles)
		if err != nil {
			return nil, fmt.Errorf("parsePackage: %w", err)
		}

		result = append(result, docs...)
	}
}
