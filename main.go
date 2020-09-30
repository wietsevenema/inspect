package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var version = "DEVELOP"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "VERSION: %s", version)
			fmt.Fprintln(w)
			fmt.Fprintln(w, r.Host)
			fmt.Fprintln(w)
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

			//FIXME: print GCP metadata info
			//FIXME: print instance stats: nr of req received
			//FIXME: print system stats (cpu mem)

			fmt.Fprintln(w, " -- Source of this service: https://github.com/wietsevenema/inspect --")
			if r.Method == http.MethodPost {
				var p interface{}
				json.NewDecoder(r.Body).Decode(&p)
				log.Print(p)
			}
		})

	log.Println("Started version: " + version)
	log.Println("Listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
