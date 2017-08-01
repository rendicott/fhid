package fhid

import (
	"fmt"
	"net/http"

	"github.build.ge.com/212601587/fhid/fhidConfig"
)

// Handler is a basic handler
func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

// HealthCheck is a health check handler.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf(`{"State":"Healthy","Version":"%s"}`, fhidConfig.Config.Version)
	fmt.Fprintf(w, msg)
}
