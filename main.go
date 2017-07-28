package main

import (
	"flag"
	"log"
	"os"

	"github.build.ge.com/212601587/fhid/fhid"
	"github.build.ge.com/212601587/fhid/fhidConfig"
)

var version string

func main() {
	// set up configuration
	var configFile string
	var versionFlag bool
	flag.StringVar(&configFile, "c", "./config.json", "Path to config file.")
	flag.BoolVar(&versionFlag, "version", false, "print version and exit")
	flag.Parse()
	if versionFlag {
		log.Printf("fhid %s\n", version)
		os.Exit(0)
	}
	err := fhidConfig.SetConfig(configFile)
	if err != nil {
		log.Printf("Error loading config file '%s', check formatting: '%s'", configFile, err)
		os.Exit(1)
	}
	log.Printf(fhidConfig.Config.ShowConfig())
	fhid.Test()
}
