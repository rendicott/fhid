package fhid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.build.ge.com/212601587/fhid/fhidLogger"

	"github.build.ge.com/212601587/fhid/fhidConfig"
)

// Status is an object to hold system status
// to be returned by things like the healthcheck handler
var Status *status

func init() {
	Status = &status{}
	Status.State = "Healthy"
	// Status.Version = &fhidConfig.Config.Version
	Status.Version = fhidConfig.Version
}

type status struct {
	State   string `json:"State"`
	Version string `json:"Version"`
}

// GetStatus returns the current string export of the
// FhidStatus struct.
func (f *status) GetStatus() (msg string) {
	b, err := json.Marshal(f)
	if err != nil {
		msg = fmt.Sprintf(`{"State":"Unhealthy","Version":"%s"}`, fhidConfig.Version)
	}
	msg = string(b)
	return msg
}

// ImageWrite handles the post to the database
func ImageWrite(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fhidLogger.Loggo.Info("Error reading body", "Error", err)
			http.Error(w, `{"Error": "Error reading body."}`, http.StatusBadRequest)
			return
		}
		image := ImageEntry{}
		err = image.ParseBody(body)
		if err != nil {
			http.Error(w, messageErrorHandler(err), http.StatusInternalServerError)
		}
	case "PUT":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	case "DELETE":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	default:
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	}
}

func messageErrorHandler(err error) string {
	msg := fmt.Sprintf(`{"Msg":"Internal Server Error","Error":"%v"}`, err)
	fhidLogger.Loggo.Error("Internal Server Error", "Error", err)
	return msg
}

func messageMethodNotAllowed() string {
	return `{"Error":"Method not allowed"}`
}

// HealthCheck is a health check handler.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	msg := Status.GetStatus()
	fmt.Fprintf(w, msg)
}
