package main

import (
	"os"
	"path/filepath"
	"strings"

	"gottb.io/goru"
	"gottb.io/goru/tool"
)

func cmdClean(cmdLine *tool.CommandLine) {
	err := clean()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
}

func clean() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	appName, err := getAppName()
	if err != nil {
		return err
	}
	appName = strings.TrimPrefix(appName, "___")
	os.RemoveAll("dist")
	os.RemoveAll(appName)
	os.RemoveAll("___" + appName)
	os.RemoveAll("generated")
	os.RemoveAll("generated.go")
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".tmpl.go") {
			os.Remove(path)
		}
		return nil
	})

	return nil
}
