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

// Configuration is a struct used
// to build the exported Config variable
type Configuration struct {
	RedisEndpoint string
	ListenPort    string
	Version       string
}

// ShowConfig returns a string of log formatted
// config for debug purposes
func (*Configuration) ShowConfig() string {
	msg := fmt.Sprintf("\n")
	msg += fmt.Sprintf("CONFIGURATION: RedisEndpoint = '%s'\n", Config.RedisEndpoint)
	msg += fmt.Sprintf("CONFIGURATION: ListenPort = '%s'\n", Config.ListenPort)
	return msg
}

// SetVersion sets version for global reference
func (c *Configuration) SetVersion(v string) bool {
	c.Version = v
	return true
}

// GetVersion gets version for global reference
func (c *Configuration) GetVersion() string {
	return c.Version
}

// SetConfig parses a config json file and returns
// and sets a package exported configuration object
// for use within other packages
func SetConfig(filename, version string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	Config.SetVersion(version)
	if err != nil {
		return err
	}
	return err
}
