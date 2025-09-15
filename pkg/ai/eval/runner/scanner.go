package runner

import (
	"errors"
	"fmt"
	"go/ast"
	"log/slog"
	"strings"

	_ "embed"

	"golang.org/x/tools/go/packages"
)

func FindEvals(dir ...string) ([]evalFunc, error) {

	slog.Info(fmt.Sprintf("finding evals in %v", dir))

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedSyntax |
			packages.NeedFiles |
			packages.NeedModule |
			packages.NeedTypes,
	}

	pkgs, err := packages.Load(cfg, dir...)

	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("errors found in packages")
	}

	var funcs []evalFunc
	for _, pkg := range pkgs {
		if pkg.PkgPath == "" || len(pkg.Syntax) == 0 {
			continue
		}

		slog.Debug("pkg", "pkg", pkg)

		for _, f := range pkg.Syntax {
			filename := pkg.Fset.Position(f.Pos()).Filename
			if !strings.HasSuffix(filename, "_eval.go") {
				continue
			}

			slog.Debug("filename", "filename", filename)

			ast.Inspect(f, func(n ast.Node) bool {
				fd, ok := n.(*ast.FuncDecl)
				if !ok || fd.Recv != nil || fd.Name == nil {
					return true
				}

				if strings.HasPrefix(fd.Name.Name, "Eval") {
					slog.Debug("func", "func", fd.Name.Name)

					modDir := ""
					if pkg.Module != nil {
						modDir = pkg.Module.Dir
					}

					funcs = append(funcs, evalFunc{
						PkgImport: pkg.PkgPath,
						Alias:     safeAlias(pkg.PkgPath),
						Name:      fd.Name.Name,
						ModDir:    modDir,
					})
				}

				return true
			})
		}
	}

	if len(funcs) == 0 {
		return nil, errors.New("no evals found")
	}

	slog.Info(fmt.Sprintf("found %d evals", len(funcs)))

	return funcs, nil
}

func safeAlias(importPath string) string {
	alias := importPath
	if idx := strings.LastIndex(alias, "/"); idx >= 0 {
		alias = alias[idx+1:]
	}
	alias = strings.ReplaceAll(alias, "-", "_")
	return alias
}

// FilterEvalsByName filters eval functions by exact name match
func FilterEvalsByName(funcs []evalFunc, name string) []evalFunc {
	var filtered []evalFunc
	for _, f := range funcs {
		if strings.HasPrefix(f.Name, name) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
