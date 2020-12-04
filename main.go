package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"cloud.google.com/go/compute/metadata"
	sigar "github.com/cloudfoundry/gosigar"
	human "github.com/dustin/go-humanize"
)

var version = "DEVELOP"

type Data struct {
	Version        string
	Environ        map[string]string
	Headers        map[string]string
	TotalMemory    uint64
	VCPU           int
	FsList         sigar.FileSystemList
	Uptime         string
	OnGoogleCloud  bool
	ServiceAccount string
	Region         string
	InstanceID     string
	ProjectID      string
}

func CGroupCPUShares() (int, error) {
	bs, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/cpu.shares")
	if err != nil {
		return 0, err
	}
	shares, err := strconv.Atoi(
		strings.TrimSpace(string(bs)),
	)
	if err != nil {
		return 0, err
	}
	return shares / 1024, nil
}

func CGroupMemory() (uint64, error) {
	bs, err := ioutil.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	if err != nil {
		return 0, err
	}
	limit, err := strconv.ParseUint(
		strings.TrimSpace(string(bs)),
		10,
		64,
	)
	if err != nil {
		return 0, err
	}
	return limit, nil
}

func Uptime() string {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)
	uptime := sigar.Uptime{}
	uptime.Get()
	time := uint64(uptime.Length)

	days := time / (60 * 60 * 24)

	if days != 0 {
		s := ""
		if days > 1 {
			s = "s"
		}
		fmt.Fprintf(w, "%d day%s, ", days, s)
	}

	hours := time / (60 * 60)
	hours %= 24

	if hours != 0 {
		s := ""
		if hours > 1 {
			s = "s"
		}
		fmt.Fprintf(w, "%d hour%s, ", hours, s)
	}

	minutes := time / 60
	minutes %= 60

	if minutes != 0 {
		s := ""
		if minutes > 1 {
			s = "s"
		}
		fmt.Fprintf(w, "%d minute%s, ", minutes, s)
	}

	seconds := time % 60
	s := ""
	if seconds > 1 {
		s = "s"
	}
	fmt.Fprintf(w, "%d second%s", seconds, s)

	w.Flush()
	return buf.String()
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}

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

			memlimit, err := CGroupMemory()
			if err != nil {
				log.Printf("%v+", err)
				memory := sigar.Mem{}
				memory.Get()
				memlimit = memory.Total
			}
			data.TotalMemory = memlimit

			data.VCPU, _ = CGroupCPUShares()

			data.Uptime = Uptime()

			if metadata.OnGCE() {
				data.OnGoogleCloud = true
				data.ServiceAccount, _ = metadata.Email("default")
				data.InstanceID, _ = metadata.InstanceID()
				if len(data.InstanceID) > 10 {
					data.InstanceID = data.InstanceID[len(data.InstanceID)-10:]
				}

				data.Region, err = metadata.NewClient(nil).Get("instance/region")
				parts := strings.Split(data.Region, "/")
				data.Region = parts[len(parts)-1]
				data.Region = strings.TrimSpace(data.Region)

				data.ProjectID, _ = metadata.ProjectID()
			}

			err = tmpl.Execute(w, data)
			if err != nil {
				log.Printf("%v+", err)
			}

			// //FIXME: print instance stats: nr of req received

		})

	log.Println("Started version: " + version)
	log.Println("Listening on port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
