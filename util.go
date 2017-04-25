package main

import (
	"net/http"
)

func serveFile(p string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, p)
	}
}
