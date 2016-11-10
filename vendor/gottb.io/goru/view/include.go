package view

import (
	"html/template"
	"strings"

	"gottb.io/goru/errors"
)

func include(name string, data ...interface{}) (template.HTML, error) {
	if !strings.HasSuffix(name, ".tmpl.html") {
		name += ".tmpl.html"
	}
	v, ok := defaultTemplateSet.views[name]
	if !ok {
		return template.HTML(""), errors.Errorf("view not found: %s", name)
	}
	b, err := v.Render(data...)
	if err != nil {
		return template.HTML(""), err
	}
	return template.HTML(b), nil
}
