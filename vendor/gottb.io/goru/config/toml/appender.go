package toml

import (
	"reflect"
	"time"

	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

func Append(value config.Value, path string, data ...interface{}) error {
	toml, ok := value.(*Value)
	if !ok {
		return ErrType
	}
	child, err := toml.getOrCreate(path)
	if err != nil {
		return err
	}
	if child.objType != TomlUnknown && child.objType != TomlArray && child.objType != TomlTableArray {
		return errors.Wrap(ErrType)
	}
	if child.objType == TomlUnknown {
		child.value = []*Value{}
	}
	for _, datum := range data {
		err = appendSingle(child, datum)
		if err != nil {
			return err
		}
	}
	return nil
}

func appendSingle(toml *Value, data interface{}) error {
	v := reflect.ValueOf(data)
	return appendValue(toml, &v)
}

func appendValue(toml *Value, v *reflect.Value) error {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		elem := v.Elem()
		return appendValue(toml, &elem)
	}
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if toml.objType != TomlUnknown && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		return appendInt(toml, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if toml.objType != TomlUnknown && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		return appendUint(toml, v)
	case reflect.Float32, reflect.Float64:
		if toml.objType != TomlUnknown && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		return appendFloat(toml, v)
	case reflect.Bool:
		if toml.objType != TomlUnknown && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		return appendBool(toml, v)
	case reflect.String:
		if toml.objType != TomlUnknown && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		return appendString(toml, v)
	case reflect.Slice, reflect.Array:
		if toml.objType != TomlUnknown && toml.objType != TomlArray {
			return errors.Wrap(ErrType)
		}
		return appendSlice(toml, v)
	case reflect.Struct:
		if t == timeType {
			if toml.objType != TomlUnknown && toml.objType != TomlArray {
				return errors.Wrap(ErrType)
			}
			return appendDatetime(toml, v)
		}
		if toml.objType != TomlUnknown && toml.objType != TomlTableArray {
			return errors.Wrap(ErrType)
		}
		return appendStruct(toml, v)
	default:
		return errors.Wrap(ErrType)
	}
	return nil
}

func appendInt(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	arr = append(arr, &Value{
		objType: TomlInt,
		value:   int64(v.Int()),
		inArray: true,
	})
	toml.value = arr
	return nil
}

func appendUint(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	arr = append(arr, &Value{
		objType: TomlInt,
		value:   int64(v.Uint()),
		inArray: true,
	})
	toml.value = arr
	return nil
}

func appendFloat(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	arr = append(arr, &Value{
		objType: TomlFloat,
		value:   float64(v.Float()),
		inArray: true,
	})
	toml.value = arr
	return nil
}

func appendBool(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	arr = append(arr, &Value{
		objType: TomlBoolean,
		value:   v.Bool(),
		inArray: true,
	})
	toml.value = arr
	return nil
}

func appendString(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	arr = append(arr, &Value{
		objType: TomlString,
		value:   v.String(),
		inArray: true,
	})
	toml.value = arr
	return nil
}

func appendDatetime(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	arr = append(arr, &Value{
		objType: TomlDatetime,
		value:   v.Interface().(time.Time),
		inArray: true,
	})
	toml.value = arr
	return nil
}

func appendSlice(toml *Value, v *reflect.Value) error {
	toml.objType = TomlArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	newValue := &Value{
		objType: TomlArray,
		inArray: true,
		value:   []*Value{},
	}
	for i := 0; i < v.Len(); i++ {
		childValue := v.Index(i)
		err := appendValue(newValue, &childValue)
		if err != nil {
			return err
		}
	}
	arr = append(arr, newValue)
	toml.value = arr
	return nil
}

func appendStruct(toml *Value, v *reflect.Value) error {
	toml.objType = TomlTableArray
	arr, ok := toml.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	newToml := &Value{
		objType: TomlTable,
		inArray: true,
	}
	err := update(newToml, v)
	if err != nil {
		return err
	}
	arr = append(arr, newToml)
	toml.value = arr
	return nil
}
