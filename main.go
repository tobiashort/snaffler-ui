package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
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
	Path             string
	Context          string
}

//go:embed html css
var embedfs embed.FS

func main() {
	var file string
	flag.StringVar(&file, "f", "", "snaffler output file in tsv format")
	flag.StringVar(&file, "file", "", "snaffler output file in tsv format")
	flag.Parse()

	if file == "" {
		flag.Usage()
		return
	}

	content := strings.TrimSpace(string(must2(os.ReadFile(file))))
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
			Path:             atIndex(values, 11),
			Context:          atIndex(values, 12),
		}

		if logEntry.Type == "[Info]" {
			continue
		}

		if logEntry.Type == "[Share]" {
			logEntry.Path = atIndex(values, 4)
		}

    logEntry.Context = strings.ReplaceAll(logEntry.Context, "\\r", "")
    logEntry.Context = strings.ReplaceAll(logEntry.Context, "\\n", "\n")

		logEntries = append(logEntries, logEntry)
	}

	tmpl := must2(template.ParseFS(embedfs, "html/*.html"))

	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		action := req.URL.Query().Get("action")
		path := req.URL.Query().Get("path")

    switch action {
    case "open":
      powershell(fmt.Sprintf("Invoke-Item '%s'", path))
    case "download":
      powershell(fmt.Sprintf("Copy-Item -Path '%s' -Destination '$HOME\\Downloads'", path))
    }

		sorted := req.URL.Query().Get("sort") != ""
		showBlacks := req.URL.Query().Get("black") != ""
		showReds := req.URL.Query().Get("red") != ""
		showYellows := req.URL.Query().Get("yellow") != ""
		showGreens := req.URL.Query().Get("green") != ""

    allTrue := func() bool {
      return showBlacks && showReds && showYellows && showGreens
    }
    
    allFalse := func() bool {
      return !showBlacks && !showReds && !showYellows && !showGreens
    }

    var logEntriesCopy []LogEntry

    if allTrue() || allFalse() {
      logEntriesCopy = make([]LogEntry, len(logEntries))
      copy(logEntriesCopy, logEntries)
    } else {
      logEntriesCopy = make([]LogEntry, 0)
      for _, logEntry := range logEntries {
        if logEntry.Triage == "Black" && showBlacks {
          goto addLogEntry
        } else if logEntry.Triage == "Red" && showReds {
          goto addLogEntry
        } else if logEntry.Triage == "Yellow" && showYellows {
          goto addLogEntry
        } else if logEntry.Triage == "Green" && showGreens {
          goto addLogEntry
        } else {
          continue
        }
        addLogEntry:
          logEntriesCopy = append(logEntriesCopy, logEntry)
      }
    }

		if sorted {
			sort.SliceStable(logEntriesCopy, func(a, b int) bool {
				sortIndex := func(l LogEntry) int {
					switch l.Triage {
					case "Black":
						return 1
					case "Red":
						return 2
					case "Yellow":
						return 3
					case "Green":
						return 4
					default:
						return 0
					}
				}
				return sortIndex(logEntriesCopy[a]) < sortIndex(logEntriesCopy[b])
			})
		}

		indexData := struct {
			LogEntries  []LogEntry
			Sorted      bool
			ShowBlacks  bool
			ShowReds    bool
			ShowYellows bool
			ShowGreens  bool
		}{
			LogEntries:  logEntriesCopy,
			Sorted:      sorted,
			ShowBlacks:  showBlacks,
			ShowReds:    showReds,
			ShowYellows: showYellows,
			ShowGreens:  showGreens,
		}

		res.Header().Add("Content-Type", "text/html")
		tmpl.ExecuteTemplate(res, "index.html", indexData)
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


func powershell(cmd string) {
  command := exec.Command("powershell.exe", "-c", cmd)
  _, err := command.CombinedOutput()
  if err != nil {
    fmt.Fprintln(os.Stderr, "[!]", err)
  }
}
