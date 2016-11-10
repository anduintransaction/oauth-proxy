package generator

import (
	"os"
	"path/filepath"
	"regexp"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/packer"
)

func packFolder(root, pkgName string) (err error) {
	runmode := os.Getenv("RUNMODE")
	goru.Printf(goru.ColorWhite, "-> Generating packer for %q, package name: %q, mode: %q\n", root, pkgName, runmode)
	generatedFolder := filepath.Join(generatedRoot, "assets", pkgName)
	err = os.MkdirAll(generatedFolder, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	defer func() {
		if err != nil {
			os.RemoveAll(generatedFolder)
		}
	}()
	generatedFile, err := os.Create(filepath.Join(generatedFolder, generatedRootFile))
	if err != nil {
		return errors.Wrap(err)
	}
	defer generatedFile.Close()
	developmentMode := false
	if runmode == "development" {
		developmentMode = true
	}
	os.MkdirAll(root, 0755)
	err = packer.Generate(pkgName, root, developmentMode, true, generatedFile)
	if err != nil {
		return err
	}
	return errors.Wrap(generatedFile.Chmod(0644))
}

func AssetGenerator(args []string, options map[string]string) {
	if len(args) < 1 {
		goru.ErrPrintln(goru.ColorRed, "-> Asset source was not specified")
		os.Exit(1)
	}
	root := args[0]
	pkgNameRegex, err := regexp.Compile("[^a-zA-Z0-9_\\.]")
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
	pkgName := pkgNameRegex.ReplaceAllString(filepath.Base(root), "")
	err = packFolder(root, pkgName)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
}
