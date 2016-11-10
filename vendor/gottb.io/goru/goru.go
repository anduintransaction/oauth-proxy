package goru

import (
	"os"
	"strings"

	"gottb.io/goru/config"
	"gottb.io/goru/config/toml"
	"gottb.io/goru/errors"
	"gottb.io/goru/tool"
	"gottb.io/goru/view"
	"gottb.io/goru/view/template"
)

const (
	defaultAddr = ":8765"
)

var Config *config.Config

func Run(r *Router) {
	cmdLine := tool.ParseCommandLine()
	runmode := GetRunMode()
	if runmode == CheckMode {
		runCheck(cmdLine)
		return
	}
	args := cmdLine.Args
	if runmode != DevMode {
		if len(args) < 2 {
			ErrPrintf(ColorWhite, "Usage: %s [address] [config file]\n", cmdLine.Name)
			os.Exit(1)
		}
	}
	addr := defaultAddr
	if len(args) > 0 {
		addr = args[0]
	}
	var err error
	if len(args) > 1 {
		Config, err = LoadConfig(args[1], "", cmdLine)
	} else {
		Config, err = LoadConfig("conf/app.toml", "conf/private/app.toml", cmdLine)
	}
	if err != nil {
		ErrPrintln(ColorRed, err)
		os.Exit(1)
	}
	Println(ColorGreen, "Running start hooks")
	for _, h := range startHooks {
		err = h(Config)
		if err != nil {
			ErrPrintln(ColorRed, err)
			os.Exit(1)
		}
	}
	Println(ColorGreen, "Init view")
	err = view.Init()
	if err != nil {
		ErrPrintln(ColorRed, err)
		os.Exit(1)
	}
	Println(ColorGreen, "Server started on "+addr)
	server := NewServer(addr, r)
	err = server.ListenAndServe()
	Println(ColorGreen, "Running stop hooks")
	for _, h := range stopHooks {
		h(Config)
	}
	if err != nil {
		ErrPrintln(ColorRed, err)
		os.Exit(1)
	} else {
		Println(ColorGreen, "Server exited successfully")
	}
}

func runCheck(cmdLine *tool.CommandLine) {
	args := cmdLine.Args
	var config *config.Config
	var err error
	if len(args) > 1 {
		config, err = LoadConfig(args[1], "", cmdLine)
	} else {
		config, err = LoadConfig("conf/app.toml", "conf/private/app.toml", cmdLine)
	}
	if err != nil {
		ErrPrintln(ColorRed, err)
		os.Exit(1)
	}
	for _, h := range startHooks {
		h(config)
	}
	err = view.Init()
	if err != nil {
		ErrPrintln(ColorRed, err)
		os.Exit(1)
	}
	err = template.Check()
	if err != nil {
		ErrPrintln(ColorRed, err)
		os.Exit(1)
	}
	Println(ColorGreen, "-> SUCCESS")
}

func LoadConfig(configFilename, extraFilename string, cmdLine *tool.CommandLine) (*config.Config, error) {
	mainConfig, err := loadConfigFromfile(configFilename)
	if err != nil {
		return nil, err
	}
	if extraFilename != "" {
		_, err := os.Stat(extraFilename)
		if err == nil {
			extraConfig, err := loadConfigFromfile(extraFilename)
			if err != nil {
				return nil, err
			}
			err = mainConfig.Merge(extraConfig)
			if err != nil {
				return nil, err
			}
		}
	}
	for key, value := range cmdLine.Options {
		if strings.HasPrefix(key, "config.") {
			key = strings.TrimPrefix(key, "config.")
			mainConfig.UpdateTo(key, value)
		}
	}
	return mainConfig, err
}

func loadConfigFromfile(filename string) (*config.Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer f.Close()
	return toml.Build(f)
}
