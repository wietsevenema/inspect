package main

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"

	sigar "github.com/cloudfoundry/gosigar"
	human "github.com/dustin/go-humanize"
)

var version = "DEVELOP"

type Data struct {
	Version string
	Environ map[string]string
	Headers map[string]string
	Memory  sigar.Mem
	FsList  sigar.FileSystemList
	Uptime  string
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			tmpl := template.Must(template.New("index.html").
				Funcs(template.FuncMap{
					"human": func(b uint64) string {
						return human.IBytes(b)
					}}).ParseFiles("index.html"))

			data := Data{
				Version: version,
				Environ: make(map[string]string),
				Headers: make(map[string]string),
			}
			for _, e := range os.Environ() {
				r := strings.SplitN(e, "=", 2)
				if len(r) == 2 {
					v := r[1]
					if len(v) > 75 {
						v = v[:75] + "..."
					}
					data.Environ[r[0]] = v
				}
			}
			headerKeys := []string{}
			for k := range r.Header {
				headerKeys = append(headerKeys, k)
			}
			sort.Strings(headerKeys)
			for _, k := range headerKeys {
				vals := r.Header[k]
				v := strings.Join(vals, ", ")
				if len(v) > 75 {
					v = v[:75] + "..."
				}
				data.Headers[k] = v
			}

			data.Memory = sigar.Mem{}
			data.Memory.Get()

			uptime := sigar.Uptime{}
			uptime.Get()
			data.Uptime = uptime.Format()

			tmpl.Execute(w, data)

			// //FIXME: print GCP metadata info
			// //FIXME: print instance stats: nr of req received
			// //FIXME: print system stats (cpu mem)

		})

	log.Println("Started version: " + version)
	log.Println("Listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
