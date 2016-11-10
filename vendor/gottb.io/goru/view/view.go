package view

import (
	"bytes"
	"io"

	"gottb.io/goru/errors"
	"gottb.io/goru/view/template"
)

type View struct {
	name string
	t    *template.Template
	r    io.ReadCloser
}

func NewView(name string, args []*template.Arg, imports map[string]string) *View {
	return &View{
		name: name,
		t: &template.Template{
			Args:    args,
			Imports: imports,
		},
	}
}

func (v *View) Render(data ...interface{}) ([]byte, error) {
	m := make(map[string]interface{})
	for i := 0; i < len(v.t.Args) && i < len(data); i++ {
		m[v.t.Args[i].DotName] = data[i]
	}
	b := &bytes.Buffer{}
	err := v.t.HTMLTemplate.Execute(b, m)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return b.Bytes(), nil
}

type ViewWrapper interface {
	GetView() *View
}

func init() {
	Funcs(map[string]interface{}{
		"include": include,
	})
}
