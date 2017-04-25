package main

import (
	"net/http"
)

func sendFile(p string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, p)
	}
}
