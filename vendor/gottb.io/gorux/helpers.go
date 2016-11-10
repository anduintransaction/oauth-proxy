package gorux

import (
	"bytes"
	"html/template"
)

func ExecuteTemplate(t *template.Template, data interface{}) ([]byte, error) {
	b := &bytes.Buffer{}
	err := t.Execute(b, data)
	return b.Bytes(), err
}

func ExecuteTemplateHTML(t *template.Template, data interface{}) (template.HTML, error) {
	b := &bytes.Buffer{}
	err := t.Execute(b, data)
	return template.HTML(b.String()), err
}
