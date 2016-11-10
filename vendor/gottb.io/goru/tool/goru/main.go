package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"gottb.io/goru"
	"gottb.io/goru/tool"
)

var errReadVersion = fmt.Errorf("Cannot determine goru library version")

func main() {
	libraryVersion, err := readGoruLibraryVersion()
	if err != nil {
		goru.ErrPrintln(goru.ColorRed, err)
		os.Exit(1)
	}
	binaryVersion := goru.Version()
	if goru.VersionCompare(libraryVersion, binaryVersion) != 0 {
		goru.ErrPrintf(goru.ColorRed, "goru library version (%s) is not the same as binary version (%s), please go to %q and run `go install`",
			libraryVersion, binaryVersion, filepath.Join(os.Getenv("GOPATH"), "src", "gottb.io", "goru", "tool", "goru"))
		os.Exit(1)
	}
	commander := tool.Commander{
		"build": &tool.Command{
			Func:        cmdBuild,
			Description: "build goru app in development mode",
		},
		"clean": &tool.Command{
			Func:        cmdClean,
			Description: "clean goru app",
		},
		"run": &tool.Command{
			Func:        cmdRun,
			Description: "run goru app",
		},
		"create": &tool.Command{
			Func:        cmdCreate,
			Description: "create goru app",
		},
		"generate": &tool.Command{
			Func:        cmdGenerate,
			Description: "generate necessary files for goru app",
		},
		"packer": &tool.Command{
			Func:        cmdPacker,
			Description: "pack file into go code",
		},
		"dist": &tool.Command{
			Func:        cmdDist,
			Description: "build goru app in production mode",
		},
		"version": &tool.Command{
			Func:        cmdVersion,
			Description: "show goru version",
		},
	}
	commander.Run()
}

func readGoruLibraryVersion() (string, error) {
	miscFile := filepath.Join("vendor", "gottb.io", "goru", "misc.go")
	_, err := os.Stat(miscFile)
	if err != nil {
		miscFile = filepath.Join(os.Getenv("GOPATH"), "src", "gottb.io", "goru", "misc.go")
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, miscFile, nil, 0)
	if err != nil {
		return "", err
	}
	obj := f.Scope.Lookup("GORU_VERSION")
	if obj == nil {
		return "", errReadVersion
	}
	valueSpec, ok := obj.Decl.(*ast.ValueSpec)
	if !ok {
		return "", errReadVersion
	}
	if len(valueSpec.Values) == 0 {
		return "", errReadVersion
	}
	lit, ok := valueSpec.Values[0].(*ast.BasicLit)
	if !ok {
		return "", errReadVersion
	}
	version := strings.Trim(lit.Value, "\"")
	return version, nil
}
