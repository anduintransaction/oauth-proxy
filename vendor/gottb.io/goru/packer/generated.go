package packer

import (
	
)

var data = []byte("package {{.PackageName}}\n{{$useImport := .UseImport}}\nimport (\n\t{{if $useImport}}\"gottb.io/goru/packer\"{{end}}\n)\n\nvar data = []byte({{printf \"%#v\" .Data}})\n\n// VVVPack is archived variable for '{{.Root}}'\nvar VVVPack = {{if $useImport}}packer.{{end}}NewPack(data, []byte({{printf \"%#v\" .Files}}), {{printf \"%#v\" .Root}}, {{printf \"%#v\" .Dev}})\n\n// Get returns packer object for '{{.Root}}'\nfunc Get() *{{if $useImport}}packer.{{end}}Pack {\n\treturn VVVPack\n}\n")

// VVVPack is archived variable for 'template'
var VVVPack = NewPack(data, []byte("{\"main.go.tmpl\":{\"Begin\":0,\"End\":458,\"Info\":{\"FileName\":\"main.go.tmpl\",\"FileSize\":458,\"FileMode\":436,\"FileModTime\":\"2016-02-16T15:25:26.941646214+07:00\"}}}"), "template", false)

// Get returns packer object for 'template'
func Get() *Pack {
	return VVVPack
}
