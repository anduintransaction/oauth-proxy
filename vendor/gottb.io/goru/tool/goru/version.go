package main

import (
	"gottb.io/goru"
	"gottb.io/goru/tool"
)

func cmdVersion(cmdLine *tool.CommandLine) {
	goru.Println(goru.ColorWhite, goru.Version())
}
