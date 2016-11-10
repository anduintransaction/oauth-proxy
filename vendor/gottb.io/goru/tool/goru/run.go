package main

import (
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"gottb.io/goru"
	"gottb.io/goru/errors"
	"gottb.io/goru/tool"
	"gottb.io/goru/utils"
	"gottb.io/goru/watcher"
)

func cmdRun(cmdLine *tool.CommandLine) {
	cwd, err := os.Getwd()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
	w := watcher.NewWatcher()
	os.Setenv("RUNMODE", "development")
	binFilename, err := getAppName()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
	cmd, done, err := run(binFilename, cmdLine)
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	if runtime.GOOS != "windows" {
		signal.Notify(signalChan, syscall.SIGTERM)
	}

	err = w.Add(cwd)
	if err != nil {
		utils.Kill(cmd)
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}

	quit := false
	needRebuild := false
	rebuilding := false
	rebuild := false
	appStarted := true
	ticker := time.NewTicker(100 * time.Millisecond)
	lastUpdate := time.Now()
	rebuildTime := 0
	for !quit {
		for !quit && !rebuild {
			select {
			case <-signalChan:
				utils.Kill(cmd)
				if !appStarted {
					quit = true
				}
			case <-done:
				if rebuilding {
					rebuild = true
					break
				}
				quit = true
			case e := <-w.Events:
				lastUpdate = time.Now()
				if needRebuild {
					break
				}
				basename := filepath.Base(e.Name)
				if strings.HasSuffix(basename, ".go") || strings.HasSuffix(basename, ".tmpl.html") || e.Op == watcher.Remove ||
					strings.HasPrefix(e.Name, filepath.Join(cwd, "conf")) {
					needRebuild = true
				}
			case err = <-w.Errors:
				utils.Kill(cmd)
			case now := <-ticker.C:
				if !rebuilding && needRebuild && now.Sub(lastUpdate) >= 500*time.Millisecond {
					if appStarted {
						utils.Kill(cmd)
						rebuilding = true
					} else {
						rebuilding = true
						rebuild = true
					}
				}
			}
		}
		if !quit {
			appStarted = false
			rebuildTime++
			goru.Println(goru.ColorPurple, "Rebuilding app (", rebuildTime, ")")
			w.Close()
			if done != nil {
				close(done)
			}
			os.Setenv("RUNMODE", "development")
			cmd, done, err = run(binFilename, cmdLine)
			if err != nil {
				goru.ErrPrintln(goru.ColorRed, "-> ", err)
			} else {
				appStarted = true
			}

			needRebuild = false
			rebuilding = false
			rebuild = false
			lastUpdate = time.Now()
			w = watcher.NewWatcher()
			w.Add(cwd)
		}
	}
	w.Close()
	if done != nil {
		close(done)
	}
	ticker.Stop()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, "-> ", err)
		os.Exit(1)
	}
}

func run(binFilename string, cmdLine *tool.CommandLine) (*exec.Cmd, chan bool, error) {
	err := build(cmdLine)
	if err != nil {
		return nil, nil, err
	}

	goru.Println(goru.ColorYellow, "Checking app")
	err = check(binFilename, cmdLine)
	if err != nil {
		return nil, nil, err
	}

	goru.Println(goru.ColorYellow, "Running app")
	os.Setenv("RUNMODE", "development")
	cmd, done, err := utils.Fork("./"+binFilename, cmdLine.Build()...)
	if err != nil {
		return nil, nil, err
	}
	return cmd, done, nil
}

func check(binFilename string, cmdLine *tool.CommandLine) error {
	os.Setenv("RUNMODE", "check")
	cmd := exec.Command("./"+binFilename, cmdLine.Build()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return errors.Wrap(cmd.Run())
}
