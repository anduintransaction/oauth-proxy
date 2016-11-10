package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"gottb.io/goru"
	"gottb.io/goru/tool"
	"gottb.io/gorux/tool/resources"
)

func cmdResources(cmdLine *tool.CommandLine) {
	src := "resources"
	dest := filepath.Join("generated", "resources")
	if _, ok := cmdLine.Options["src"]; ok {
		src = cmdLine.Options["src"]
	}
	goru.Printf(goru.ColorWhite, "-> Generating resources from %q to %q\n", src, dest)
	if fi, _ := os.Open(src); fi == nil {
		goru.ErrPrintf(goru.ColorRed, "-> Source not found: %q\n", src)
		os.Exit(1)
	}
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	infos, err := ioutil.ReadDir(src)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}
		name := info.Name()
		err = resources.Generate(name, filepath.Join(src, name), dest)
		if err != nil {
			goru.ErrPrintln(goru.ColorRed, err)
			os.Exit(1)
		}
	}
	cmd := exec.Command("goru", "generate", "asset", dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
}
