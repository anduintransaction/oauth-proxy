package tool

import "testing"

func TestParseCommandLine(t *testing.T) {
	args := []string{"args1", "-key1=value1", "-key2=value 2", "-key3=value=3", "-key4", "args 2"}
	cmdLine := parseCommandLine(args)
	if cmdLine.Args[0] != "args1" || cmdLine.Args[1] != "args 2" {
		t.Fatal(cmdLine.Args)
	}
	if cmdLine.Options["key1"] != "value1" || cmdLine.Options["key2"] != "value 2" ||
		cmdLine.Options["key3"] != "value=3" || cmdLine.Options["key4"] != "" {
		t.Fatal(cmdLine.Options)
	}
}
