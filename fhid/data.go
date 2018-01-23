package fhid

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.build.ge.com/212601587/fhid/fhidConfig"

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

// amiEntry just holds basic structure of an AMI ID
// and an AMI region.
type amiEntry struct {
	AmiID     string
	AmiRegion string
}

// tags is a struct for holding AMI tags
type tags struct {
	Name  string
	Value string
}

// releaseNotes holds specific structure for packer
// aws builds
type releaseNotes struct {
	BuildLog   []string
	OutputAmis []*amiEntry
	Tags       []*tags
}

// ImageEntry holds the structure of the image
// entry to push and pull to the database.
type imageEntry struct {
	ImageID      string
	Version      string
	BaseOS       string
	ReleaseNotes *releaseNotes
	CreateDate   string
}

type imageQueryResults struct {
	Results []imageEntry
}

// ParseBodyWrite is the method to parse the body of the ImageEntry object from
// the web request.
func (i *imageEntry) ParseBodyWrite(rbody []byte, score int) (key string, err error) {
	fhidLogger.Loggo.Info("Processing image body request", "Body", string(rbody))
	err = json.Unmarshal(rbody, i)
	if err != nil {
		return "", err
	}
	t := time.Now()
	tstring := t.Format("2006-01-02 15:04:05")
	key = getUUID()
	i.ImageID = key
	i.CreateDate = tstring
	srep, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return "", err
	}

	err = Rset(key, string(srep), score)
	return key, err
}

// Rget returns the value of keyname.
func Rget(keyname string) (value string, err error) {
	value, err = redis.String(Rconn.Do("GET", keyname))
	if err == nil {
		fhidLogger.Loggo.Debug("Retrieved entry successfully", "KeyName", keyname, "Value", value)
		return value, err
	}
	fhidLogger.Loggo.Error("Error retrieving Redis data", "Error", err)
	return "", err
}

// Rmembers gets members of a set and returns the []string
func Rmembers(setName string) (results []string, err error) {
	n, err := redis.Strings(Rconn.Do("ZRANGE", setName, 0, -1))
	return n, err
}

// Rset sets the value of keyname to value.
func Rset(keyname, value string, score int) error {
	n, err := Rconn.Do("SET", keyname, value)
	if err == nil {
		fhidLogger.Loggo.Info("Wrote entry successfully", "KeyName", keyname, "Value", n)
	} else {
		fhidLogger.Loggo.Error("Error writing Redis data", "Error", Rconn.Err())
		return err
	}
	n, err = Rconn.Do("ZADD", fhidConfig.Config.RedisImageIndexSet, score, keyname)
	if err == nil {
		fhidLogger.Loggo.Debug("Successfully wrote keyname to index", "KeyName", keyname, "Index", fhidConfig.Config.RedisImageIndexSet)
	} else {
		fhidLogger.Loggo.Error("Error writing index entry", "Error", Rconn.Err())
		return err
	}
	return err
}

func getUUID() string {
	return uuid.NewV4().String()
}

// SetupConnection tests the connection to the Redis datalayer
func SetupConnection() error {
	p := pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", fhidConfig.Config.RedisEndpoint)
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
