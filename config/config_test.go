package config

import (
	"bytes"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestModuleUnmarshalYAML(t *testing.T) {
	testCases := map[string]struct {
		yml  string
		want *Module
	}{
		// Minimum required
		"ok-min": {
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
		"missing-username": {
			`password: test
secret: secret`,
			nil,
		},
		// Missing password
		"missing-password": {
			`username: test
secret: secret`,
			nil,
		},
		// Missing secret
		"missing-secret": {
			`username: test
password: test`,
			nil,
		},
		// All possible options
		"all-options": {
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
		"invalid-yaml": {`abcdefg`, nil},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tc.yml))
			d := yaml.NewDecoder(r)
			d.KnownFields(true)

			var m Module
			err := d.Decode(&m)
			if err == nil && tc.want == nil {
				t.Errorf("expected an error but got %v", m)
			} else if err != nil && tc.want != nil {
				t.Errorf("expected no error but got %v", err)
			} else if tc.want != nil {
				assertModuleEquals(t, &m, tc.want)
			}
		})
	}
}

func assertModuleEquals(t *testing.T, have *Module, want *Module) {
	t.Helper()

	if have.Username != want.Username {
		t.Errorf("Username: wanted %q but got %q", want.Username, have.Username)
	}
	if have.Password != want.Password {
		t.Errorf("Password: wanted %q but got %q", want.Password, have.Password)
	}
	if !bytes.Equal(have.Secret, want.Secret) {
		t.Errorf("Secret: wanted %q but got %q", want.Secret, have.Secret)
	}
	if have.SingleConnect != want.SingleConnect {
		t.Errorf("SingleConnect: wanted %t but got %t", want.SingleConnect, have.SingleConnect)
	}
	if have.LegacySingleConnect != want.LegacySingleConnect {
		t.Errorf("LegacySingleConnect: wanted %t but got %t", want.LegacySingleConnect, have.LegacySingleConnect)
	}
	if have.PrivLevel != want.PrivLevel {
		t.Errorf("PrivLevel: wanted %d but got %d", want.PrivLevel, have.PrivLevel)
	}
	if have.Port != want.Port {
		t.Errorf("Port: wanted %q but got %q", want.Port, have.Port)
	}
	if have.Timeout != want.Timeout {
		t.Errorf("Timeout: wanted %v but got %v", want.Timeout, have.Timeout)
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
	assertModuleEquals(t, &m, &want)
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
