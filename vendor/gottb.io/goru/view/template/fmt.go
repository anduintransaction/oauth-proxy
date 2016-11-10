package template

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
)

type formatter struct {
	imports map[string]string
	index   int
	context string
}

func newFormatter() *formatter {
	return &formatter{
		imports: make(map[string]string),
		index:   0,
		context: "",
	}
}

func (f *formatter) sprint(t reflect.Type) (string, error) {
	pkgPath := t.PkgPath()
	if pkgPath != "" {
		if pkgPath == "main" {
			return "", &Error{f.context, fmt.Sprintf("template: cannot register function with type in main package: %s", f.context), ""}
		}
		pkgName := f.imports[pkgPath]
		if pkgName == "" {
			pkgName = fmt.Sprintf("%s%d", filepath.Base(pkgPath), f.index)
			f.index++
			f.imports[pkgPath] = pkgName
		}
		return fmt.Sprintf("%s.%s", pkgName, t.Name()), nil
	}
	if t.Kind() <= reflect.Complex128 || t.Kind() == reflect.String {
		return fmt.Sprint(t), nil
	}
	switch t.Kind() {
	case reflect.Ptr:
		return f.printPtr(t)
	case reflect.Slice:
		return f.printSlice(t)
	case reflect.Array:
		return f.printArray(t)
	case reflect.Map:
		return f.printMap(t)
	case reflect.Chan:
		return f.printChan(t)
	case reflect.Struct:
		return f.printStruct(t)
	case reflect.Interface:
		return f.printInterface(t)
	case reflect.Func:
		return f.printFunc(t)
	}
	return "", nil
}

func (f *formatter) printPtr(t reflect.Type) (string, error) {
	s, err := f.sprint(t.Elem())
	if err != nil {
		return "", err
	}
	return "*" + s, nil
}

func (f *formatter) printSlice(t reflect.Type) (string, error) {
	s, err := f.sprint(t.Elem())
	if err != nil {
		return "", err
	}
	return "[]" + s, nil
}

func (f *formatter) printArray(t reflect.Type) (string, error) {
	s, err := f.sprint(t.Elem())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[%d]%s", t.Len(), s), nil
}

func (f *formatter) printMap(t reflect.Type) (string, error) {
	sKey, err := f.sprint(t.Key())
	if err != nil {
		return "", err
	}
	sValue, err := f.sprint(t.Elem())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("map[%s]%s", sKey, sValue), nil
}

func (f *formatter) printChan(t reflect.Type) (string, error) {
	s, err := f.sprint(t.Elem())
	if err != nil {
		return "", err
	}
	switch t.ChanDir() {
	case reflect.RecvDir:
		return fmt.Sprintf("<-chan %s", s), nil
	case reflect.SendDir:
		return fmt.Sprintf("chan<- %s", s), nil
	default:
		return fmt.Sprintf("chan %s", s), nil
	}
}

func (f *formatter) printStruct(t reflect.Type) (string, error) {
	b := &bytes.Buffer{}
	b.WriteString("struct{")
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			return "", &Error{f.context, fmt.Sprintf("unexported field in struct in function %s", f.context), ""}
		}
		s, err := f.sprint(field.Type)
		if err != nil {
			return "", err
		}
		if !field.Anonymous {
			b.WriteString(field.Name)
			b.WriteString(" ")
		}
		b.WriteString(s)
		if i < t.NumField()-1 {
			b.WriteString("; ")
		}
	}
	b.WriteString("}")
	return b.String(), nil
}

func (f *formatter) printInterface(t reflect.Type) (string, error) {
	if t.Name() == "error" {
		return "error", nil
	}
	b := &bytes.Buffer{}
	b.WriteString("interface{")
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if method.PkgPath != "" {
			return "", &Error{f.context, fmt.Sprintf("unexported method in interface in function %s", f.context), ""}
		}
		s, err := f.sprint(method.Type)
		if err != nil {
			return "", err
		}
		b.WriteString(method.Name)
		b.WriteString(strings.TrimPrefix(s, "func"))
		if i < t.NumMethod()-1 {
			b.WriteString("; ")
		}
	}
	b.WriteString("}")
	return b.String(), nil
}

func (f *formatter) printFunc(t reflect.Type) (string, error) {
	b := &bytes.Buffer{}
	b.WriteString("func(")
	for i := 0; i < t.NumIn(); i++ {
		if i < t.NumIn()-1 {
			s, err := f.sprint(t.In(i))
			if err != nil {
				return "", err
			}
			b.WriteString(s)
			b.WriteString(", ")
		} else {
			if t.IsVariadic() {
				s, err := f.sprint(t.In(i).Elem())
				if err != nil {
					return "", err
				}
				b.WriteString("...")
				b.WriteString(s)
			} else {
				s, err := f.sprint(t.In(i))
				if err != nil {
					return "", err
				}
				b.WriteString(s)
			}
		}
	}
	b.WriteString(")")
	if t.NumOut() >= 1 {
		b.WriteString(" ")
	}
	if t.NumOut() >= 2 {
		b.WriteString("(")
	}
	for i := 0; i < t.NumOut(); i++ {
		s, err := f.sprint(t.Out(i))
		if err != nil {
			return "", err
		}
		b.WriteString(s)
		if i < t.NumOut()-1 {
			b.WriteString(", ")
		}
	}
	if t.NumOut() >= 2 {
		b.WriteString(")")
	}
	return b.String(), nil
}
