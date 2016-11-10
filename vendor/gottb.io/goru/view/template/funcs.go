package template

import (
	"fmt"
	"html/template"
	"reflect"
)

var Funcs = map[string]interface{}{
	"and":      and,
	"call":     call,
	"html":     html,
	"index":    index,
	"js":       js,
	"len":      length,
	"not":      not,
	"or":       or,
	"print":    fmt.Sprint,
	"printf":   fmt.Sprintf,
	"println":  fmt.Sprintln,
	"urlquery": urlquery,

	"eq": eq, // ==
	"ge": ge, // >=
	"gt": gt, // >
	"le": le, // <=
	"lt": lt, // <
	"ne": ne, // !=
}

func AddFunc(funcMap template.FuncMap) {
	for name, f := range funcMap {
		if _, ok := Funcs[name]; ok {
			panic("func has been registered: " + name)
		}
		t := reflect.TypeOf(f)
		if t.Kind() != reflect.Func {
			panic(fmt.Sprintf("%s is not a function (%s)", name, t))
		}
		Funcs[name] = f
	}
}

func and(arg0 interface{}, args ...interface{}) interface{} {
	return nil
}

func or(arg0 interface{}, args ...interface{}) interface{} {
	return nil
}

func not(arg interface{}) bool {
	return false
}

func eq(arg1 interface{}, arg2 ...interface{}) (bool, error) {
	return false, nil
}

func ne(arg1, arg2 interface{}) (bool, error) {
	return false, nil
}

func lt(arg1, arg2 interface{}) (bool, error) {
	return false, nil
}

func le(arg1, arg2 interface{}) (bool, error) {
	return false, nil
}

func gt(arg1, arg2 interface{}) (bool, error) {
	return false, nil
}

func ge(arg1, arg2 interface{}) (bool, error) {
	return false, nil
}

func html(args ...interface{}) string {
	return ""
}

func js(args ...interface{}) string {
	return ""
}

func urlquery(args ...interface{}) string {
	return ""
}

func call(fn interface{}, args ...interface{}) (interface{}, error) {
	return nil, nil
}

func index(item interface{}, indices ...interface{}) (interface{}, error) {
	return nil, nil
}

func length(item interface{}) (int, error) {
	return 0, nil
}
