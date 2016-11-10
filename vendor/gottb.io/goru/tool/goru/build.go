package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/tool"
)

func cmdBuild(cmdLine *tool.CommandLine) {
	os.Setenv("RUNMODE", "development")
	err := build(cmdLine)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
}

func getAppName() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if os.Getenv("RUNMODE") == "development" {
		return "___" + filepath.Base(cwd), nil
	} else {
		return filepath.Base(cwd), nil
	}
}

func runBuild(cmdLine *tool.CommandLine) error {
	appName, err := getAppName()
	if err != nil {
		return err
	}
	args := []string{"build", "-i", "-o", appName}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run())
}

func build(cmdLine *tool.CommandLine) error {
	err := clean()
	if err != nil {
		return err
	}
	goru.Println(goru.ColorYellow, "Generating files")
	err = runGenerate()
	if err != nil {
		return err
	}
	goru.Println(goru.ColorGreen, "-> SUCCESS")

	goru.Println(goru.ColorYellow, "Building app")
	err = runBuild(cmdLine)
	if err != nil {
		return err
	}
	goru.Println(goru.ColorGreen, "-> SUCCESS")
	return nil
}
