package generator

import (
	"gottb.io/goru/packer"
)

var data = []byte("package generated\n\nimport (\n\t_ \"{{.TargetPkg}}\"\n)\npackage {{.PkgName}}\n\nimport (\n\t_ \"{{.TargetPkgName}}\"\n)\npackage {{.PkgName}}\n\nimport (\n\t\"gottb.io/goru/view\"\n\t\"gottb.io/goru/view/template\"\n\t{{range $k, $v := .Imports}}\n\t{{$k}} \"{{$v}}\"\n\t{{end}}\n)\n\ntype {{.TypeName}} struct {\n\tv *view.View\n}\n\nfunc (v *{{.TypeName}}) GetView() *view.View {\n\treturn v.v\n}\n\nvar {{.VarName}} = &{{.TypeName}}{\n\tv: view.NewView(\n\t\t\"{{.Path}}\",\n\t\t[]*template.Arg{\n\t\t{{range .Args}}\n\t\t\t{{printf \"%#v\" .}},\n\t\t{{end}}\n\t\t},\n\t\t{{printf \"%#v\" .Imports}},\n\t),\n}\n\nfunc (v *{{.TypeName}}) Render({{.Declaration}}) ([]byte, error) {\n\treturn v.v.Render({{.Call}})\n}\npackage views\n\nimport (\n\t\"gottb.io/goru/view\"\n\tx \"{{.PackagePath}}\"\n)\n\nfunc init() {\n\tf, err := VVVPack.Open(\"{{.Path}}\")\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\terr = view.Add(\"{{.Path}}\", f, x.{{.VarName}}.GetView())\n\tif err != nil {\n\t\tpanic(err)\n\t}\n}\n")

// VVVPack is archived variable for 'template'
var VVVPack = packer.NewPack(data, []byte("{\"import.go.tmpl\":{\"Begin\":0,\"End\":50,\"Info\":{\"FileName\":\"import.go.tmpl\",\"FileSize\":50,\"FileMode\":436,\"FileModTime\":\"2016-02-15T09:08:03.958986867+07:00\"}},\"main_init.go.tmpl\":{\"Begin\":50,\"End\":107,\"Info\":{\"FileName\":\"main_init.go.tmpl\",\"FileSize\":57,\"FileMode\":436,\"FileModTime\":\"2016-01-25T14:04:36.470851421+07:00\"}},\"view.go.tmpl\":{\"Begin\":107,\"End\":635,\"Info\":{\"FileName\":\"view.go.tmpl\",\"FileSize\":528,\"FileMode\":436,\"FileModTime\":\"2016-02-26T16:17:17.191150261+07:00\"}},\"view_init.go.tmpl\":{\"Begin\":635,\"End\":883,\"Info\":{\"FileName\":\"view_init.go.tmpl\",\"FileSize\":248,\"FileMode\":436,\"FileModTime\":\"2016-02-17T08:13:25.620322083+07:00\"}}}"), "template", false)

// Get returns packer object for 'template'
func Get() *packer.Pack {
	return VVVPack
}
