package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type LanguageConfiguration map[int64]struct {
	CompileImage   string   `yaml:"compile_image" json:"compile_image"`
	ExecImage      string   `yaml:"exec_image" json:"exec_image"`
	CompileCommand []string `yaml:"compile_command" json:"compile_command"`
	ExecCommand    []string `yaml:"exec_command" json:"exec_command"`
	SourceFileName string   `yaml:"source_file_name" json:"source_file_name"`
}

func LoadLanguageConfiguration(path string) (LanguageConfiguration, error) {
	var lc LanguageConfiguration

	b, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	switch filepath.Ext(path) {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(b, &lc); err != nil {
			return nil, err
		}
	case ".json":
		if err := json.Unmarshal(b, &lc); err != nil {
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
			SourceFileName: "main.cpp",
		},
	}

	switch fileType {
	case "json":
		json.NewEncoder(w).Encode(temp)
	case "yaml", "yml":
		b, _ := yaml.Marshal(temp)
		w.Write(b)
	default:
		return false
	}
	fmt.Fprintln(w, "NOTE: The key is lid(language id) on the database of popcon-sc.")
	fmt.Fprintln(w, "NOTE: If compilation is not neccesary, make compile_image empty.")

	return true
}
