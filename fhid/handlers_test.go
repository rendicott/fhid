package fhid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.build.ge.com/212601587/fhid/fhidConfig"
	"github.build.ge.com/212601587/fhid/fhidLogger"
)

type imagePostResponse struct {
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

const imageQuery1 = `
{
	"Version": {"StringMatch": "1.2.3.145"}
}
`

const imageQuery3 = `
{
	"BaseOS": {"StringMatch": ".*Ubuntu.*"}
}
`

func writeConfigFile() (*bytes.Buffer, error) {
	seed := `{
        "RedisEndpoint": "localhost:6379",
        "ListenPort": "8090",
		"ListenHost": "127.0.0.1"
}`
	var b bytes.Buffer
	_, err := b.WriteString(seed)
	if err != nil {
		fmt.Println("error: ", err)
		return &b, err
	}
	return &b, err
}

func writeBufferToFile(filename string, b *bytes.Buffer) error {
	err := ioutil.WriteFile(filename, b.Bytes(), 0644)
	return err
}

func deleteFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		panic(err)
	}
}

func setup() error {
	filename := "handlers_test_config_temp.json"
	configbuff, err := writeConfigFile()
	err = writeBufferToFile(filename, configbuff)
	if err != nil {
		fhidLogger.Loggo.Error("Error writing temp config file", "Error", err)
		return err
	}
	defer deleteFile(filename)
	err = fhidConfig.SetConfig(filename)
	err = SetupConnection()
	if err != nil {
		fhidLogger.Loggo.Error("Error in Redis test connection", "Error", err)
		TeardownConnection()
		return err
	}
	return err

}

func initLog() {
	fhidLogger.SetLogger(false, "fhid_test.log.json", "debug")
}

func TestImageQuery(t *testing.T) {
	err := setup()
	initLog()
	if err != nil {
		t.Errorf("Unable to connect to Redis for testing: %s", err)
	}

	queryBody := bytes.NewBufferString(imageQuery1)

	req, err := http.NewRequest("POST", "/image_query", queryBody)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImagesQuery)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

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
	expected := `{"Error": "Key 'ImageID' not found in URL string."}` + "\n"
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
	var j imagePostResponse
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		t.Errorf("Unable to unmarshal response JSON: %s", err)
	}
	fmt.Printf("Parsed j.Data into '%s'", j.Data)

	// now we retrieve the entry
	uriQuery := fmt.Sprintf("/images?ImageID=%s", j.Data)
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
	imageGoodResponse := `{`
	imageGoodResponse += `"ImageID":"%s",`
	imageGoodResponse += `"Version":"1.2.3.145",`
	imageGoodResponse += `"BaseOS":"Ubuntu14.04",`
	imageGoodResponse += `"ReleaseNotes":"Did the thing"}`
	imageGoodResponse = fmt.Sprintf(imageGoodResponse, j.Data)

	// Check the response body is what we expect.
	expected := strings.Replace(imageGoodResponse, "\n", "", -1)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
