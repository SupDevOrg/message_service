package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Message Service...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	port := ":8080"
	log.Printf("Server started on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
