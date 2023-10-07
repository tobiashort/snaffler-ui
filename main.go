package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func must2[T any](v T, err error) T {
	must(err)
	return v
}

type LogEntry struct {
	UserHost         string
	Timestamp        string
	Type             string
	Triage           string
	Permission       string
	Rule             string
	FileSize         string
	FileLastModified string
	FilePath         string
	Context          string
}

//go:embed html css
var embedfs embed.FS

func main() {
	file := flag.String("file", "", "snaffler output file in tsv format")
	flag.Parse()

	if *file == "" {
		flag.Usage()
		return
	}

	content := strings.TrimSpace(string(must2(os.ReadFile(*file))))
	lines := strings.Split(content, "\r\n")
	logEntries := make([]LogEntry, 0)

	atIndex := func(values []string, index int) string {
		if index > len(values)-1 {
			return ""
		}
		return values[index]
	}

	for _, line := range lines {
		values := strings.Split(line, "\t")

		logEntry := LogEntry{
			UserHost:         atIndex(values, 0),
			Timestamp:        atIndex(values, 1),
			Type:             atIndex(values, 2),
			Triage:           atIndex(values, 3),
			Permission:       atIndex(values, 5),
			Rule:             atIndex(values, 8),
			FileSize:         atIndex(values, 9),
			FileLastModified: atIndex(values, 10),
			FilePath:         atIndex(values, 11),
			Context:          atIndex(values, 12),
		}

    if logEntry.Type == "[Info]" {
      continue
    }

		logEntries = append(logEntries, logEntry)
	}

  tmpl := must2(template.ParseFS(embedfs, "html/*.html"))

  http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
    res.Header().Add("Content-Type", "text/html")
    tmpl.ExecuteTemplate(res, "index.html", logEntries)
  })

  http.HandleFunc("/style.css", func(res http.ResponseWriter, req *http.Request) {
    res.Header().Add("Content-Type", "text/css")
    file := must2(embedfs.Open("css/style.css"))
    io.Copy(res, file)
  })

  addr := ":8111"
  fmt.Printf("[*] listen on %s\n", addr)
  http.ListenAndServe(addr, nil)
}
