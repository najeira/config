package config

import (
	//"fmt"
	"testing"
)

type person struct {
	Name string
	Age  int
}

func TestString(t *testing.T) {
	c := Config{}
	c.LoadString(`name=piyo
age=28`)

	v, err := c.String("name")
	if v != "piyo" || err != nil {
		t.Errorf("invalid name: %s, %v\n", v, err)
	}

	v, err = c.String("age")
	if v != "28" || err != nil {
		t.Errorf("invalid age: %s, %v\n", v, err)
	}
}

func TestInt(t *testing.T) {
	c := Config{}
	c.LoadString(`name=piyo
age=28`)

	v, err := c.Int("name")
	if v != 0 || err == nil {
		t.Errorf("invalid name: %s, %v\n", v, err)
	}

	v, err = c.Int("age")
	if v != 28 || err != nil {
		t.Errorf("invalid age: %d, %v\n", v, err)
	}
}

func TestLoad(t *testing.T) {
	c := Config{}
	c.LoadString(`Name=piyo
Age=28`)

	p := person{}
	err := c.Load(&p)
	if err != nil {
		t.Errorf("load fail: %v\n", err)
	}

	if p.Name != "piyo" {
		t.Errorf("invalid name: %s\n", p.Name)
	}

	if p.Age != 28 {
		t.Errorf("invalid age: %d\n", p.Age)
	}
}
