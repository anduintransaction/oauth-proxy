package toml

import (
	"fmt"
	"reflect"
	"time"

	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

func Update(value config.Value, path string, data interface{}) error {
	toml, ok := value.(*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	child, err := toml.getOrCreate(path)
	if err != nil {
		return err
	}
	v := reflect.ValueOf(data)
	return update(child, &v)
}

func update(toml *Value, v *reflect.Value) error {
	t := v.Type()
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if toml.inArray && toml.objType != TomlInt {
			return errors.Wrap(ErrType)
		}
		return updateInt(toml, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if toml.inArray && toml.objType != TomlInt {
			return errors.Wrap(ErrType)
		}
		return updateUint(toml, v)
	case reflect.Float32, reflect.Float64:
		if toml.inArray && toml.objType != TomlFloat {
			return errors.Wrap(ErrType)
		}
		return updateFloat(toml, v)
	case reflect.Bool:
		if toml.inArray && toml.objType != TomlBoolean {
			return errors.Wrap(ErrType)
		}
		return updateBool(toml, v)
	case reflect.String:
		if toml.inArray && toml.objType != TomlString {
			return errors.Wrap(ErrType)
		}
		return updateString(toml, v)
	case reflect.Array, reflect.Slice:
		if toml.inArray && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		if isValidTensor(t) {
			return updateArray(toml, v, false)
		} else if isValidArrayOfStruct(t) {
			return updateArray(toml, v, true)
		}
		return errors.Wrap(ErrType)
	case reflect.Map:
		if toml.inArray && toml.objType != TomlTable {
			return errors.Wrap(ErrType)
		}
		return updateMap(toml, v)
	case reflect.Struct:
		if t == timeType {
			if toml.inArray && toml.objType != TomlDatetime {
				return errors.Wrap(ErrType)
			}
			return updateDatetime(toml, v)
		}
		if toml.inArray && toml.objType != TomlTable {
			return errors.Wrap(ErrType)
		}
		return updateStruct(toml, v)
	case reflect.Ptr:
		elem := v.Elem()
		return update(toml, &elem)
	default:
		return errors.Wrap(ErrType)
	}
	return nil
}

func updateInt(toml *Value, v *reflect.Value) error {
	toml.objType = TomlInt
	toml.value = v.Int()
	return nil
}

func updateUint(toml *Value, v *reflect.Value) error {
	toml.objType = TomlInt
	toml.value = int64(v.Uint())
	return nil
}

func updateFloat(toml *Value, v *reflect.Value) error {
	toml.objType = TomlFloat
	toml.value = v.Float()
	return nil
}

func updateBool(toml *Value, v *reflect.Value) error {
	toml.objType = TomlBoolean
	toml.value = v.Bool()
	return nil
}

func updateString(toml *Value, v *reflect.Value) error {
	toml.objType = TomlString
	toml.value = v.String()
	return nil
}

func updateDatetime(toml *Value, v *reflect.Value) error {
	toml.objType = TomlDatetime
	value, ok := v.Interface().(time.Time)
	if !ok {
		return errors.Wrap(ErrType)
	}
	toml.value = value
	return nil
}

func updateArray(toml *Value, v *reflect.Value, isArrayOfTable bool) error {
	if isArrayOfTable {
		toml.objType = TomlTableArray
	} else {
		toml.objType = TomlArray
	}
	length := v.Len()
	arr := make([]*Value, length, length)
	for i := 0; i < length; i++ {
		childValue := v.Index(i)
		newToml := &Value{}
		err := update(newToml, &childValue)
		if err != nil {
			return err
		}
		newToml.inArray = true
		arr[i] = newToml
	}
	toml.value = arr
	return nil
}

func updateMap(toml *Value, v *reflect.Value) error {
	toml.objType = TomlTable
	m := make(map[string]*Value)
	for _, k := range v.MapKeys() {
		childValue := v.MapIndex(k)
		newToml := &Value{}
		err := update(newToml, &childValue)
		if err != nil {
			return err
		}
		m[fmt.Sprint(k.Interface())] = newToml
	}
	toml.value = m
	return nil
}

func updateStruct(toml *Value, v *reflect.Value) error {
	toml.objType = TomlTable
	m := make(map[string]*Value)
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("config")
		if len(sf.PkgPath) > 0 || sf.Anonymous || len(tag) == 0 {
			continue
		}
		childValue := v.Field(i)
		newToml := &Value{}
		err := update(newToml, &childValue)
		if err != nil {
			return err
		}
		m[tag] = newToml
	}
	toml.value = m
	return nil
}

func isValidTensor(t reflect.Type) bool {
	for {
		elem := t.Elem()
		if isScalar(elem) {
			return true
		}
		if elem.Kind() != reflect.Array && elem.Kind() != reflect.Slice {
			return false
		}
		t = elem
	}
	return false
}

func isValidArrayOfStruct(t reflect.Type) bool {
	elem := t.Elem()
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	return elem.Kind() == reflect.Struct && elem != timeType
}

func isScalar(t reflect.Type) bool {
	kind := t.Kind()
	if kind >= reflect.Bool && kind <= reflect.Float64 || kind == reflect.String {
		return true
	}
	return t == timeType
}
