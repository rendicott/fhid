package fhidConfig

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is the exported configuration
// that other packages can use during
// runtime
var Config *Configuration

// Globally accessible version.
var Version string

// Configuration is a struct used
// to build the exported Config variable
type Configuration struct {
	RedisEndpoint      string
	RedisImageIndexSet string
	ListenPort         string
	ListenHost         string
}

// ShowConfig returns a string of log formatted
// config for debug purposes
func (*Configuration) ShowConfig() string {
	msg := fmt.Sprintf("\n")
	msg += fmt.Sprintf("CONFIGURATION: RedisEndpoint = '%s'\n", Config.RedisEndpoint)
	msg += fmt.Sprintf("CONFIGURATION: ListenPort = '%s'\n", Config.ListenPort)
	msg += fmt.Sprintf("CONFIGURATION: ListenHost = '%s'\n", Config.ListenHost)
	msg += fmt.Sprintf("CONFIGURATION: RedisImageIndexSet = '%s'\n", Config.RedisImageIndexSet)
	return msg
}

// SetConfig parses a config json file and returns
// and sets a package exported configuration object
// for use within other packages
func SetConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		return err
	}
	return err
}
