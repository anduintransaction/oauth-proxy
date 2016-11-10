package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/tool"
)

func cmdDist(cmdLine *tool.CommandLine) {
	err := dist(cmdLine)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
}

func dist(cmdLine *tool.CommandLine) (err error) {
	distFolder := cmdLine.Options["out"]
	if distFolder == "" {
		distFolder = "dist"
	}
	defer func() {
		if err != nil {
			os.RemoveAll(distFolder)
		}
	}()
	err = build(cmdLine)
	if err != nil {
		return err
	}
	binFilename, err := getAppName()
	if err != nil {
		return err
	}
	goru.Println(goru.ColorYellow, "Checking app")
	err = check(binFilename, cmdLine)
	if err != nil {
		return err
	}
	os.Setenv("RUNMODE", "production")
	err = os.MkdirAll(distFolder, 0755)
	if err != nil {
		return errors.Wrap(err)
	}
	goru.Println(goru.ColorYellow, "Moving file to dist folder")
	err = os.Rename(binFilename, filepath.Join(distFolder, strings.TrimPrefix(binFilename, "___")))
	if err != nil {
		return errors.Wrap(err)
	}
	goru.Println(goru.ColorGreen, "-> SUCCESS")

	goru.Println(goru.ColorYellow, "Creating config file")
	config, err := goru.LoadConfig("conf/app.toml", "conf/production/app.toml", cmdLine)
	if err != nil {
		return err
	}
	configFile, err := os.Create(filepath.Join(distFolder, "app.toml"))
	if err != nil {
		return errors.Wrap(err)
	}
	defer configFile.Close()
	_, err = fmt.Fprint(configFile, config)
	if err != nil {
		return errors.Wrap(err)
	}
	configFile.Chmod(0644)
	goru.Println(goru.ColorGreen, "-> SUCCESS")
	return nil
}
