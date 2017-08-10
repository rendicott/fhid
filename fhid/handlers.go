package fhid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

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

// getMapKey searches for a key in a map.
func getMapKey(m map[string]string, key string) (value string, err error) {
	if x, found := m[key]; found {
		return x, err
	}
	// if we made it here the key doesn't exist
	return "", errors.New("key not found")
}

// HandlerImagesQuery handles posted queries to search
// for images.
func HandlerImagesQuery(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	case "POST":
		fhidLogger.Loggo.Info("ImageQuery request")
		fhidLogger.Loggo.Debug("ImageQuery Body captured", "Body", r.Body)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fhidLogger.Loggo.Crit("Error processing body", "Error", err)
		} else {
			var query imageQuery
			err = query.processBody(body)
			results, err := query.execute()
			if err != nil {
				http.Error(w, messageErrorHandlerQuery(err), http.StatusInternalServerError)
			} else {
				fhidLogger.Loggo.Debug("Got query results", "Results", results)
			}
		}

	case "DELETE":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	case "PUT":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	default:
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	}
}

// HandlerImages handles the post to the database
func HandlerImages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fhidLogger.Loggo.Info("Request URL captured", "URL", r.URL)
		u, err := url.Parse(r.URL.String())
		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			fhidLogger.Loggo.Error("Error processing URL", "Error", err)
			http.Error(w, `{"Error": "Error processing URL"}`, http.StatusBadRequest)
		}
		fhidLogger.Loggo.Debug("Parsed URL query successfully", "Query", q)
		key := "ImageID"
		value, ok := q[key]
		if !ok {
			fhidLogger.Loggo.Info("Key not found in URL string", "Key", key)
		}
		fhidLogger.Loggo.Debug("Parsed ImageID", "ImageID", value)
		if len(value) < 1 {
			msg := fmt.Sprintf(`{"Error": "Key '%s' not found in URL string."}`, key)
			http.Error(w, msg, http.StatusBadRequest)
		} else {
			data, err := Rget(value[0])
			if err != nil {
				msg := fmt.Sprintf(`{"Error": "Error fullfilling request for '%s': '%s'"}`, value, err)
				http.Error(w, msg, http.StatusBadRequest)
			} else {
				fhidLogger.Loggo.Debug("Retrieved data successfully", "Data", data)
				fmt.Fprintf(w, data)
			}
		}

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fhidLogger.Loggo.Info("Error reading body", "Error", err)
			http.Error(w, `{"Error": "Error reading body."}`, http.StatusBadRequest)
			return
		}
		image := imageEntry{}
		key, err := image.ParseBodyWrite(body)
		if err != nil {
			http.Error(w, messageErrorHandler(err), http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, messageSuccessData(key))

		}
	case "PUT":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	case "DELETE":
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	default:
		http.Error(w, messageMethodNotAllowed(), http.StatusMethodNotAllowed)
	}
}

func messageInvalidRequest(err error) string {
	msg := fmt.Sprintf(`{"Msg":"Invalid Request","Error":"%v"}`, err)
	fhidLogger.Loggo.Error("Invalid Request", "Error", err)
	return msg
}

func messageErrorHandler(err error) string {
	msg := fmt.Sprintf(`{"Msg":"Internal Server Error","Error":"%v"}`, err)
	fhidLogger.Loggo.Error("Internal Server Error", "Error", err)
	return msg
}

func messageErrorHandlerQuery(err error) string {
	msg := fmt.Sprintf(`{"Msg":"Query failed.","Error":"%v"}`, err)
	fhidLogger.Loggo.Error("Query failed", "Error", err)
	return msg
}

func messageSuccess() string {
	return `{"Success": "True"}`
}

func messageSuccessData(s string) string {
	return fmt.Sprintf(`{"Success": "True", "Data": "%s"}`, s)
}

func messageMethodNotAllowed() string {
	return `{"Error":"Method not allowed"}`
}

// HealthCheck is a health check handler.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	msg := Status.GetStatus()
	fmt.Fprintf(w, msg)
}
