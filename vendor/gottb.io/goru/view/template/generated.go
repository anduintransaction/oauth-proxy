package template

import (
	"gottb.io/goru/packer"
)

var data = []byte("package check\n\nimport (\n\t{{range $k, $v := .Imports}}\n\t{{$v}} \"{{$k}}\"\n\t{{end}}\n)\n\n{{range $k, $v := .Funcs}}\nvar fff_{{$k}} {{$v}}\n{{end}}\npackage check\n\nvar hhh_err error\npackage check\n\nimport (\n\t{{range $k, $v := .Imports}}\n\t{{$k}} \"{{$v}}\"\n\t{{end}}\n)\n\nfunc vvv_{{.Name}}({{.Declaration}}) string {\n    return \"\"\n}\npackage check\n\nimport (\n\t{{range $k, $v := .Imports}}\n\t{{$k}} \"{{$v}}\"\n\t{{end}}\n)\n\n{{range .Args}}\nvar {{.Name}} {{.Type}}\n{{end}}\n")

// VVVPack is archived variable for 'template'
var VVVPack = packer.NewPack(data, []byte("{\"functions.go.tmpl\":{\"Begin\":0,\"End\":140,\"Info\":{\"FileName\":\"functions.go.tmpl\",\"FileSize\":140,\"FileMode\":436,\"FileModTime\":\"2016-03-04T16:47:51.841688414+07:00\"}},\"helpers.go.tmpl\":{\"Begin\":140,\"End\":173,\"Info\":{\"FileName\":\"helpers.go.tmpl\",\"FileSize\":33,\"FileMode\":436,\"FileModTime\":\"2016-03-04T16:47:56.953670922+07:00\"}},\"sig.go.tmpl\":{\"Begin\":173,\"End\":318,\"Info\":{\"FileName\":\"sig.go.tmpl\",\"FileSize\":145,\"FileMode\":436,\"FileModTime\":\"2016-03-04T16:47:46.314707326+07:00\"}},\"variables.go.tmpl\":{\"Begin\":318,\"End\":449,\"Info\":{\"FileName\":\"variables.go.tmpl\",\"FileSize\":131,\"FileMode\":436,\"FileModTime\":\"2016-03-02T13:40:42.337643004+07:00\"}}}"), "template", false)

// Get returns packer object for 'template'
func Get() *packer.Pack {
	return VVVPack
}
