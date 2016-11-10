//go:generate goru packer template packer -out=.
package packer

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gottb.io/goru/errors"
	"gottb.io/goru/utils"
)

func Generate(packageName, root string, dev, useImport bool, w io.Writer) error {
	rootFolder, err := os.Open(root)
	if err != nil {
		return errors.Wrap(err)
	}
	defer rootFolder.Close()

	stat, err := rootFolder.Stat()
	if err != nil {
		return errors.Wrap(err)
	}
	if !stat.IsDir() {
		return errors.Wrap(errInvalidDirectory)
	}

	data := []byte{}
	packDescriptions := make(map[string]*PackDescription)
	if !dev {
		err = utils.Walk(root, func(path string, info os.FileInfo) error {
			if info.IsDir() {
				return nil
			}
			packDescription := &PackDescription{
				Info: &FileInfo{
					FileName:    info.Name(),
					FileSize:    info.Size(),
					FileMode:    info.Mode(),
					FileModTime: info.ModTime(),
				},
			}
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Wrap(err)
			}
			packDescription.Begin = len(data)
			data = append(data, b...)
			packDescription.End = len(data)
			packDescriptions[strings.TrimLeft(strings.TrimPrefix(path, root), "/")] = packDescription
			return nil
		}, ".packer_ignore")
		if err != nil {
			return err
		}
	}
	packJson, err := json.Marshal(packDescriptions)
	if err != nil {
		return errors.Wrap(err)
	}

	mainTmpl, err := LoadTemplate(VVVPack, "main.go.tmpl")
	if err != nil {
		return err
	}
	tmplData := &struct {
		PackageName string
		Files       string
		Root        string
		Dev         bool
		UseImport   bool
		Data        string
	}{
		PackageName: filepath.Base(packageName),
		Files:       string(packJson),
		Root:        root,
		Dev:         dev,
		UseImport:   useImport,
		Data:        string(data),
	}
	err = mainTmpl.Execute(w, tmplData)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

var filenameRegex = regexp.MustCompile("[^a-zA-Z0-9]")
var underscoreRegex = regexp.MustCompile("\\_+")

func makeVariableName(name string) string {
	return "vvv" + makePublicVariableName(underscoreRegex.ReplaceAllString(filenameRegex.ReplaceAllString(strings.ToLower(name), "_"), "_"))
}

var slugRegex = regexp.MustCompile("(\\s|-|_)+")

func makePublicVariableName(name string) string {
	return strings.Replace(strings.Title(slugRegex.ReplaceAllString(name, " ")), " ", "", -1)
}
