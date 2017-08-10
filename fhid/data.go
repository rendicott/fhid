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

// Rconn is the package level redis connection
var Rconn ResourceConn

// ResourceConn adapts a Redigo connection to a Vitess Resource.
type ResourceConn struct {
	redis.Conn
}

// Close should close the redis connection
func (r ResourceConn) Close() {
	r.Conn.Close()
}

// ImageEntry holds the structure of the image
// entry to push and pull to the database.
type imageEntry struct {
	ImageID      string
	Version      string
	BaseOS       string
	ReleaseNotes string
}

// ParseBodyWrite is the method to parse the body of the ImageEntry object from
// the web request.
func (i *imageEntry) ParseBodyWrite(rbody []byte) (key string, err error) {
	fhidLogger.Loggo.Info("Processing image body request", "Body", string(rbody))
	err = json.Unmarshal(rbody, i)
	if err != nil {
		return "", err
	}
	key = getUUID()
	i.ImageID = key
	srep, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	err = Rset(key, string(srep))
	return key, err
}

// Query takes properties of self and uses it as a
// search query. Supports regex strings as the values.
func (i *imageEntry) Query() error {
	fhidLogger.Loggo.Info("Processing image query...")
	var err error
	return err
}

// Rget returns the value of keyname.
func Rget(keyname string) (value string, err error) {
	value, err = redis.String(Rconn.Do("GET", keyname))
	if err == nil {
		fhidLogger.Loggo.Info("Retrieved entry successfully", "KeyName", keyname, "Value", value)
		return value, err
	}
	fhidLogger.Loggo.Error("Error retrieving Redis data", "Error", err)
	return "", err
}

// Rset sets the value of keyname to value.
func Rset(keyname, value string) error {
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

// SetupConnection tests the connection to the Redis datalayer
func SetupConnection() error {
	p := pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", ":6379")
		if err != nil {
			fhidLogger.Loggo.Crit("Error connecting to Redis.", "Error", err)
			os.Exit(1)
		}
		return ResourceConn{c}, err
	}, 1, 2, time.Minute)
	ctx := context.TODO()
	r, err := p.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	Rconn = r.(ResourceConn)
	return err
}

// TeardownConnection closes the connection to Redis
func TeardownConnection() {
	Rconn.Close()
}
