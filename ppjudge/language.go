package main

import (
	"io"
	"io/ioutil"

	"path/filepath"

	"errors"

	"encoding/json"

	"fmt"

	"gopkg.in/yaml.v2"
)

type LanguageConfiguration map[int64]struct {
	CompileImage   string   `yaml:"compile_image" json:"compile_image"`
	ExecImage      string   `yaml:"exec_image" json:"exec_image"`
	CompileCommand []string `yaml:"compile_command" json:"compile_command"`
	ExecCommand    []string `yaml:"exec_command" json:"exec_command"`
}

func LoadLanguageConfiguration(path string) (LanguageConfiguration, error) {
	var lc LanguageConfiguration

	b, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	switch filepath.Ext(path) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(b, lc); err != nil {
			return nil, err
		}
	case "json":
		if err := json.Unmarshal(b, lc); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Unsupported file type")
	}

	return lc, nil
}

func EchoLanguageConfigurationTemplate(w io.Writer, fileType string) bool {
	temp := LanguageConfiguration{
		10: {
			ExecImage:      "popcon-cpp",
			CompileImage:   "popcon-cpp",
			CompileCommand: []string{"g++", "-O2", "-o", "/work/a.out", "-std=c++14", "/work/main.cpp"},
			ExecCommand:    []string{"/work/a.out"},
		},
	}
	fmt.Fprintln(w, "NOTE: The key is lid(language id) on the database of popcon-sc.")
	switch fileType {
	case "none:":
		return false
	case "json":
		json.NewEncoder(w).Encode(temp)
	case "yaml", "yml":
		b, _ := yaml.Marshal(temp)
		w.Write(b)
	}

	return true
}
