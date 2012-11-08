package imjasonh

import (
	"fmt"
	"net/http"
)

func init() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to app.imjasonh.com")
	})
	http.HandleFunc("/slurp", slurp)
}