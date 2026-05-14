package logging

import (
	"net/http"
	"time"
)

func LoggingMiddleware(
	logger logger,
	normalizer pathNormalizer,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rr, r)
			path := normalizer.Normalize(r)
			duration := time.Since(start).Seconds()

			logger.Infof(PrettyRequestLogFormat,
				time.Now().Format("2006/01/02 15:04:05"),
				r.Method,
				path,
				rr.status,
				duration*1000,
			)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

type Color string

const (
	ColorLightRed    Color = "\x1b[91m"
	ColorLightGreen  Color = "\x1b[92m"
	ColorLightYellow Color = "\x1b[93m"
	ColorLightBlue   Color = "\x1b[94m"
	ColorPurple      Color = "\x1b[95m"
	ColorCyan        Color = "\x1b[96m"
	ColorReset       Color = "\x1b[0m"

	PrettyRequestLogFormat string = string(ColorPurple) + "time=%s " +
		string(ColorLightBlue) + "method=%s " +
		string(ColorLightGreen) + "path=%s " +
		string(ColorLightYellow) + "status=%d " +
		string(ColorCyan) + "duration=%fms" +
		string(ColorReset)
)
