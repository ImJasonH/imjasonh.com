package imjasonh

import (
	"fmt"
	"net/http"
)

const body = `<html><body>
  <h3>Welcome to app.imjasonh.com</h3>
  <p>Constantly under construction since 2012</p>
  <ul>
    <li><a href="/go">Go</a>: Simple short URL redirector</li>
    <li><a href="/slurp">Slurp</a>: Simple copy-to-Cloud-Storage utility</li>
    <li><a href="/voice">Voice</a>: Analyze your Google Voice usage</li>
    <li><a href="/mail">Mail</a>: Send email to <i>something</i>@imjasonh-hrd.appspot.com and view it online temporarily</li>
  </ul>
</body></html>
`

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, body)
	})
}
