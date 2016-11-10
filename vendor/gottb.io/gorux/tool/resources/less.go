package resources

import (
	"os"
	"os/exec"
	"path/filepath"

	"gottb.io/goru/errors"
)

type lessGenerator struct {
}

func (g *lessGenerator) Generate(src, dest string) error {
	dest = filepath.Join(dest, "css")
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	sourceFile := filepath.Join(src, "main.less")
	destFile := filepath.Join(dest, "main-less.css")
	cmd := exec.Command("lessc", sourceFile, destFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run())
}
