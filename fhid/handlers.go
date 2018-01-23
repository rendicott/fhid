package fhid

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.build.ge.com/212601587/fhid/fhidLogger"

	"github.build.ge.com/212601587/fhid/fhidConfig"
)

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
			query := newImageQuery()
			err = query.processBody(body)
			results, err := query.execute()
			if err != nil {
				http.Error(w, messageErrorHandlerQuery(err), http.StatusInternalServerError)
			} else {
				fmt.Fprintf(w, results)
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
			}
			var iqr imageQueryResults
			var ie imageEntry
			err = json.Unmarshal([]byte(data), &ie)
			if err != nil {
				msg := fmt.Sprintf(`{"Error": "Error processing object retrieved from database. %s}`, err)
				http.Error(w, msg, http.StatusBadRequest)
			}
			iqr.Results = append(iqr.Results, ie)
			rdata, err := json.MarshalIndent(&iqr, "", "    ")
			if err != nil {
				msg := fmt.Sprintf(`{"Error": "Error processing objects retrieved from database. %s}`, err)
				http.Error(w, msg, http.StatusBadRequest)
			}
			fhidLogger.Loggo.Debug("Retrieved data successfully", "Data", string(rdata))
			fmt.Fprintf(w, string(rdata))
		}

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fhidLogger.Loggo.Error("Error reading body", "Error", err)
			msg := fmt.Sprintf(`{"Success": "False", "Data": "%s", "Error": "Error reading body."}`, err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		// have to parse in a score in case we want to override ording in testing
		u, err := url.Parse(r.URL.String())
		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			fhidLogger.Loggo.Error("Error processing URL", "Error", err)
			msg := fmt.Sprintf(`{"Success": "False", "Data": "%s", "Error": "Error processing URL"}`, err)
			http.Error(w, msg, http.StatusBadRequest)
		}
		fhidLogger.Loggo.Debug("Parsed URL query successfully", "Query", q)
		score := 0
		key := "Score"
		value, ok := q[key]
		if ok {
			fhidLogger.Loggo.Info("Found score override in url query", "Key", key)
			fhidLogger.Loggo.Debug("Parsed Score", "Score", value)
			score, err = strconv.Atoi(value[0])
			if err != nil {
				fhidLogger.Loggo.Error("Error parsing score overide, defaulting to zero", "Error", err)
			}
		} else {
			score = 0
		}
		image := imageEntry{}
		key, err = image.ParseBodyWrite(body, score)
		if err != nil {
			fhidLogger.Loggo.Error("Error writing to database", "Error", err)
			msg := fmt.Sprintf(`{"Success": "False", "Data": "%s", "Error": "Error in body parse and post."}`, err)
			http.Error(w, msg, http.StatusInternalServerError)
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

// HealthCheck is a health check handler.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := &status{}
	status.State = "Healthy"
	// Status.Version = &fhidConfig.Config.Version
	status.Version = fhidConfig.Version
	msg := status.getStatus()
	fmt.Fprintf(w, msg)
}
