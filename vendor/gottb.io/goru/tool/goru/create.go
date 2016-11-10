//go:generate goru packer template generated
package main

import (
	"os"

	"gottb.io/goru"
	"gottb.io/goru/tool"
	"gottb.io/goru/tool/creator"
)

var creators = map[string]creator.Creator{
	"app": creator.AppCreator,
}

func cmdCreate(cmdLine *tool.CommandLine) {
	args := cmdLine.Args
	options := cmdLine.Options
	if len(args) == 0 {
		goru.ErrPrintln(goru.ColorWhite, "Usage: goru create <type> <args...>")
		os.Exit(1)
	}
	creatorName := args[0]
	creator, ok := creators[creatorName]
	if !ok {
		goru.ErrPrintln(goru.ColorWhite, "Creator %s not found\n", creatorName)
		goru.ErrPrintln(goru.ColorWhite, "Available creators:")
		for name := range creators {
			goru.ErrPrintln(goru.ColorWhite, "\t%s\n", name)
		}
		os.Exit(1)
	}
	creator(args[1:], options)
}
