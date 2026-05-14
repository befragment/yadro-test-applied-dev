package routing

import (
	"net/http"
)

type nodeHandler interface {
	GetNode(w http.ResponseWriter, r *http.Request)
	GetPorts(w http.ResponseWriter, r *http.Request)
}

type logHandler interface {
	Parse(w http.ResponseWriter, r *http.Request)
	GetLog(w http.ResponseWriter, r *http.Request)
	Topology(w http.ResponseWriter, r *http.Request)
}

type logger interface {
	Infof(format string, args ...interface{})
}

type pathNormalizer interface {
	Normalize(r *http.Request) string
}
