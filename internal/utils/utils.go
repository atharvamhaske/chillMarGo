package utils

import (
	"encoding/json"
	"net"
	"net/http"
)

func UserIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteResponse(w http.ResponseWriter, v any) {
	WriteJSON(w, http.StatusOK, v)
}