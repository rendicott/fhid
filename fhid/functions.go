package fhid

import "encoding/json"
import "github.build.ge.com/212601587/fhid/fhidLogger"

type imageQuerySub struct {
	StringMatch string
	Function    string
	Value       string // e.g., 'latest' or '.*'
}

type imageQuery struct {
	Version      *imageQuerySub
	BaseOS       *imageQuerySub
	ReleaseNotes *imageQuerySub
}

func (iq *imageQuery) processBody(rbody []byte) error {
	err := json.Unmarshal(rbody, &iq)
	return err
}

func (iq *imageQuery) execute() (qresults []imageEntry, err error) {
	fhidLogger.Loggo.Info("Executing query...")
	return qresults, err
}
