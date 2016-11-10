package packer

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"gottb.io/goru/errors"
)

type PackDescription struct {
	Begin int
	End   int
	Info  *FileInfo
}

type Pack struct {
	files map[string]File
	root  string
	dev   bool
}

func NewPack(data, packData []byte, root string, dev bool) *Pack {
	pack := &Pack{
		files: make(map[string]File),
		root:  root,
		dev:   dev,
	}
	if !dev {
		var packDescriptions map[string]*PackDescription
		err := json.Unmarshal(packData, &packDescriptions)
		if err != nil {
			return pack
		}
		for path, packDescription := range packDescriptions {
			pack.files[path] = NewPackerFile(data[packDescription.Begin:packDescription.End], packDescription.Info)
		}
	}
	return pack
}

func (p *Pack) Open(name string) (File, error) {
	name = strings.TrimLeft(name, "/")
	if p.dev {
		f, err := os.Open(filepath.Join(p.root, name))
		if err != nil {
			return nil, errors.Wrap(err)
		}
		return f, nil
	}
	f, ok := p.files[name]
	if !ok {
		return nil, errors.Wrap(&os.PathError{"open", name, errNotFound})
	}
	return f.(*file).Clone(), nil
}

func (p *Pack) List() []string {
	l := []string{}
	for name := range p.files {
		l = append(l, name)
	}
	sort.Strings(l)
	return l
}

func LoadTemplate(packer *Pack, name string) (*template.Template, error) {
	f, err := packer.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	t := template.New(name)
	return t.Parse(string(b))
}
