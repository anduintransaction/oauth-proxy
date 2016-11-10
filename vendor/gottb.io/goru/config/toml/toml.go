package toml

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

type Error struct {
	reason string
}

func (err *Error) Error() string {
	return err.reason
}

var (
	ErrNotFound = &Error{"not found"}
	ErrType     = &Error{"invalid type"}
)

func Build(r io.Reader) (*config.Config, error) {
	decoder := NewDecoder(r)
	value, err := decoder.Decode()
	if err != nil {
		return nil, err
	}
	return config.NewConfig(value, Unmarshal, Update, Append, Merge), nil
}

type TomlType int

const (
	TomlUnknown TomlType = iota
	TomlString
	TomlInt
	TomlFloat
	TomlBoolean
	TomlDatetime
	TomlArray
	TomlTable
	TomlTableArray
)

type Value struct {
	objType TomlType
	value   interface{}
	inArray bool
}

func (t *Value) Get(name string) (config.Value, error) {
	keys, err := t.parsePath(name)
	if err != nil {
		return nil, err
	}
	current := t
	for _, key := range keys {
		switch current.objType {
		case TomlTable:
			value, ok := current.value.(map[string]*Value)
			if !ok {
				return nil, errors.Wrap(ErrType)
			}
			child, ok := value[key]
			if !ok {
				return nil, errors.Wrap(ErrNotFound)
			}
			current = child
		case TomlArray, TomlTableArray:
			index, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, errors.Wrap(ErrType)
			}
			value, ok := current.value.([]*Value)
			if !ok {
				return nil, errors.Wrap(ErrType)
			}
			if len(value) <= int(index) {
				return nil, errors.Wrap(ErrNotFound)
			}
			current = value[int(index)]
		default:
			return nil, errors.Wrap(ErrType)
		}
	}
	return current, nil
}

func (t *Value) At(index int) (config.Value, error) {
	if t.objType != TomlArray && t.objType != TomlTableArray {
		return nil, errors.Wrap(ErrType)
	}
	value := t.value.([]*Value)
	if len(value) <= index {
		return nil, errors.Wrap(ErrNotFound)
	}
	return value[index], nil
}

func (t *Value) Len() int {
	switch t.objType {
	case TomlArray, TomlTableArray:
		value, ok := t.value.([]*Value)
		if !ok {
			return 0
		}
		return len(value)
	case TomlTable:
		value, ok := t.value.(map[string]*Value)
		if !ok {
			return 0
		}
		return len(value)
	default:
		return 0
	}
}

func (t *Value) Keys() []string {
	if t.objType != TomlTable {
		return nil
	}
	keys := []string{}
	value, ok := t.value.(map[string]*Value)
	if !ok {
		return nil
	}
	for k := range value {
		keys = append(keys, k)
	}
	return keys
}

func (t *Value) String() string {
	b := &bytes.Buffer{}
	encoder := NewEncoder(b)
	encoder.Encode(t)
	return b.String()
}

func (t *Value) getOrCreate(path string) (*Value, error) {
	keys, err := t.parsePath(path)
	if err != nil {
		return nil, err
	}
	current := t
	for i, key := range keys {
		switch current.objType {
		case TomlTable:
			value, ok := current.value.(map[string]*Value)
			if !ok {
				return nil, errors.Wrap(ErrType)
			}
			child, ok := value[key]
			if !ok {
				if i < len(keys)-1 {
					value[key] = newTable()
				} else {
					value[key] = &Value{objType: TomlUnknown}
				}
				current = value[key]
			} else {
				current = child
			}
		case TomlArray, TomlTableArray:
			index, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				return nil, errors.Wrap(ErrType)
			}
			value, ok := current.value.([]*Value)
			if !ok {
				return nil, errors.Wrap(ErrType)
			}
			if len(value) <= int(index) {
				return nil, errors.Wrap(ErrNotFound)
			}
			current = value[int(index)]
		default:
			return nil, errors.Wrap(ErrType)
		}
	}
	return current, nil
}

func (t *Value) make(keys []string, arrayTable bool) (*Value, error) {
	current := t
	for i, key := range keys {
		if current.objType != TomlTable {
			return nil, errors.Wrap(ErrType)
		}
		switch value := current.value.(type) {
		case map[string]*Value:
			child, ok := value[key]
			if !ok {
				if i < len(keys)-1 || !arrayTable {
					value[key] = newTable()
					current = value[key]
				} else {
					value[key] = newTableArray()
					current = newTable()
					current.inArray = true
					value[key].add(current)
				}
			} else {
				if child.objType != TomlTable && child.objType != TomlTableArray {
					return nil, errors.Wrap(ErrType)
				}
				if child.objType == TomlTable {
					if i >= len(keys)-1 && arrayTable {
						return nil, errors.Wrap(ErrType)
					}
					current = child
				} else {
					switch childValue := child.value.(type) {
					case []*Value:
						if len(childValue) == 0 {
							return nil, errors.Wrap(ErrNotFound)
						}
						if i >= len(keys)-1 {
							if !arrayTable {
								return nil, errors.Wrap(ErrType)
							}
							current = newTable()
							current.inArray = true
							child.add(current)
						} else {
							current = childValue[len(childValue)-1]
						}
					default:
						return nil, errors.Wrap(ErrType)
					}
				}
			}
		default:
			return nil, errors.Wrap(ErrType)
		}
	}
	return current, nil
}

func (t *Value) set(key string, child *Value) error {
	if t.objType != TomlTable {
		return errors.Wrap(ErrType)
	}
	value, ok := t.value.(map[string]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	value[key] = child
	return nil
}

func (t *Value) add(child *Value) error {
	if t.objType != TomlArray && t.objType != TomlTableArray {
		return errors.Wrap(ErrType)
	}
	value, ok := t.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	value = append(value, child)
	t.value = value
	return nil
}

func (t *Value) parsePath(path string) ([]string, error) {
	b := bytes.NewBufferString(path)
	lex := &lexer{}
	tokens, err := lex.lex(b)
	if err != nil {
		return nil, err
	}
	p := &parser{}
	keys := []string{}
	for _, tok := range tokens {
		if p.isValidKey(tok) {
			keys = append(keys, tok.stringValue)
		}
	}
	return keys, nil
}

func newTable() *Value {
	return &Value{
		objType: TomlTable,
		value:   make(map[string]*Value),
	}
}

func newArray() *Value {
	return &Value{
		objType: TomlArray,
		value:   []*Value{},
	}
}

func newTableArray() *Value {
	return &Value{
		objType: TomlTableArray,
		value:   []*Value{},
	}
}

func newString(str string) *Value {
	return &Value{
		objType: TomlString,
		value:   str,
	}
}

func newInt(i int64) *Value {
	return &Value{
		objType: TomlInt,
		value:   i,
	}
}

func newFloat(f float64) *Value {
	return &Value{
		objType: TomlFloat,
		value:   f,
	}
}

func newBool(b bool) *Value {
	return &Value{
		objType: TomlBoolean,
		value:   b,
	}
}

func newTime(t time.Time) *Value {
	return &Value{
		objType: TomlDatetime,
		value:   t,
	}
}
