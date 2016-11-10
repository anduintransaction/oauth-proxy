package toml

import (
	"gottb.io/goru/config"
	"gottb.io/goru/errors"
)

func Merge(value config.Value, other config.Value) error {
	toml, ok := value.(*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	otherToml, ok := other.(*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	if (toml.objType >= TomlDatetime || toml.inArray) && toml.objType != otherToml.objType {
		return errors.Wrap(ErrType)
	}
	if toml.objType != TomlTable {
		toml.objType = otherToml.objType
		toml.value = otherToml.value
		return nil
	}
	m, ok := toml.value.(map[string]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	otherMap, ok := otherToml.value.(map[string]*Value)
	if !ok {
		return errors.Wrap(ErrType)
	}
	for k, v := range otherMap {
		child, ok := m[k]
		if !ok {
			m[k] = v
		} else {
			err := Merge(child, v)
			if err != nil {
				return err
			}
		}
	}
	toml.value = m
	return nil
}
