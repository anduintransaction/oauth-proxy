package toml

import (
	"reflect"
	"strconv"
	"time"

	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

func Unmarshal(value config.Value, data interface{}) error {
	toml, ok := value.(*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v := reflect.ValueOf(data)
	return parse(toml, &v)
}

var timeType = reflect.TypeOf(time.Time{})

func parse(toml *Value, v *reflect.Value) error {
	t := v.Type()
	if !v.CanSet() && t.Kind() != reflect.Ptr {
		return errors.Wrap(ErrType)
	}
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		elem := v.Elem()
		return parse(toml, &elem)
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertInt(toml, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertUint(toml, v)
	case reflect.Float32, reflect.Float64:
		return convertFloat(toml, v)
	case reflect.Bool:
		return convertBool(toml, v)
	case reflect.String:
		return convertString(toml, v)
	case reflect.Array, reflect.Slice:
		return convertSlice(toml, v)
	case reflect.Map:
		return convertMap(toml, v)
	case reflect.Struct:
		if t == timeType {
			return convertDatetime(toml, v)
		}
		return convertStruct(toml, v)
		return nil
	default:
		return errors.Wrap(ErrType)
	}
	return nil
}

func convertInt(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlInt {
		return errors.Wrap(ErrType)
	}
	x, ok := toml.value.(int64)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v.SetInt(x)
	return nil
}

func convertUint(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlInt {
		return errors.Wrap(ErrType)
	}
	x, ok := toml.value.(int64)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v.SetUint(uint64(x))
	return nil
}

func convertFloat(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlFloat {
		return errors.Wrap(ErrType)
	}
	x, ok := toml.value.(float64)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v.SetFloat(x)
	return nil
}

func convertBool(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlBoolean {
		return errors.Wrap(ErrType)
	}
	x, ok := toml.value.(bool)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v.SetBool(x)
	return nil
}

func convertString(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlString {
		return errors.Wrap(ErrType)
	}
	str, ok := toml.value.(string)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v.SetString(str)
	return nil
}

func convertDatetime(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlDatetime {
		return errors.Wrap(ErrType)
	}
	t, ok := toml.value.(time.Time)
	if !ok {
		return errors.Wrap(ErrType)
	}
	v.Set(reflect.ValueOf(t))
	return nil
}

func convertSlice(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlArray && toml.objType != TomlTableArray {
		return errors.Wrap(ErrType)
	}
	values, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	l := len(values)
	s := reflect.MakeSlice(v.Type(), l, l)
	for i, child := range values {
		childValue := reflect.New(v.Type().Elem())
		err := parse(child, &childValue)
		if err != nil {
			return err
		}
		s.Index(i).Set(childValue.Elem())
	}
	v.Set(s)
	return nil
}

func convertMap(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlTable {
		return errors.Wrap(ErrType)
	}
	values, ok := toml.value.(map[string]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	m := reflect.MakeMap(v.Type())
	for key, child := range values {
		childValue := reflect.New(v.Type().Elem())
		err := parse(child, &childValue)
		if err != nil {
			return err
		}
		childKey, err := convertStringToValue(key, v.Type().Key())
		if err != nil {
			return err
		}
		m.SetMapIndex(childKey, childValue.Elem())
	}
	v.Set(m)
	return nil
}

func convertStruct(toml *Value, v *reflect.Value) error {
	if toml.objType != TomlTable {
		return errors.Wrap(ErrType)
	}
	values, ok := toml.value.(map[string]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("config")
		if len(sf.PkgPath) > 0 || sf.Anonymous || len(tag) == 0 {
			continue
		}
		childValue, ok := values[tag]
		if !ok {
			continue
		}
		child := v.Field(i)
		err := parse(childValue, &child)
		if err != nil {
			return err
		}
	}
	return nil
}

func convertStringToValue(str string, t reflect.Type) (reflect.Value, error) {
	switch t.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(str)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		return reflect.ValueOf(v), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		switch t.Kind() {
		case reflect.Int:
			return reflect.ValueOf(int(v)), nil
		case reflect.Int8:
			return reflect.ValueOf(int8(v)), nil
		case reflect.Int16:
			return reflect.ValueOf(int16(v)), nil
		case reflect.Int32:
			return reflect.ValueOf(int32(v)), nil
		case reflect.Int64:
			return reflect.ValueOf(int64(v)), nil
		default:
			return reflect.ValueOf(nil), nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		switch t.Kind() {
		case reflect.Uint:
			return reflect.ValueOf(int(v)), nil
		case reflect.Uint8:
			return reflect.ValueOf(uint8(v)), nil
		case reflect.Uint16:
			return reflect.ValueOf(uint16(v)), nil
		case reflect.Uint32:
			return reflect.ValueOf(uint32(v)), nil
		case reflect.Uint64:
			return reflect.ValueOf(uint64(v)), nil
		default:
			return reflect.ValueOf(nil), nil
		}
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return reflect.ValueOf(nil), err
		}
		switch t.Kind() {
		case reflect.Float32:
			return reflect.ValueOf(float32(v)), nil
		case reflect.Float64:
			return reflect.ValueOf(float64(v)), nil
		default:
			return reflect.ValueOf(nil), nil
		}
	case reflect.String:
		return reflect.ValueOf(str), nil
	default:
		return reflect.ValueOf(nil), errors.Wrap(ErrType)
	}
}
