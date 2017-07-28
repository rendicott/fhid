package fhid

import (
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

var Rconn ResourceConn

// ResourceConn adapts a Redigo connection to a Vitess Resource.
type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

func Test() {
	p := pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", ":6379")
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
	n, err := Rconn.Do("INFO")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("info=%s", n)
	err = rset(getUuid(), "bar")
}

func rset(keyname, value string) error {
	n, err := Rconn.Do("SET", keyname, value)
	log.Printf("Wrote '%s'? '%s'", keyname, n)
	return err
}

func getUuid() string {
	return uuid.NewV4().String()
}
