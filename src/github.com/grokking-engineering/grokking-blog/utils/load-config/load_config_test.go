package loadConfig

import (
	"os"
	"testing"
)

type TestConfig struct {
	Foo struct {
		Bar string `config:"BAR"`
		Baz struct {
			Quix string `config:"QUIX"`
		} `config:"baz"`
	} `config:"foo"`
}

func TestFromEnv(T *testing.T) {
	testEnv := map[string]string{
		"BAR":  "xbar",
		"QUIX": "xquix",
	}
	for k, v := range testEnv {
		err := os.Setenv(k, v)
		if err != nil {
			panic(err)
		}
	}

	var config TestConfig
	fromEnv(map[string]interface{}{}, &config, "config")

	if config.Foo.Bar != "xbar" || config.Foo.Baz.Quix != "xquix" {
		T.Error("Expect config", config)
	}
}
