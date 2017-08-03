package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.build.ge.com/212601587/fhid/fhid"
	"github.build.ge.com/212601587/fhid/fhidConfig"
	"github.build.ge.com/212601587/fhid/fhidLogger"
)

var version string

func main() {
	// set up configuration
	var configFile string
	var logLevel string
	var logFile string
	var versionFlag bool
	var daemonFlag bool
	flag.StringVar(&configFile, "c", "./config.json", "Path to config file.")
	flag.StringVar(&logFile, "logfile", "fhid.log.json", "JSON logfile location")
	flag.StringVar(&logLevel, "loglevel", "info", "Log level (info or debug)")
	flag.BoolVar(&versionFlag, "version", false, "print version and exit")
	flag.BoolVar(&daemonFlag, "daemon", false, "run as daemon with no stdout")
	flag.Parse()
	// if daemon just log to file
	fhidLogger.SetLogger(daemonFlag, logFile, logLevel)
	if configFile == "" {
		fhidLogger.Loggo.Crit("Please specify config file with -c flag")
		os.Exit(1)
	}
	version = getVersion()
	if versionFlag {
		fmt.Printf("fhid %s\n", version)
		os.Exit(0)
	}
	err := fhidConfig.SetConfig(configFile)
	fhidConfig.Version = version
	fhidLogger.Loggo.Info("fhid: Fixham Harbour Image Database", "version", version)
	fhidLogger.Loggo.Info("Set fhidConfig.Version", "Version", fhidConfig.Version)
	if err != nil {
		fhidLogger.Loggo.Error("Error loading config file, check formatting.", "filename", configFile, "Error", err)
		os.Exit(1)
	}
	fhidLogger.Loggo.Info("Loaded config", "Config", fhidConfig.Config.ShowConfig())

	http.HandleFunc("/images", fhid.ImageWrite)
	http.HandleFunc("/healthcheck", fhid.HealthCheck)
	http.ListenAndServe(":"+fhidConfig.Config.ListenPort, nil)

}

func getVersion() string {
	if version == "" {
		version = "0.0"
	}
	return version
}
