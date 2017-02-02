package config

import (
	"github.com/jacobstr/confer"
	"os"
	"fmt"
	"path"
	"runtime"
	"strings"
)

func init() {
	LoadConfig()
}

var App *confer.ConfigManager

// So we don't overwrite our existing configs
var loaded = false

// Load our configuration if it hasn't already been loaded.
func LoadConfig() {
	if loaded == false {
		load()
	}

	loaded = true
}

// Load configuration data.
func load() (*confer.ConfigManager) {
	// Load the config
	App = confer.NewConfiguration()
	App.SetRootPath(getLoadPath())
	if errs := App.ReadPaths(getConfigPaths()...); errs != nil {
		fmt.Println(errs)
	}
	return App
}

// Determines the root path for our configuration data.
func getLoadPath() string {
	// Some magic to get the abs path of the file
	_, filename, _, _ := runtime.Caller(1)
	baseDir := strings.Join([]string{path.Dir(filename), "yml"}, "/")
	return baseDir
}

// Determines the applicable set of config files to load based on our
// current environment.
func getConfigPaths() []string {
	bursaEnv := os.Getenv("BURSA_ENV");

	var paths []string

	paths = append(paths, "server.yml")

	// Server configuration
	if (bursaEnv != "") {
		paths = append(paths, fmt.Sprintf("server.%s.yml", bursaEnv))
	} else {
		paths = append(paths, "server.development.yml")
	}

	// Database configuration
	return append(paths, "database.yml")
}
