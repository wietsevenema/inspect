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
			fmt.Fprintln(w, "ENV VARS\n========")
			for _, e := range os.Environ() {
				fmt.Fprintln(w, e)
			}
			fmt.Fprintln(w)
			fmt.Fprintln(w, "HTTP HEADERS\n============")
			for k, v := range r.Header {
				fmt.Fprintln(w, k, v)
			}
			fmt.Fprintln(w)

			fmt.Fprintln(w, " -- Source of this service: https://github.com/wietsevenema/inspect --")
		})

	log.Println("Listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
