package imjasonh

import (
	"net/http"
)

const body = `<html><body>
  <h3>Welcome to app.imjasonh.com</h3>
  <p>My App Engine playground, constantly under construction since 2012</p>
  <ul>
    <li><a href="/go">Go</a>: Simple short URL redirector</li>
    <li><a href="/war">100 Games of War</a>: Simulate playing 100 games of the popular card game War&trade;</li>
  </ul>
</body></html>
`

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	})
}
