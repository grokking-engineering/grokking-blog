package loadConfig

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/grokking-engineering/grokking-blog/utils/logs"
)

var (
	l = logs.New("load-config")
)

func FromFileAndEnv(cfg interface{}, configPath string) error {
	err := FromFile(cfg, configPath)
	if err != nil {
		return err
	}

	envMap := make(map[string]interface{})
	fromEnv(envMap, cfg, "json")
	l.WithFields(envMap).Info("Environment config")
	return nil
}

func FromFile(cfg interface{}, configPath string) error {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}

	l.WithFields(logs.M{
		"path": absPath,
	}).Info("Load config from file")

	data, err := ioutil.ReadFile(absPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, cfg)
	if err != nil {
		return err
	}

	return nil
}

func fromEnv(envMap map[string]interface{}, v interface{}, tagName string) {
	vConfig := reflect.ValueOf(v)
	vConfig = reflect.Indirect(vConfig)
	if vConfig.Kind() != reflect.Struct {
		l.Fatal("Config must be a struct")
	}

	tConfig := vConfig.Type()
	for n, i := vConfig.NumField(), 0; i < n; i++ {
		vField := vConfig.Field(i)
		tField := tConfig.Field(i)

		tag := tField.Tag.Get(tagName)
		if tag == "" || strings.HasPrefix(tag, "-") {
			continue
		}

		if vField.Kind() == reflect.Struct {
			fromEnv(envMap, vField.Addr().Interface(), tagName)
			continue
		}

		if vField.Kind() != reflect.String {
			l.WithFields(nil).Fatalf("Field %v must be a string", tField.Name)
		}

		env := os.Getenv(tag)
		if env != "" {
			envMap[tag] = env
			vField.SetString(env)
		}
	}
}
