package routing

import (
	"net/http"
)

type parseHandler interface {
	Parse(w http.ResponseWriter, r *http.Request)
}

type topologyHandler interface {
	GetTopology(w http.ResponseWriter, r *http.Request)
}

type nodeHandler interface {
	GetNode(w http.ResponseWriter, r *http.Request)
	GetPorts(w http.ResponseWriter, r *http.Request)
}

type logHandler interface {
	GetLog(w http.ResponseWriter, r *http.Request)
}

type logger interface {
	Info(args ...interface{})
}