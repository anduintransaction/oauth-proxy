package resources

import (
	"os"
	"os/exec"
	"path/filepath"

	"gottb.io/goru/errors"
)

type tsGenerator struct {
}

func (g *tsGenerator) Generate(src, dest string) error {
	dest = filepath.Join(dest, "js")
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	sourceFile := filepath.Join(src, "main.ts")
	destFile := filepath.Join(dest, "main-ts.js")
	cmd := exec.Command("tsc", "--out", destFile, sourceFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run())
}
