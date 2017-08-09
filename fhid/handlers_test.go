package fhid

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.build.ge.com/212601587/fhid/fhidLogger"
)

const imageGood = `
{
"Version": "1.2.3.145",
"BaseOS": "Ubuntu14.04",
"ReleaseNotes": "Did the thing"
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
	fhidLogger.SetLogger(false, "fhid_test.log.json", "debug")
}

func TestImagePost(t *testing.T) {
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
	/*
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
		// Check the status code is what we expect.
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// Check the response body is what we expect.
		expected := `{"alive": true}`
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}
	*/
}
