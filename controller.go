package main

import (
	"net/http"
)

type handler struct{}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/", &handler{})

	http.ListenAndServe(":8080", mux)
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost:
		Process(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
