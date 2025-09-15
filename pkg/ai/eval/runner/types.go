package runner

type evalFunc struct {
	ModDir    string
	PkgImport string
	Alias     string
	Name      string
}

func newEvalFunc(modDir, pkgImport, alias, name string) evalFunc {
	return evalFunc{
		ModDir:    modDir,
		PkgImport: pkgImport,
		Alias:     alias,
		Name:      name,
	}
}
