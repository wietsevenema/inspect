package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "ENV VARS:")
			for _, e := range os.Environ() {
				fmt.Fprintln(w, e)
			}
		})

	log.Println("Listening on port: " + port)
	http.ListenAndServe(":"+port, nil)
}
