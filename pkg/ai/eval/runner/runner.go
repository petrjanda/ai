package runner

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	_ "embed"
)

func RunEvalsByModule(funcs []evalFunc) error {
	byMod := map[string][]evalFunc{}
	for _, f := range funcs {
		byMod[f.ModDir] = append(byMod[f.ModDir], f)
	}

	slog.Info(fmt.Sprintf("running %d evals", len(funcs)))

	for modDir, list := range byMod {
		if modDir == "" {
			// fallback: find go.mod by walking up from any source file:
			// modDir = findGoModDir(filepath.Dir(someFilename))
			// (or just skip with a warning)
			log.Printf("warning: no module dir for some evals; skipping")
			continue
		}

		tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("eval-runner-%d.go", time.Now().UnixNano()))
		src, _ := renderMain(list)
		if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
			return err
		}
		defer os.Remove(tmpFile)

		cmd := exec.Command("go", "run", "-tags=eval", tmpFile) // use -tags if you use //go:build eval
		cmd.Dir = modDir                                        // <-- run from module root (where go.mod lives)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		cmd.Env = os.Environ() // GO111MODULE=on by default in modern Go
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
