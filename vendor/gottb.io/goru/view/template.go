package view

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"

	"gottb.io/goru/errors"
	gorutemplate "gottb.io/goru/view/template"
)

var (
	ErrNotFound = fmt.Errorf("not found")
)

type TemplateSet struct {
	views map[string]*View
	root  *template.Template
}

func NewTemplateSet() *TemplateSet {
	ts := &TemplateSet{
		views: make(map[string]*View),
		root:  template.New("*root*"),
	}
	return ts
}

func (ts *TemplateSet) Add(name string, r io.ReadCloser, v *View) error {
	v.r = r
	ts.views[name] = v
	return nil
}

func (ts *TemplateSet) Funcs(funcMap template.FuncMap) {
	ts.root.Funcs(funcMap)
}

func (ts *TemplateSet) Init() error {
	for name, v := range ts.views {
		err := ts.initView(name, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TemplateSet) initView(name string, v *View) error {
	b, err := ioutil.ReadAll(v.r)
	if err != nil {
		return errors.Wrap(err)
	}
	defer v.r.Close()
	t := ts.root.New(name)
	t, err = t.Parse(string(b))
	if err != nil {
		return errors.Wrap(err)
	}
	v.t.HTMLTemplate = t
	v.t.Content = b
	return nil
}

var defaultTemplateSet = NewTemplateSet()

func Add(name string, r io.ReadCloser, v *View) error {
	err := defaultTemplateSet.Add(name, r, v)
	if err != nil {
		return err
	}
	if os.Getenv("RUNMODE") == "check" {
		gorutemplate.Templates[name] = v.t
	}
	return nil
}

func Funcs(funcMap template.FuncMap) {
	defaultTemplateSet.Funcs(funcMap)
	if os.Getenv("RUNMODE") == "check" {
		gorutemplate.AddFunc(funcMap)
	}
}

func Init() error {
	return defaultTemplateSet.Init()
}
