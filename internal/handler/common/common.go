package common

import (
	"encoding/json"
	"net/http"
)

func Ping(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "pong"}); err != nil {
		return
	}
}
