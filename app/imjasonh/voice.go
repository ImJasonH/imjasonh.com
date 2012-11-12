package imjasonh

import (
	"archive/zip"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const voiceForm = `<html><body>
<form method="POST" action="/voice" enctype="multipart/form-data">
  <input type="file" name="file" /><br />
  <input type="submit" value="Submit" />
</form></body></html>
`

// voice accepts a zip file of Google Voice messages from Google Takeout
// and prints information about the calls/texts contained within.
func voice(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		fmt.Fprintf(w, voiceForm)
	} else if r.Method == "POST" {
		f, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		z, err := zip.NewReader(f, r.ContentLength)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		m := make(map[string]int64)
		for _, f := range z.File {
			full := f.Name

			if strings.HasSuffix(full, ".mp3") {
				continue
			}

			parts := strings.Split(full, "/")
			if len(parts) != 4 {
				continue
			}
			fn := parts[len(parts)-1]

			parts = strings.Split(fn, "_-_")

			m[parts[0]]++

			format := "2012-01-02T15-04-05Z.html"
			d, _ := time.Parse(format, parts[1])
			fmt.Fprintf(w, d.String())
		}
		fmt.Fprintf(w, fmt.Sprintf("Number of files: %d", len(m)))
	} else {
		http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
	}
}
