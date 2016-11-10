package resources

import (
	"os"
	"os/exec"
	"path/filepath"

	"gottb.io/goru/errors"
)

type scssGenerator struct {
}

func (g *scssGenerator) Generate(src, dest string) error {
	dest = filepath.Join(dest, "css")
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	sourceFile := filepath.Join(src, "main.scss")
	destFile := filepath.Join(dest, "main-scss.css")
	cmd := exec.Command("scss", sourceFile, destFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run())
}
