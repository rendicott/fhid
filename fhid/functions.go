package fhid

import (
	"encoding/json"
	"regexp"

	"github.build.ge.com/212601587/fhid/fhidConfig"
	fi "github.build.ge.com/212601587/fhid/fhidLogger"
)

type imageQuerySub struct {
	StringMatch string
	Function    string
	Value       string // e.g., 'latest' or '.*'
}

func newImageQuerySub() *imageQuerySub {
	iqs := imageQuerySub{}
	iqs.StringMatch = ""
	iqs.Function = ""
	iqs.Value = ""
	return &iqs
}

type imageQuery struct {
	Version      *imageQuerySub
	BaseOS       *imageQuerySub
	BuildNotes   *imageQuerySub
	ReleaseNotes *imageQuerySub
}

// newImageQuery instantiates and returns a blank imageQuery so that
// unset fields can be queried assuming default values.
func newImageQuery() imageQuery {
	iq := imageQuery{}
	iq.Version = newImageQuerySub()
	iq.BaseOS = newImageQuerySub()
	iq.BuildNotes = newImageQuerySub()
	iq.ReleaseNotes = newImageQuerySub()
	return iq
}

func (iq *imageQuery) processBody(rbody []byte) error {
	err := json.Unmarshal(rbody, &iq)
	fi.Loggo.Info("Processed query", "imageQuery.BuildNotes.StringMatch", iq.BuildNotes.StringMatch)
	return err
}

// imageQuery loops through the query properties and tries to detect
// which type of query search to run then executes and returns
// true if the search matches the given buildEntry
func (iq *imageQuery) search(ie *buildEntry) (match bool, err error) {
	switch {
	case iq.Version.StringMatch != "":
		fi.Loggo.Debug("Detected StringMatch on Version")
		match, err = iq.stringMatch(ie.Version, iq.Version.StringMatch)
	case iq.BaseOS.StringMatch != "":
		fi.Loggo.Debug("Detected StringMatch on BaseOS")
		match, err = iq.stringMatch(ie.BaseOS, iq.BaseOS.StringMatch)
	case iq.ReleaseNotes.StringMatch != "":
		fi.Loggo.Debug("Detected StringMatch on ReleaseNotes")
		rnb, err := json.Marshal(ie.ReleaseNotes)
		if err != nil {
			return match, err
		}
		match, err = iq.stringMatch(string(rnb), iq.ReleaseNotes.StringMatch)
	case iq.BuildNotes.StringMatch != "":
		fi.Loggo.Debug("Detected StringMatch on BuildNotes")
		rnb, err := json.Marshal(ie.BuildNotes)
		if err != nil {
			return match, err
		}
		match, err = iq.stringMatch(string(rnb), iq.BuildNotes.StringMatch)
	default:
		fi.Loggo.Info("No queries could be parsed.")
	}
	return match, err
}

func (iq *imageQuery) execute() (sresults string, err error) {
	var qresults []buildEntry
	fi.Loggo.Info("Executing query...")
	results, err := Rmembers(fhidConfig.Config.RedisImageIndexSet)
	if err != nil {
		fi.Loggo.Error("Error in getting index set", "Error", err)
		return sresults, err
	}
	fi.Loggo.Debug("Got set", "Set", fhidConfig.Config.RedisImageIndexSet, "Value", results)
	for _, key := range results {
		val, err := Rget(key)
		if err != nil {
			fi.Loggo.Error("Error retreiving key.", "Error", err, "Key", key)
		}
		fi.Loggo.Debug("Got value", "Value", val)
		var ie buildEntry
		err = json.Unmarshal([]byte(val), &ie)
		if err != nil {
			fi.Loggo.Error("Error unmarshaling retrieved value.", "Error", err)
		}
		match, err := iq.search(&ie)
		if err != nil {
			fi.Loggo.Error("Error search val for match", "Error", err)
		}
		if match == true {
			qresults = append(qresults, ie)
		}
	}
	fi.Loggo.Info("Query returned no errors.", "NumberOfResults", len(qresults))
	var iqr imageQueryResults
	iqr.Results = qresults
	bsresults, err := json.MarshalIndent(iqr, "", "    ")
	return string(bsresults), err
}

func (iq *imageQuery) stringMatch(value, reg string) (bool, error) {
	matched, err := regexp.MatchString(reg, value)
	return matched, err
}
