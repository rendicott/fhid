package main

import (
	"flag"
	"log"
	"net/http"
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
	version = getVersion()
	if versionFlag {
		log.Printf("fhid %s\n", version)
		os.Exit(0)
	}
	err := fhidConfig.SetConfig(configFile, version)
	if err != nil {
		log.Printf("Error loading config file '%s', check formatting: '%s'", configFile, err)
		os.Exit(1)
	}
	log.Printf(fhidConfig.Config.ShowConfig())
	fhid.Test()

	http.HandleFunc("/", fhid.Handler)
	http.HandleFunc("/healthcheck", fhid.HealthCheck)
	http.ListenAndServe(":"+fhidConfig.Config.ListenPort, nil)

}

func getVersion() string {
	if version == "" {
		version = "0.0"
	}
	return version
}
