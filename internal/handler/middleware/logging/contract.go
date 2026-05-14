package logging

import "net/http"

type pathNormalizer interface {
	Normalize(r *http.Request) string
}

type logger interface {
	Infof(format string, args ...interface{})
}
