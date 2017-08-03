package fhid

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"

	"github.build.ge.com/212601587/fhid/fhidLogger"
)

var Rconn ResourceConn

// ResourceConn adapts a Redigo connection to a Vitess Resource.
type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

// ImageEntry holds the structure of the image
// entry to push and pull to the database.
type ImageEntry struct {
	Version      string
	BaseOS       string
	ReleaseNotes string
}

// ParseBody is the method to parse the body of the ImageEntry object from
// the web request.
func (i *ImageEntry) ParseBody(rbody []byte) (err error) {
	fhidLogger.Loggo.Info("Processing image body request", "Body", string(rbody))
	err = json.Unmarshal(rbody, i)
	if err != nil {
		return err
	}

	srep, err := json.Marshal(i)
	if err != nil {
		return err
	}
	err = Rset(getUUID(), string(srep))
	return err
}

// Rset sets the value of keyname to value.
func Rset(keyname, value string) error {
	p := pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", ":6379")
		if err != nil {
			fhidLogger.Loggo.Crit("Error connecting to Redis.", "Error", err)
			os.Exit(1)
		}
		return ResourceConn{c}, err
	}, 1, 2, time.Minute)
	defer p.Close()
	ctx := context.TODO()
	r, err := p.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Put(r)
	Rconn = r.(ResourceConn)
	n, err := Rconn.Do("SET", keyname, value)
	if err == nil {
		fhidLogger.Loggo.Info("Wrote entry successfully", "KeyName", keyname, "Value", n)
	} else {
		fhidLogger.Loggo.Error("Error writing Redis data", "Error", Rconn.Err())
	}
	return err
}

func getUUID() string {
	return uuid.NewV4().String()
}
