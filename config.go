package config

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const Separator = "="

var ErrNotFound = errors.New("not found")
var ErrUnsupportedType = errors.New("unsupported type")

type Reader interface {
	ReadString(delim byte) (string, error)
}

type Config struct {
	Values   map[string]string
	Defaults map[string]string
}

func (c *Config) LoadFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return c.LoadReader(bufio.NewReader(file))
}

func (c *Config) LoadString(str string) error {
	return c.LoadReader(bytes.NewBuffer([]byte(str)))
}

func (c *Config) LoadReader(r Reader) error {
	if c.Values == nil {
		c.Values = make(map[string]string)
	}
	for {
		line, err := r.ReadString(byte('\n'))
		if err != nil {
			if err == io.EOF {
				// last line
				c.setLine(line)
				return nil
			}
			return err
		}
		c.setLine(line)
	}
	return nil
}

func (c *Config) setLine(line string) {
	trim := strings.TrimSpace(line)
	if trim == "" {
		return // empty
	} else if strings.HasPrefix(trim, "#") {
		return // comment
	} else if strings.Index(trim, Separator) <= 0 {
		return // invalid
	}
	pair := strings.SplitN(trim, Separator, 2)
	if len(pair) != 2 {
		return // invalid
	}
	name := strings.TrimSpace(pair[0])
	value := strings.TrimSpace(pair[1])
	c.Values[name] = value
}

func (c *Config) SetDefault(name, value string) {
	if c.Defaults == nil {
		c.Defaults = make(map[string]string)
	}
	n := strings.TrimSpace(name)
	v := strings.TrimSpace(value)
	c.Defaults[n] = v
}

func (c *Config) String(name string) (string, error) {
	v, ok := c.Values[name]
	if !ok {
		v, ok = c.Defaults[name]
		if !ok {
			return "", ErrNotFound
		}
	}
	return v, nil
}

func (c *Config) Int(name string) (int, error) {
	v, err := c.String(name)
	if err != nil {
		return 0, err
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func (c *Config) Bool(name string) (bool, error) {
	v, err := c.String(name)
	if err != nil {
		return false, err
	}
	i, err := strconv.ParseBool(v)
	if err != nil {
		return false, err
	}
	return i, nil
}

func (c *Config) Load(st interface{}) error {
	v := reflect.ValueOf(st)
	k := v.Kind()
	if k != reflect.Ptr && k != reflect.Interface {
		return ErrUnsupportedType
	} else if v.IsNil() {
		return ErrUnsupportedType
	}
	e := v.Elem()
	switch e.Kind() {
	case reflect.Struct:
		return c.loadStruct(e)
	//case reflect.Map:
	//	return c.loadMap(e)
	case reflect.Interface, reflect.Ptr:
		return c.Load(e)
	default:
		return ErrUnsupportedType
	}
}

func (c *Config) loadStruct(v reflect.Value) error {
	t := v.Type()
	n := t.NumField()
	for i := 0; i < n; i++ {
		name := fieldName(t.Field(i))
		f := v.Field(i)
		err := c.loadField(f, name)
		if err != nil && err != ErrNotFound {
			return err
		}
	}
	return nil
}

func fieldName(f reflect.StructField) string {
	if f.Anonymous {
		return ""
	}
	tag := f.Tag.Get("config")
	if tag != "" {
		if tag == "-" {
			return ""
		}
		tagParts := strings.Split(tag, ",")
		if len(tagParts) >= 1 {
			return strings.TrimSpace(tagParts[0])
		}
	}
	return f.Name
}

func (c *Config) loadField(f reflect.Value, name string) error {
	switch f.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.loadFieldInt(f, name)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return c.loadFieldUint(f, name)
	//case reflect.Float32, reflect.Float64:
	//	return c.loadFieldFloat(f, name)
	case reflect.String:
		return c.loadFieldString(f, name)
	default:
		return ErrUnsupportedType
	}
}

func (c *Config) loadFieldInt(f reflect.Value, name string) error {
	if name == "" {
		return nil
	}
	i, err := c.Int(name)
	if err != nil {
		return err
	}
	f.SetInt(int64(i))
	return nil
}

func (c *Config) loadFieldUint(f reflect.Value, name string) error {
	if name == "" {
		return nil
	}
	i, err := c.Int(name)
	if err != nil {
		return err
	}
	f.SetUint(uint64(i))
	return nil
}

func (c *Config) loadFieldBool(f reflect.Value, name string) error {
	if name == "" {
		return nil
	}
	i, err := c.Bool(name)
	if err != nil {
		return err
	}
	f.SetBool(i)
	return nil
}

func (c *Config) loadFieldString(f reflect.Value, name string) error {
	if name == "" {
		return nil
	}
	i, err := c.String(name)
	if err != nil {
		return err
	}
	f.SetString(i)
	return nil
}
