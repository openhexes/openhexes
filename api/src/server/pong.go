package server

import "net/http"

type Ponger struct{}

func (h *Ponger) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}
