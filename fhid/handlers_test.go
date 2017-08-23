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
	"github.com/alicebob/miniredis"
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

const imageGood2 = `
{
"Version":"3.4.3.99",
"BaseOS":"Centos7",
"ReleaseNotes":"Did the thing again"
}
`

const imageGoodExpected = `
{
"Results":[{
"ImageID":".*",
"Version":"1.2.3.145",
"BaseOS":"Ubuntu14.04",
"ReleaseNotes":"Did the thing"}]}`

const imageGoodNonMatcher = `
{
"Version":"9999999999",
"BaseOS":"Winders",
"ReleaseNotes":"bar foo"
}
`

const expectedResponseImageQueryReleaseNotes = `
{
"Results": [
{
"ImageID": ".*",
"CreateDate": ".*",
"Version":"1.2.3.145",
"BaseOS":"Ubuntu14.04",
"ReleaseNotes":"Did the thing"
},
{
"ImageID": ".*",
"CreateDate": ".*",
"Version":"3.4.3.99",
"BaseOS":"Centos7",
"ReleaseNotes":"Did the thing again"
}
]
}
`

const expectedResponseImageQueryVersion = `
{
"Results":[{
"ImageID":".*",
"CreateDate": ".*",
"Version":"1.2.3.145",
"BaseOS":"Ubuntu14.04",
"ReleaseNotes":"Did the thing"}]}
`

const expectedResponseImageQueryBaseOS = `
{
"Results":[{
"ImageID":".*",
"CreateDate": ".*",
"Version":"1.2.3.145",
"BaseOS":"Ubuntu14.04",
"ReleaseNotes":"Did the thing"}]}
`

const imageQueryVersion = `
{
	"Version": {"StringMatch": "1.2.3.145"}
}
`

const imageQueryReleaseNotes = `
{
	"ReleaseNotes": {"StringMatch": ".*Did.*"}
}
`

const imageQueryBaseOS = `
{
	"BaseOS": {"StringMatch": ".*Ubuntu.*"}
}
`

func resultsMatchExpected(results, expected string) (match bool, err error) {
	fhidLogger.Loggo.Info("Checking to see if results match expected.")
	// unmarshal actual results
	var iqrGot imageQueryResults
	bresults := []byte(results)
	err = json.Unmarshal(bresults, &iqrGot)
	if err != nil {
		fhidLogger.Loggo.Error("Error unmarshaling query results", "Error", err)
		return false, err
	}

	// unmarshal expected results
	var iqrWant imageQueryResults
	bexpected := []byte(expected)
	err = json.Unmarshal(bexpected, &iqrWant)
	if err != nil {
		fhidLogger.Loggo.Error("Error unmarshaling expected results", "Error", err)
		return false, err
	}

	// now loop through the expected and match one for one to results
	// except for ImageID since it's GUUID and changes every time
	match = true
	if len(iqrWant.Results) != len(iqrWant.Results) {
		return false, err
	}
	for idx, exp := range iqrWant.Results {
		switch {
		case exp.Version != iqrGot.Results[idx].Version:
			return false, err
		case exp.ReleaseNotes != iqrGot.Results[idx].ReleaseNotes:
			return false, err
		case exp.BaseOS != iqrGot.Results[idx].BaseOS:
			return false, err
		}
	}
	return match, err
}

func writeConfigFile() (*bytes.Buffer, error) {
	seed := `{
        "RedisEndpoint": "localhost:6379",
		"RedisImageIndexSet": "IMAGE_INDEX",
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

func setup(fake bool, addr string) error {
	filename := "handlers_test_config_temp.json"
	configbuff, err := writeConfigFile()
	err = writeBufferToFile(filename, configbuff)
	if err != nil {
		fhidLogger.Loggo.Error("Error writing temp config file", "Error", err)
		return err
	}
	err = fhidConfig.SetConfig(filename)
	if err != nil {
		fhidLogger.Loggo.Error("Error parsing config file", "Error", err)
		return err
	}
	if fake == true {
		fhidConfig.Config.RedisEndpoint = addr
	}
	fhidLogger.Loggo.Debug("Connecting to Redis at address.", "Address", fhidConfig.Config.RedisEndpoint)
	err = SetupConnection()
	if err != nil {
		fhidLogger.Loggo.Error("Error in Redis test connection", "Error", err)
		TeardownConnection()
		return err
	}
	defer deleteFile(filename)
	return err

}

func initLog() {
	fhidLogger.SetLogger(false, "fhid_test.log.json", "debug")
}

func runFakeRedis() (addr string, err error) {
	s, err := miniredis.Run()
	if err != nil {
		return "", err
	}
	addr = s.Addr()
	return addr, err
}

func seedQueryData() error {
	postBody := bytes.NewBufferString(imageGood)
	req, err := http.NewRequest("POST", "/images/?Score=0", postBody)
	if err != nil {
		return err
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImages)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		fhidLogger.Loggo.Error("handler returned wrong status code",
			"Got", status, "Want", http.StatusOK)
	}
	var j imagePostResponse
	err = json.Unmarshal([]byte(rr.Body.String()), &j)
	if err != nil {
		fhidLogger.Loggo.Error("Unable to unmarshal response JSON",
			"Error", err)
		return err
	}

	// write a second entry
	postBody = bytes.NewBufferString(imageGood2)
	req, err = http.NewRequest("POST", "/images/?Score=1", postBody)
	if err != nil {
		return err
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		fhidLogger.Loggo.Error("handler returned wrong status code",
			"Got", status, "Want", http.StatusOK)
	}

	// write a third entry that shouldn't match query
	postBody = bytes.NewBufferString(imageGoodNonMatcher)
	req, err = http.NewRequest("POST", "/images/?Score=2", postBody)
	if err != nil {
		return err
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		fhidLogger.Loggo.Error("handler returned wrong status code",
			"Got", status, "Want", http.StatusOK)
	}
	return err
}

func TestImageQueryReleaseNotes(t *testing.T) {
	initLog()
	// we initialize the fake redis instance
	addr, err := runFakeRedis()
	fhidLogger.Loggo.Info("Done starting fake Redis.")
	if err != nil {
		t.Errorf("Unable to start fake Redis for testing: %s", err)
	}
	err = setup(true, addr)
	if err != nil {
		t.Errorf("Unable to connect to fake Redis for testing: %s", err)
	}
	// First we need to post some entries to the DB
	// We have to use the score URL query so we force order or results
	// since the Redis index set we're using is a sorted set and only
	// sorts if proper score weights are given.
	err = seedQueryData()
	if err != nil {
		t.Errorf("Error seeding query data. '%s'", err)
	}
	// set up response recorder
	rr := httptest.NewRecorder()
	queryBody := bytes.NewBufferString(imageQueryReleaseNotes)
	req, err := http.NewRequest("POST", "/image_query", queryBody)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImagesQuery)
	handler.ServeHTTP(rr, req)
	// Check the response body is what we expect.
	expected := strings.Replace(expectedResponseImageQueryReleaseNotes, "\n", "", -1)
	match, err := resultsMatchExpected(rr.Body.String(), expected)
	if !match {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'",
			rr.Body.String(), expected)
	}
}

func TestImageQueryVersion(t *testing.T) {
	initLog()
	// we initialize the fake redis instance
	addr, err := runFakeRedis()
	fhidLogger.Loggo.Info("Done starting fake Redis.")
	if err != nil {
		t.Errorf("Unable to start fake Redis for testing: %s", err)
	}
	err = setup(true, addr)
	if err != nil {
		t.Errorf("Unable to connect to fake Redis for testing: %s", err)
	}
	// First we need to post some entries to the DB
	// We have to use the score URL query so we force order or results
	// since the Redis index set we're using is a sorted set and only
	// sorts if proper score weights are given.
	err = seedQueryData()
	if err != nil {
		t.Errorf("Error seeding query data. '%s'", err)
	}
	// set up response recorder
	rr := httptest.NewRecorder()
	queryBody := bytes.NewBufferString(imageQueryVersion)
	req, err := http.NewRequest("POST", "/image_query", queryBody)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImagesQuery)
	handler.ServeHTTP(rr, req)
	// Check the response body is what we expect.
	expected := strings.Replace(expectedResponseImageQueryVersion, "\n", "", -1)
	match, err := resultsMatchExpected(rr.Body.String(), expected)
	if !match {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'",
			rr.Body.String(), expected)
	}
}

func TestImageGetNone(t *testing.T) {
	err := setup(false, "")
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

func TestImagePostGetReal(t *testing.T) {
	initLog()
	err := setup(false, "")
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

	// Check the response body is what we expect.
	expected := strings.Replace(imageGoodExpected, "\n", "", -1)
	match, err := resultsMatchExpected(rr.Body.String(), expected)
	if !match {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestImageQueryBaseOS(t *testing.T) {
	initLog()
	// we initialize the fake redis instance
	addr, err := runFakeRedis()
	fhidLogger.Loggo.Info("Done starting fake Redis.")
	if err != nil {
		t.Errorf("Unable to start fake Redis for testing: %s", err)
	}
	err = setup(true, addr)
	if err != nil {
		t.Errorf("Unable to connect to fake Redis for testing: %s", err)
	}
	// First we need to post some entries to the DB
	// We have to use the score URL query so we force order or results
	// since the Redis index set we're using is a sorted set and only
	// sorts if proper score weights are given.
	err = seedQueryData()
	if err != nil {
		t.Errorf("Error seeding query data. '%s'", err)
	}
	// set up response recorder
	rr := httptest.NewRecorder()
	queryBody := bytes.NewBufferString(imageQueryBaseOS)
	req, err := http.NewRequest("POST", "/image_query", queryBody)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler := http.HandlerFunc(HandlerImagesQuery)
	handler.ServeHTTP(rr, req)
	// Check the response body is what we expect.
	expected := strings.Replace(expectedResponseImageQueryBaseOS, "\n", "", -1)
	match, err := resultsMatchExpected(rr.Body.String(), expected)
	if !match {
		t.Errorf("handler returned unexpected body: got '%v' want '%v'",
			rr.Body.String(), expected)
	}
}

func TestImagePostGetFake(t *testing.T) {
	initLog()
	addr, err := runFakeRedis()
	fhidLogger.Loggo.Info("Done starting fake Redis.")
	if err != nil {
		t.Errorf("Unable to start fake Redis for testing: %s", err)
	}
	err = setup(true, addr)
	if err != nil {
		t.Errorf("Unable to connect to fake Redis for testing: %s", err)
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

	// Check the response body is what we expect.
	expected := strings.Replace(imageGoodExpected, "\n", "", -1)
	match, err := resultsMatchExpected(rr.Body.String(), expected)
	if !match {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
