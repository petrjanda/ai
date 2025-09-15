package main

import (
	"context"
	"fmt"
	"time"
	"github.com/getsynq/ai/pkg/ai/eval"
{{- range $pkg, $funcs := . }}
	{{(index $funcs 0).Alias}} "{{$pkg}}"
{{- end }}
)

type EvalFunc func(context.Context) (*eval.Suite, error)

type Case struct {
	Name string
	Func EvalFunc
}

func run(cases []Case) {
	ctx := context.Background()
	var failed int
	start := time.Now()
	for _, cs := range cases {
		suite, err := cs.Func(ctx)
		fmt.Printf("=== RUN   %s, %d cases, %d tasks\n", cs.Name, len(suite.Cases), len(suite.Tasks))
		if err != nil {
			fmt.Printf("--- FAIL: %s (%v)\n    error: %v\n", cs.Name, time.Since(start), err)
			failed++
		} else {
			if err := suite.Run(); err != nil {
				fmt.Printf("--- FAIL: %s (%v)\n    error: %v\n", cs.Name, time.Since(start), err)
			}
			fmt.Printf("--- PASS: %s (%v)\n", cs.Name, time.Since(start))
		}
	}
	if failed > 0 {
		fmt.Printf("FAIL: %d failing eval(s)\n", failed)
	} else {
		fmt.Println("PASS")
	}
}

func main() {
	cases := []Case{
		{{- range $pkg, $funcs := . }}
		{{- range $i, $f := $funcs }}
		{Name: "{{$f.PkgImport}}/{{$f.Name}}", Func: {{$f.Alias}}.{{$f.Name}}},
		{{- end }}
		{{- end }}
	}
	run(cases)
}
