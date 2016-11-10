package main

import (
	"os"
	"path/filepath"

	"gottb.io/goru"
	"gottb.io/goru/packer"
	"gottb.io/goru/tool"
	"gottb.io/goru/utils"
)

func cmdPacker(cmdLine *tool.CommandLine) {
	if len(cmdLine.Args) < 2 {
		goru.ErrPrintln(goru.ColorWhite, "USAGE: goru packer <folder> <package name> [-out=folder] [-d]")
		os.Exit(1)
	}
	root := cmdLine.Args[0]
	pkgName := cmdLine.Args[1]
	outputFolder, ok := cmdLine.Options["out"]
	if !ok {
		outputFolder = pkgName
	}
	err := os.MkdirAll(outputFolder, 0755)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	pkgPath, err := utils.PkgPath(outputFolder)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	useImport := true
	if pkgPath == "gottb.io/goru/packer" {
		useImport = false
	}
	generatedFile, err := os.Create(filepath.Join(outputFolder, "generated.go"))
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	defer generatedFile.Close()
	_, developmentMode := cmdLine.Options["d"]
	err = packer.Generate(pkgName, root, developmentMode, useImport, generatedFile)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	generatedFile.Chmod(0644)
}
