package config

import (
	"errors"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultPort    = "probe"
	defaultTimeout = 5
)

type Config struct {
	Modules map[string]Module `yaml:"modules"`
}

type Module struct {
	Username            string
	Password            string
	Secret              []byte
	SingleConnect       bool
	LegacySingleConnect bool
	PrivLevel           uint8
	Port                string
	Timeout             time.Duration
}

func (m *Module) UnmarshalYAML(unmarshal func(interface{}) error) error {
	temp := struct {
		Username            string `yaml:"username"`
		Password            string `yaml:"password"`
		Secret              string `yaml:"secret"`
		SingleConnect       bool   `yaml:"single_connect"`
		LegacySingleConnect bool   `yaml:"legacy_single_connect"`
		Timeout             int    `yaml:"timeout"`
		PrivLevel           uint8  `yaml:"privilege_level"`
		Port                string `yaml:"port"`
	}{
		SingleConnect:       false,
		Timeout:             defaultTimeout,
		Port:                defaultPort,
		PrivLevel:           0,
		LegacySingleConnect: false,
	}

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
	m.Username = temp.Username
	m.Password = temp.Password
	m.Secret = []byte(temp.Secret)
	m.SingleConnect = temp.SingleConnect
	m.LegacySingleConnect = temp.LegacySingleConnect
	m.Timeout = time.Second * time.Duration(temp.Timeout)
	m.Port = temp.Port
	m.PrivLevel = temp.PrivLevel

	return nil
}

func LoadFromFile(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := &Config{}

	if err := yaml.UnmarshalStrict(b, c); err != nil {
		return nil, err
	}

	if len(c.Modules) == 0 {
		return nil, errors.New("a config must have at least one module")
	}

	return c, nil
}
