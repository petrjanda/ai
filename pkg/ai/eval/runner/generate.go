package runner

import (
	"bytes"
	"html/template"

	_ "embed"
)

//go:embed eval.go.tpl
var mainTplSrc string

func renderMain(funcs []evalFunc) (string, error) {
	var buf bytes.Buffer

	// Group functions by package import path to avoid duplicate imports
	grouped := make(map[string][]evalFunc)
	for _, f := range funcs {
		grouped[f.PkgImport] = append(grouped[f.PkgImport], f)
	}

	err := template.Must(
		template.New("main").Parse(mainTplSrc),
	).
		Funcs(template.FuncMap{
			"safeAlias": safeAlias,
		}).
		Execute(&buf, grouped)

	return buf.String(), err
}
