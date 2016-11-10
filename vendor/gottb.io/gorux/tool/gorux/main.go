package main

import "gottb.io/goru/tool"

func main() {
	commander := tool.Commander{
		"resources": &tool.Command{
			Func:        cmdResources,
			Description: "generate less, sass, typescript or coffeescript resources",
		},
	}
	commander.Run()
}
