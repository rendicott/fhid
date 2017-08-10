package fhid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.build.ge.com/212601587/fhid/fhidLogger"
)

type ImagePostResponse struct {
	Success string
	Data    string
}

const imageGood = `
{
"Version":"1.2.3.145",
"BaseOS":"Ubuntu14.04",
"ReleaseNotes":"Did the thing"
}
`

func setup() error {
	err := SetupConnection()
	if err != nil {
		fhidLogger.Loggo.Error("Error in Redis test connection", "Error", err)
		TeardownConnection()
		os.Exit(1)
	}
	return err

}

func initLog() {
	fhidLogger.SetLogger(false, "fhid_test.log.json", "info")
}

func TestImageGetNone(t *testing.T) {
	err := setup()
	initLog()
	if err != nil {
		t.Errorf("Unable to connect to Redis for testing: %s", err)
	}

	req, err := http.NewRequest("GET", "/images", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImages)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
	// Check the response body is what we expect.
	expected := `{"Error": "Key 'ImageId' not found in URL string."}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestImagePostGet(t *testing.T) {
	err := setup()
	initLog()
	if err != nil {
		t.Errorf("Unable to connect to Redis for testing: %s", err)
	}

	// First we need to post an entry to the DB
	postBody := bytes.NewBufferString(imageGood)
	req, err := http.NewRequest("POST", "/images", postBody)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImages)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var j ImagePostResponse
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Errorf("Unable to unmarshal response JSON: %s", err)
	}
	fmt.Printf("Parsed j.Data into '%s'", j.Data)

	// now we retrieve the entry
	uriQuery := fmt.Sprintf("/images?ImageId=%s", j.Data)
	req, err = http.NewRequest("GET", uriQuery, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := strings.Replace(imageGood, "\n", "", -1)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
