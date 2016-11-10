package tool

import (
	"fmt"
	"os"
)

type Command struct {
	Func        func(cmdLine *CommandLine)
	Description string
}

type Commander map[string]*Command

func (c Commander) Run() {
	cmdLine := ParseCommandLine()
	if len(cmdLine.Args) == 0 {
		c.showHelp()
		return
	}
	cmdName := cmdLine.Args[0]
	command, ok := c[cmdName]
	if !ok || cmdName == "help" {
		c.showHelp()
		return
	}
	cmdLine.Args = cmdLine.Args[1:]
	command.Func(cmdLine)
}

func (c Commander) showHelp() {
	if len(c) > 0 {
		fmt.Fprintln(os.Stderr, "Available commands:")
		for cmdName, command := range c {
			fmt.Fprintf(os.Stderr, "\t%s\t%s\n", cmdName, command.Description)
		}
	}
	os.Exit(1)
}
