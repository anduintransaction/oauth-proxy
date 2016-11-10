package main

import (
	"os"
	"os/exec"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/tool"
	"gottb.io/goru/tool/generator"
)

var generators = map[string]generator.Generator{
	"asset": generator.AssetGenerator,
	"view":  generator.ViewGenerator,
}

func cmdGenerate(cmdLine *tool.CommandLine) {
	args := cmdLine.Args
	options := cmdLine.Options
	if len(args) == 0 {
		goru.ErrPrintln(goru.ColorWhite, "Usage: goru generate <generator> <args...>")
		os.Exit(1)
	}
	generatorName := args[0]
	generator, ok := generators[generatorName]
	if !ok {
		goru.ErrPrintln(goru.ColorWhite, "Generator %s not found\n", generatorName)
		goru.ErrPrintln(goru.ColorWhite, "Available generators:")
		for name := range generators {
			goru.ErrPrintln(goru.ColorWhite, "\t%s\n", name)
		}
		os.Exit(1)
	}
	generator(args[1:], options)
}

func runGenerate() error {
	cmd := exec.Command("go", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run())
}
