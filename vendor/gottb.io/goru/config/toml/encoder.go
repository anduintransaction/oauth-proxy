package toml

import (
	"fmt"
	"io"
	"strings"
	"time"

	"gottb.io/goru/errors"
)

type Encoder struct {
	w     io.Writer
	trail []string
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w, []string{}}
}

func (encoder *Encoder) Encode(t *Value) error {
	return encoder.encode(t, false)
}

func (encoder *Encoder) encode(t *Value, inline bool) error {
	switch t.objType {
	case TomlTable:
		return encoder.writeTable(t, inline)
	case TomlString, TomlInt, TomlFloat, TomlBoolean:
		return encoder.writePrimitive(t, inline)
	case TomlDatetime:
		return encoder.writeDateTime(t, inline)
	case TomlArray:
		return encoder.writeArray(t, inline)
	case TomlTableArray:
		return encoder.writeTableArray(t, inline)
	default:
		return errors.Wrap(ErrType)
	}
}

func (encoder *Encoder) writeTable(t *Value, inline bool) error {
	m, ok := t.value.(map[string]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	primitives := make(map[string]*Value)
	subtables := make(map[string]*Value)
	for k, v := range m {
		switch v.objType {
		case TomlTable, TomlTableArray:
			subtables[k] = v
		default:
			primitives[k] = v
		}
	}
	if len(primitives) > 0 && len(encoder.trail) > 0 || t.inArray || len(m) == 0 {
		tableName := strings.Join(encoder.trail, ".")
		if t.inArray {
			fmt.Fprintln(encoder.w, "[["+tableName+"]]")
		} else {
			fmt.Fprintln(encoder.w, "["+tableName+"]")
		}
	}
	for k, v := range primitives {
		fmt.Fprint(encoder.w, k, " = ")
		err := encoder.Encode(v)
		if err != nil {
			return err
		}
	}
	for k, v := range subtables {
		encoder.trail = append(encoder.trail, encoder.quoteKey(k))
		err := encoder.Encode(v)
		if err != nil {
			return err
		}
		encoder.trail = encoder.trail[0 : len(encoder.trail)-1]
	}
	return nil
}

func (encoder *Encoder) writePrimitive(t *Value, inline bool) error {
	if t.objType == TomlString {
		fmt.Fprintf(encoder.w, "%q", t.value)
	} else {
		fmt.Fprintf(encoder.w, "%v", t.value)
	}
	if !inline {
		fmt.Fprint(encoder.w, "\n")
	}
	return nil
}

func (encoder *Encoder) writeDateTime(t *Value, inline bool) error {
	value, ok := t.value.(time.Time)
	if !ok {
		return errors.Wrap(ErrType)
	}
	fmt.Fprintf(encoder.w, "%s", value.Format(time.RFC3339Nano))
	if !inline {
		fmt.Fprint(encoder.w, "\n")
	}
	return nil
}

func (encoder *Encoder) writeArray(t *Value, inline bool) error {
	a, ok := t.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	fmt.Fprint(encoder.w, "[")
	for i, v := range a {
		err := encoder.encode(v, true)
		if err != nil {
			return err
		}
		if i < len(a)-1 {
			fmt.Fprint(encoder.w, ",")
		}
	}
	fmt.Fprint(encoder.w, "]")
	if !inline {
		fmt.Fprint(encoder.w, "\n")
	}
	return nil
}

func (encoder *Encoder) writeTableArray(t *Value, inline bool) error {
	a, ok := t.value.([]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	for _, v := range a {
		err := encoder.Encode(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (encoder *Encoder) quoteKey(key string) string {
	mustBeQuoted := false
	for _, r := range key {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') && r != '-' && r != '_' {
			mustBeQuoted = true
			break
		}
	}
	if !mustBeQuoted {
		return key
	}
	return fmt.Sprintf("%q", key)
}
