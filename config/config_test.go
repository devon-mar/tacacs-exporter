package config

import (
	"bytes"
	"testing"
	"time"

	yaml "gopkg.in/yaml.v2"
)

func TestModuleUnmarshalYAML(t *testing.T) {
	testCases := []struct {
		yml  string
		want *Module
	}{
		// Minimum required
		{
			`username: test
password: password
secret: secret`,
			&Module{
				Username: "test",
				Password: "password",
				Secret:   []byte("secret"),
				Timeout:  time.Second * 5,
				Port:     defaultPort,
			},
		},
		// Missing username
		{
			`password: test
secret: secret`,
			nil,
		},
		// Missing password
		{
			`username: test
secret: secret`,
			nil,
		},
		// Missing secret
		{
			`username: test
password: test`,
			nil,
		},
		// All possible options
		{

			`username: test
password: test
secret: secret
single_connect: true
legacy_single_connect: true
timeout: 2
privilege_level: 1
port: tty0`,
			&Module{
				Username:            "test",
				Password:            "test",
				Secret:              []byte("secret"),
				SingleConnect:       true,
				LegacySingleConnect: true,
				Timeout:             time.Second * time.Duration(2),
				PrivLevel:           1,
				Port:                "tty0",
			},
		},
		// Invalid YAML
		{`abcdefg`, nil},
	}

	for i, tc := range testCases {
		m := Module{}
		err := yaml.UnmarshalStrict([]byte(tc.yml), &m)
		if err == nil && tc.want == nil {
			t.Errorf("[%d] expected an error but got %v", i, m)
		} else if err != nil && tc.want != nil {
			t.Errorf("[%d] expected no error but got %v", i, err)
		} else if tc.want != nil {
			assertModuleEquals(t, i, &m, tc.want)
		}
	}
}

func assertModuleEquals(t *testing.T, idx int, have *Module, want *Module) {
	t.Helper()

	if have.Username != want.Username {
		t.Errorf("[%d] Username: wanted %q but got %q", idx, want.Username, have.Username)
	}
	if have.Password != want.Password {
		t.Errorf("[%d] Password: wanted %q but got %q", idx, want.Password, have.Password)
	}
	if !bytes.Equal(have.Secret, want.Secret) {
		t.Errorf("[%d] Secret: wanted %q but got %q", idx, want.Secret, have.Secret)
	}
	if have.SingleConnect != want.SingleConnect {
		t.Errorf("[%d] SingleConnect: wanted %t but got %t", idx, want.SingleConnect, have.SingleConnect)
	}
	if have.LegacySingleConnect != want.LegacySingleConnect {
		t.Errorf("[%d] LegacySingleConnect: wanted %t but got %t", idx, want.LegacySingleConnect, have.LegacySingleConnect)
	}
	if have.PrivLevel != want.PrivLevel {
		t.Errorf("[%d] PrivLevel: wanted %d but got %d", idx, want.PrivLevel, have.PrivLevel)
	}
	if have.Port != want.Port {
		t.Errorf("[%d] Port: wanted %q but got %q", idx, want.Port, have.Port)
	}
	if have.Timeout != want.Timeout {
		t.Errorf("[%d] Timeout: wanted %v but got %v", idx, want.Timeout, have.Timeout)
	}
}

func TestLoadFromFile(t *testing.T) {
	c, err := LoadFromFile("testdata/valid.yml")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	m, ok := c.Modules["test"]
	if !ok {
		t.Fatalf("Module 'test' not found")
	}
	want := Module{
		Username: "test",
		Password: "test",
		Secret:   []byte("test"),
		Port:     defaultPort,
		Timeout:  time.Second * 5,
	}
	assertModuleEquals(t, 0, &m, &want)
}

func TestLoadFromFileInvalidPath(t *testing.T) {
	_, err := LoadFromFile("testdata/invalid.yml")
	if err == nil {
		t.Errorf("Expected an error but got nil")
	}
}

func TestLoadFromFileNoModules(t *testing.T) {
	_, err := LoadFromFile("testdata/no_modules.yml")
	if err == nil {
		t.Errorf("Expected an error but got nil")
	}
}

func TestLoadFromFileInvalidYAML(t *testing.T) {
	_, err := LoadFromFile("testdata/invalid_yaml.yml")
	if err == nil {
		t.Errorf("Expected an error but got nil")
	}
}
