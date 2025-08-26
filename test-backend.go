package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Mock FDO Backend Server\n")
		fmt.Fprintf(w, "Request: %s %s\n", r.Method, r.URL.Path)
		fmt.Fprintf(w, "Headers: %v\n", r.Header)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	fmt.Println("Starting mock FDO backend on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
