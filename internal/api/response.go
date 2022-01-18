package api

import (
	"encoding/json"
	"net/http"
)

// response returns http response using status and data
func response(w http.ResponseWriter, status int, data interface{}) {
	dataByte, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dataByte)
}
