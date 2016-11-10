package tool

import (
	"os"
	"strings"
)

type CommandLine struct {
	Name    string
	Args    []string
	Options map[string]string
}

func (cmdLine *CommandLine) Build() []string {
	args := []string{}
	for k, v := range cmdLine.Options {
		if v == "" {
			args = append(args, "-"+k)
		} else {
			args = append(args, "-"+k+"="+v)
		}
	}
	for _, arg := range cmdLine.Args {
		args = append(args, arg)
	}
	return args
}

func ParseCommandLine() *CommandLine {
	cmdLine := parseCommandLine(os.Args[1:])
	cmdLine.Name = os.Args[0]
	return cmdLine
}

func parseCommandLine(args []string) *CommandLine {
	cmdLine := &CommandLine{
		Args:    []string{},
		Options: make(map[string]string),
	}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			pieces := strings.SplitN(arg[1:], "=", 2)
			if len(pieces) >= 2 {
				cmdLine.Options[pieces[0]] = pieces[1]
			} else {
				cmdLine.Options[pieces[0]] = ""
			}
		} else {
			cmdLine.Args = append(cmdLine.Args, arg)
		}
	}
	return cmdLine
}
