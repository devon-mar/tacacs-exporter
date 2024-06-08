package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort    = "probe"
	defaultTimeout = 5
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Username            string        `yaml:"username"`
	Password            string        `yaml:"password"`
	Secret              []byte        `yaml:"-"`
	SingleConnect       bool          `yaml:"single_connect"`
	LegacySingleConnect bool          `yaml:"legacy_single_connect"`
	PrivLevel           uint8         `yaml:"privilege_level"`
	Port                string        `yaml:"port"`
	Timeout             time.Duration `yaml:"-"`
}

func (m *Module) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tempModule Module
	temp := struct {
		Secret      string `yaml:"secret"`
		Timeout     int    `yaml:"timeout"`
		*tempModule `yaml:",inline"`
	}{
		Timeout:    defaultTimeout,
		tempModule: (*tempModule)(m),
	}
	// Defaults
	temp.Timeout = defaultTimeout
	temp.Port = defaultPort

	if err := unmarshal(&temp); err != nil {
		return err
	}

	if temp.Username == "" {
		return errors.New("username must not be empty")
	}
	if temp.Password == "" {
		return errors.New("password must not be empty")
	}
	if temp.Secret == "" {
		return errors.New("secret must not be empty")
	}
	m.Secret = []byte(temp.Secret)
	m.Timeout = time.Second * time.Duration(temp.Timeout)

	return nil
}

func LoadFromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	c := &Config{}

	if err := d.Decode(c); err != nil {
		return nil, err
	}

	if len(c.Modules) == 0 {
		return nil, errors.New("a config must have at least one module")
	}

	return c, nil
}
