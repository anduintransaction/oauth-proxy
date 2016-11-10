package config

import (
	"fmt"
	"io"
	"time"
)

type Type int

const (
	Scalar Type = iota
	Array
	Map
)

type Config struct {
	value        Value
	unmarshaller Unmarshaller
	updater      Updater
	appender     Appender
	merger       Merger
}

func NewConfig(value Value, unmarshaller Unmarshaller, updater Updater, appender Appender, merger Merger) *Config {
	return &Config{value, unmarshaller, updater, appender, merger}
}

func (c *Config) Get(name string) (*Config, error) {
	value, err := c.value.Get(name)
	if err != nil {
		return nil, err
	}
	return &Config{value, c.unmarshaller, c.updater, c.appender, c.merger}, nil
}

func (c *Config) At(index int) (*Config, error) {
	value, err := c.value.At(index)
	if err != nil {
		return nil, err
	}
	return &Config{value, c.unmarshaller, c.updater, c.appender, c.merger}, nil
}

func (c *Config) Len() int {
	return c.value.Len()
}

func (c *Config) Keys() []string {
	return c.value.Keys()
}

func (c *Config) Unmarshal(data interface{}) error {
	return c.unmarshaller(c.value, data)
}

func (c *Config) UpdateTo(path string, data interface{}) error {
	return c.updater(c.value, path, data)
}

func (c *Config) Update(data interface{}) error {
	return c.updater(c.value, "", data)
}

func (c *Config) AppendTo(path string, data ...interface{}) error {
	return c.appender(c.value, path, data...)
}

func (c *Config) Append(data ...interface{}) error {
	return c.appender(c.value, "", data...)
}

func (c *Config) Merge(other *Config) error {
	return c.merger(c.value, other.value)
}

func (c *Config) Int() (int, error) {
	var i int
	err := c.Unmarshal(&i)
	return i, err
}

func (c *Config) Int64() (int64, error) {
	var i int64
	err := c.Unmarshal(&i)
	return i, err
}

func (c *Config) Float() (float64, error) {
	var f float64
	err := c.Unmarshal(&f)
	return f, err
}

func (c *Config) Str() (string, error) {
	var str string
	err := c.Unmarshal(&str)
	return str, err
}

func (c *Config) Bool() (bool, error) {
	var b bool
	err := c.Unmarshal(&b)
	return b, err
}

func (c *Config) Time() (time.Time, error) {
	var t time.Time
	err := c.Unmarshal(&t)
	return t, err
}

func (c *Config) String() string {
	return fmt.Sprint(c.value)
}

func Int(c *Config, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	return c.Int()
}

func Int64(c *Config, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	return c.Int64()
}

func Float(c *Config, err error) (float64, error) {
	if err != nil {
		return 0, err
	}
	return c.Float()
}

func Str(c *Config, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return c.Str()
}

func Bool(c *Config, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return c.Bool()
}

func Time(c *Config, err error) (time.Time, error) {
	if err != nil {
		return time.Time{}, err
	}
	return c.Time()
}

type Value interface {
	Get(name string) (Value, error)
	At(index int) (Value, error)
	Len() int
	Keys() []string
}

type Unmarshaller func(value Value, data interface{}) error
type Updater func(value Value, path string, data interface{}) error
type Appender func(value Value, path string, data ...interface{}) error
type Merger func(value Value, other Value) error
type Builder func(r io.Reader) (*Config, error)
